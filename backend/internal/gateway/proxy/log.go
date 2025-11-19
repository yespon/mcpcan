package proxy

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/fatedier/golib/log"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
)

// LogWriter defines minimal interface for writing a gateway log into storage.
// Implementations must persist the provided record.
type LogWriter interface {
	Create(ctx context.Context, g *model.GatewayLog) error
}

// mysqlWriter is the default LogWriter that writes via MySQL repository.
type mysqlWriter struct{}

// Create persists log using the global MySQL repository.
func (w *mysqlWriter) Create(ctx context.Context, g *model.GatewayLog) error {
	return mysql.GatewayLogRepo.Create(ctx, g)
}

// queueItem wraps a gateway log and its estimated memory size for budgeting.
type queueItem struct {
	rec  *model.GatewayLog
	size int64
}

// GatewayLogQueue is a bounded in-memory queue that writes logs sequentially.
// It enforces a memory budget and drops oldest items when exceeding the budget.
type GatewayLogQueue struct {
	mu        sync.Mutex
	cond      *sync.Cond
	q         *list.List
	maxBytes  int64
	currBytes int64
	writer    LogWriter
	closed    bool
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// DefaultLogQueueBudget defines the default memory budget for the queue (200 MB).
const DefaultLogQueueBudget int64 = 200 * 1024 * 1024

// NewGatewayLogQueue constructs a new queue and starts the writer worker.
func NewGatewayLogQueue(maxBytes int64, writer LogWriter) *GatewayLogQueue {
	if maxBytes <= 0 {
		maxBytes = DefaultLogQueueBudget
	}
	if writer == nil {
		writer = &mysqlWriter{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	q := &GatewayLogQueue{
		q:        list.New(),
		maxBytes: maxBytes,
		writer:   writer,
		ctx:      ctx,
		cancel:   cancel,
	}
	q.cond = sync.NewCond(&q.mu)
	q.wg.Add(1)
	go q.worker()
	return q
}

// estimateSize approximates the memory footprint of a record.
func estimateSize(g *model.GatewayLog) int64 {
	var logLen int64
	if g.Log != nil {
		logLen = int64(len(g.Log))
	}
	// Rough estimate: sum of string lengths + JSON + small overhead.
	return int64(len(g.InstanceID)+len(g.TokenHeader)+len(g.Token)+len(g.Usages)) + logLen + 128
}

// Enqueue adds a record to the queue; drops oldest records to respect budget.
func (gq *GatewayLogQueue) Enqueue(g *model.GatewayLog) error {
	if g == nil {
		return errors.New("nil log record")
	}
	// Ensure timestamps are set if not already.
	if g.CreatedAt.IsZero() {
		now := time.Now()
		g.CreatedAt = now
		g.UpdatedAt = now
	}

	item := &queueItem{rec: g, size: estimateSize(g)}
	gq.mu.Lock()
	defer gq.mu.Unlock()
	if gq.closed {
		return errors.New("queue closed")
	}
	// Push new item.
	gq.q.PushBack(item)
	gq.currBytes += item.size
	// Trim oldest until within budget (keep at least the newest item).
	for gq.currBytes > gq.maxBytes && gq.q.Len() > 1 {
		front := gq.q.Front()
		if front != nil {
			fi := front.Value.(*queueItem)
			gq.q.Remove(front)
			gq.currBytes -= fi.size
		} else {
			break
		}
	}
	gq.cond.Signal()
	return nil
}

// Close gracefully stops the worker and flushes the queue.
func (gq *GatewayLogQueue) Close() {
	gq.mu.Lock()
	gq.closed = true
	gq.mu.Unlock()
	gq.cond.Broadcast()
	gq.cancel()
	gq.wg.Wait()
}

// worker drains the queue and writes records sequentially.
func (gq *GatewayLogQueue) worker() {
	defer gq.wg.Done()
	for {
		gq.mu.Lock()
		for gq.q.Len() == 0 && !gq.closed {
			gq.cond.Wait()
		}
		if gq.q.Len() == 0 && gq.closed {
			gq.mu.Unlock()
			return
		}
		elem := gq.q.Front()
		item := elem.Value.(*queueItem)
		gq.q.Remove(elem)
		gq.currBytes -= item.size
		gq.mu.Unlock()

		_ = gq.writer.Create(gq.ctx, item.rec)
	}
}

// GatewayLogQ is the global queue instance used by the gateway proxy.
var GatewayLogQ *GatewayLogQueue

// InitGatewayLogQueue initializes the global log queue with the default budget.
func InitGatewayLogQueue() {
	GatewayLogQ = NewGatewayLogQueue(DefaultLogQueueBudget, &mysqlWriter{})
}

// RecordGatewayLog builds a log record and enqueues it to be persisted.
func RecordGatewayLog(traceID string, instanceID string, tokenHeader string, token string, usages []string, level log.Level, event model.Event, log *model.Log) error {
	if GatewayLogQ == nil {
		InitGatewayLogQueue()
	}
	if strings.TrimSpace(instanceID) == "" {
		return nil
	}
	logRaw, _ := json.Marshal(log)

	rec := &model.GatewayLog{
		InstanceID:  instanceID,
		TokenHeader: tokenHeader,
		Token:       token,
		Usages:      strings.TrimSpace(strings.Join(usages, ",")),
		Level:       level,
		Event:       event,
		Log:         json.RawMessage(logRaw),
	}
	return GatewayLogQ.Enqueue(rec)
}

var allowedEvents = map[model.Event]struct{}{
	model.EventRequest:                 {},
	model.EventResponse:                {},
	model.EventPanicRecovered:          {},
	model.EventRequestValidationFail:   {},
	model.EventDirectorBefore:          {},
	model.EventDirectorAfter:           {},
	model.EventInstanceMissing:         {},
	model.EventSSEFlagMissing:          {},
	model.EventPathTooShort:            {},
	model.EventUpstreamURLParseFail:    {},
	model.EventProtocolUnsupported:     {},
	model.EventAccessUnsupported:       {},
	model.EventSSEStart:                {},
	model.EventSSECancel:               {},
	model.EventGzipReaderFailed:        {},
	model.EventSSEEndpointRewrite:      {},
	model.EventSSEEof:                  {},
	model.EventSSEReadError:            {},
	model.EventClientCanceled:          {},
	model.EventProxyErrorLog:           {},
	model.EventUpstreamConnInterrupted: {},
	model.EventUpstreamError:           {},
}

func isAllowedEvent(e model.Event) bool {
	_, ok := allowedEvents[e]
	return ok
}

func WriteMCPLog(traceID, instanceID string, tokenHeader string, token string, level log.Level, event model.Event, usages []string, msg string) {
	if strings.TrimSpace(instanceID) == "" {
		return
	}
	if !isAllowedEvent(event) {
		return
	}
	log := &model.Log{
		Event:   event,
		Level:   level,
		Message: msg,
		TS:      time.Now().Format(time.RFC3339Nano),
	}
	_ = RecordGatewayLog(traceID, instanceID, tokenHeader, strings.TrimSpace(token), usages, level, event, log)
}

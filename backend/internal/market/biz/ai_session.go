package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/llm"
	_ "github.com/kymo-mcp/mcpcan/pkg/llm/openai"
)

type AiSessionBiz struct {
	ctx context.Context
}

var GAiSessionBiz *AiSessionBiz

func init() {
	GAiSessionBiz = NewAiSessionBiz(context.Background())
}

func NewAiSessionBiz(ctx context.Context) *AiSessionBiz {
	return &AiSessionBiz{
		ctx: ctx,
	}
}

func (b *AiSessionBiz) Create(ctx context.Context, req *pb.CreateSessionRequest, userID int64) (*model.AiSession, error) {
	// Validate ModelAccessID exists
	if _, err := mysql.AiModelAccessRepo.FindByID(ctx, req.ModelAccessID); err != nil {
		return nil, fmt.Errorf("invalid model access id")
	}

	session := &model.AiSession{
		UserID:        userID,
		Name:          req.Name,
		ModelAccessID: req.ModelAccessID,
		MaxContext:    int(req.MaxContext),
		ToolsConfig:   json.RawMessage(req.ToolsConfig),
	}

	if err := mysql.AiSessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (b *AiSessionBiz) Update(ctx context.Context, req *pb.UpdateSessionRequest) (*model.AiSession, error) {
	session, err := mysql.AiSessionRepo.FindByID(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	if req.Name != "" {
		session.Name = req.Name
	}
	if req.ModelAccessID != 0 {
		if _, err := mysql.AiModelAccessRepo.FindByID(ctx, req.ModelAccessID); err != nil {
			return nil, fmt.Errorf("invalid model access id")
		}
		session.ModelAccessID = req.ModelAccessID
	}
	if req.ToolsConfig != "" {
		session.ToolsConfig = json.RawMessage(req.ToolsConfig)
	}
	if req.MaxContext != 0 {
		session.MaxContext = int(req.MaxContext)
	}

	if err := mysql.AiSessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (b *AiSessionBiz) Delete(ctx context.Context, id int64) error {
	return mysql.AiSessionRepo.Delete(ctx, id)
}

func (b *AiSessionBiz) Get(ctx context.Context, id int64) (*model.AiSession, error) {
	return mysql.AiSessionRepo.FindByID(ctx, id)
}

func (b *AiSessionBiz) List(ctx context.Context, userID int64, page, pageSize int) ([]*model.AiSession, int64, error) {
	sessions, err := mysql.AiSessionRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(sessions))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > int(total) {
		start = int(total)
	}
	if end > int(total) {
		end = int(total)
	}

	return sessions[start:end], total, nil
}

func (b *AiSessionBiz) GetMessages(ctx context.Context, sessionID int64, limit int) ([]*model.AiMessage, error) {
	if limit <= 0 {
		limit = 20
	}
	return mysql.AiMessageRepo.FindBySessionID(ctx, sessionID, limit)
}

// Chat prepares the chat stream and saves the user message
func (b *AiSessionBiz) Chat(ctx context.Context, sessionID int64, content string, toolsConfig string) (<-chan llm.StreamResponse, error) {
	// 1. Load Session
	session, err := mysql.AiSessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	// 2. Load Model Access
	modelAccess, err := mysql.AiModelAccessRepo.FindByID(ctx, session.ModelAccessID)
	if err != nil {
		return nil, fmt.Errorf("model access config not found")
	}

	// 3. Load History
	limit := session.MaxContext
	if limit <= 0 {
		limit = 20
	}
	historyMessages, err := mysql.AiMessageRepo.GetLastN(ctx, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to load history")
	}

	// 4. Init LLM Provider
	providerType := llm.ProviderOpenAI
	if modelAccess.Provider != "" {
		providerType = llm.ProviderType(modelAccess.Provider)
	}

	provider, err := llm.NewProvider(providerType, llm.ProviderConfig{
		BaseURL: modelAccess.BaseUrl,
		APIKey:  modelAccess.ApiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init provider: %s", err.Error())
	}

	// 5. Construct Messages
	var messages []llm.Message
	for _, msg := range historyMessages {
		m := llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
		if msg.ToolCalls != "" && msg.ToolCalls != "[]" && msg.ToolCalls != "null" {
			var toolCalls []llm.ToolCall
			if err := json.Unmarshal([]byte(msg.ToolCalls), &toolCalls); err == nil {
				m.ToolCalls = toolCalls
			}
		}
		if msg.ToolCallID != "" {
			m.ToolCallID = msg.ToolCallID
		}
		messages = append(messages, m)
	}

	// Add Current User Message
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: content,
	})

	// 6. Save User Message
	userMsg := &model.AiMessage{
		SessionID:  sessionID,
		Role:       "user",
		Content:    content,
		CreateTime: time.Now(),
	}
	// Estimate tokens? For now 0
	if err := mysql.AiMessageRepo.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %s", err.Error())
	}

	// 7. Init MCP Tools
	var tools []llm.Tool
	mcpManager := NewMcpManager()

	// Use provided toolsConfig if available, otherwise use session's
	configToUse := string(session.ToolsConfig)
	if toolsConfig != "" {
		configToUse = toolsConfig
	}

	if len(configToUse) > 0 && configToUse != "{}" && configToUse != "null" {
		if err := mcpManager.Initialize(ctx, configToUse); err != nil {
			mcpManager.Close()
			return nil, fmt.Errorf("failed to init mcp tools: %v", err)
		}
		var err error
		tools, err = mcpManager.GetTools(ctx)
		if err != nil {
			mcpManager.Close()
			return nil, fmt.Errorf("failed to get tools: %v", err)
		}
	}

	// 8. Create Stream Loop
	outCh := make(chan llm.StreamResponse)

	go func() {
		defer close(outCh)
		defer mcpManager.Close()

		currentMessages := messages
		maxTurns := 10 // Safety limit for tool loops

		for turn := 0; turn < maxTurns; turn++ {
			req := llm.ChatRequest{
				Model:    modelAccess.ModelName,
				Messages: currentMessages,
				Stream:   true,
				Tools:    tools,
			}

			stream, err := provider.StreamChat(ctx, req)
			if err != nil {
				outCh <- llm.StreamResponse{Error: err}
				return
			}

			var accumulatedContent string
			accumulatedToolCalls := make(map[int]*llm.ToolCall)

			for resp := range stream {
				if resp.Error != nil {
					outCh <- resp
					return
				}

				if resp.Usage != nil {
					outCh <- llm.StreamResponse{Usage: resp.Usage}
				}

				if resp.Content != "" {
					accumulatedContent += resp.Content
					outCh <- llm.StreamResponse{Content: resp.Content}
				}

				if len(resp.ToolCalls) > 0 {
					for _, tc := range resp.ToolCalls {
						if _, exists := accumulatedToolCalls[tc.Index]; !exists {
							accumulatedToolCalls[tc.Index] = &llm.ToolCall{
								Index: tc.Index,
								ID:    tc.ID,
								Type:  tc.Type,
								Function: llm.ToolCallFunction{
									Name: tc.Function.Name,
								},
							}
						}
						accumulatedToolCalls[tc.Index].Function.Arguments += tc.Function.Arguments
						if tc.ID != "" {
							accumulatedToolCalls[tc.Index].ID = tc.ID
						}
						if tc.Function.Name != "" {
							accumulatedToolCalls[tc.Index].Function.Name = tc.Function.Name
						}
						if tc.Type != "" {
							accumulatedToolCalls[tc.Index].Type = tc.Type
						}
					}
				}
			}

			// Turn finished. Check for tool calls.
			if len(accumulatedToolCalls) == 0 {
				// No tools called, save assistant message and exit
				asstMsg := &model.AiMessage{
					SessionID:  sessionID,
					Role:       "assistant",
					Content:    accumulatedContent,
					CreateTime: time.Now(),
				}
				mysql.AiMessageRepo.Create(ctx, asstMsg)
				return
			}

			// Prepare Tool Calls
			var toolCalls []llm.ToolCall
			maxIndex := -1
			for idx := range accumulatedToolCalls {
				if idx > maxIndex {
					maxIndex = idx
				}
			}
			for i := 0; i <= maxIndex; i++ {
				if tc, ok := accumulatedToolCalls[i]; ok {
					toolCalls = append(toolCalls, *tc)
				}
			}

			// Append Assistant Message
			currentMessages = append(currentMessages, llm.Message{
				Role:      "assistant",
				Content:   accumulatedContent,
				ToolCalls: toolCalls,
			})

			toolCallsJSON, _ := json.Marshal(toolCalls)
			dbAsstMsg := &model.AiMessage{
				SessionID:  sessionID,
				Role:       "assistant",
				Content:    accumulatedContent,
				ToolCalls:  string(toolCallsJSON),
				CreateTime: time.Now(),
			}
			mysql.AiMessageRepo.Create(ctx, dbAsstMsg)

			// Execute Tools
			for _, tc := range toolCalls {
				resultStr := ""
				mcpResult, err := mcpManager.CallTool(ctx, tc.Function.Name, tc.Function.Arguments)
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					var texts []string
					for _, c := range mcpResult.Content {
						// Marshal content to JSON string to avoid field access issues
						if b, err := json.Marshal(c); err == nil {
							texts = append(texts, string(b))
						} else {
							texts = append(texts, fmt.Sprintf("%v", c))
						}
					}
					resultStr = strings.Join(texts, "\n")
					if mcpResult.IsError {
						resultStr = "Tool Error: " + resultStr
					}
				}

				// Append Tool Message
				currentMessages = append(currentMessages, llm.Message{
					Role:       "tool",
					Content:    resultStr,
					ToolCallID: tc.ID,
				})

				dbToolMsg := &model.AiMessage{
					SessionID:  sessionID,
					Role:       "tool",
					Content:    resultStr,
					ToolCallID: tc.ID,
					CreateTime: time.Now(),
				}
				mysql.AiMessageRepo.Create(ctx, dbToolMsg)

				// Notify Client
				outCh <- llm.StreamResponse{
					ToolOutputs: []llm.ToolOutput{
						{
							ID:     tc.ID,
							Name:   tc.Function.Name,
							Result: resultStr,
						},
					},
				}
			}
			// Continue loop for next turn
		}
	}()

	return outCh, nil
}

func (b *AiSessionBiz) SaveMessage(ctx context.Context, msg *model.AiMessage) error {
	msg.CreateTime = time.Now()
	return mysql.AiMessageRepo.Create(ctx, msg)
}

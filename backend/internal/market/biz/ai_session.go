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
		ModelName:     req.ModelName, // 模型名称在会话中指定
		Temperature:   float64(req.Temperature),
		SystemPrompt:  req.SystemPrompt,
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
	if req.ModelName != "" {
		session.ModelName = req.ModelName
	}
	// Temperature handling (since float default is 0, we need careful check or allow 0)
	// Proto3 zero value is 0. But 0 temperature is valid. 
	// However, usually update requests carry what changed.
	// For simplicity in this generated proto, we assume if it's there we update it.
	// But actually, proto3 doesn't distinguish between unset and 0.
	// We might need a wrapper or assume front-end sends current value.
	session.Temperature = float64(req.Temperature)
	
	if req.SystemPrompt != "" {
		session.SystemPrompt = req.SystemPrompt
	}

	if err := mysql.AiSessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (b *AiSessionBiz) Delete(ctx context.Context, id int64) error {
	// 1. Get all messages for the session to find files
	// Note: using a large limit to get all messages, or better, use FindBySessionID without limit if available
	messages, err := mysql.AiMessageRepo.FindBySessionID(ctx, id, 10000)
	if err != nil {
		// Log error but proceed with deletion? Or fail? 
		// For now, fail safe
		return fmt.Errorf("failed to get messages for cleanup: %v", err)
	}

	// 2. Identify files to delete
	var fileURLs []string
	for _, msg := range messages {
		// Simple parsing for ![image](/files/...)
		// Optimally use regex or more robust parsing
		if strings.Contains(msg.Content, "![image](/files/") {
			lines := strings.Split(msg.Content, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "![image](") && strings.HasSuffix(line, ")") {
					url := strings.TrimSuffix(strings.TrimPrefix(line, "![image]("), ")")
					fileURLs = append(fileURLs, url)
				}
			}
		}
	}

	// 3. Delete files from storage
	if len(fileURLs) > 0 {
		if err := GAiFileManager.DeleteFiles(fileURLs); err != nil {
			// Log error but continue with DB deletion
			fmt.Printf("Failed to delete files for session %d: %v\n", id, err)
		}
	}

	// 4. Delete session from DB
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

func (b *AiSessionBiz) GetMessages(ctx context.Context, sessionID int64, page, pageSize int) ([]*model.AiMessage, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return mysql.AiMessageRepo.FindBySessionIDPaged(ctx, sessionID, page, pageSize)
}

// Chat prepares the chat stream and saves the user message
func (b *AiSessionBiz) Chat(ctx context.Context, req *pb.ChatRequest) (<-chan llm.StreamResponse, error) {
	sessionID := req.SessionID
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

	// Inject Session's SystemPrompt if configured
	if session.SystemPrompt != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: session.SystemPrompt,
		})
	}

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
	userMessage := llm.Message{
		Role:    "user",
		Content: req.Content,
	}

	// Handle Attachments (Multimodal)
	if len(req.Attachments) > 0 {
		var parts []llm.MessageContentPart
		
		// Text part first
		if req.Content != "" {
			parts = append(parts, llm.MessageContentPart{
				Type: "text", 
				Text: req.Content,
			})
		}

		// Image parts
		for _, attach := range req.Attachments {
			// Only support images for now, can extend to files if model supports
			if strings.HasPrefix(attach.Type, "image") || attach.Type == "image" {
				parts = append(parts, llm.MessageContentPart{
					Type: "image_url",
					ImageURL: &llm.MessageImageURL{
						URL: attach.Url,
					},
				})
			}
		}
		userMessage.MultiContent = parts
	}
	messages = append(messages, userMessage)

	// 6. Save User Message
	// Append image markdown to content so it can be retrieved for deletion (and history)
	// Format: ![image](/files/xxx.png)
	dbContent := req.Content
	for _, attach := range req.Attachments {
		if strings.HasPrefix(attach.Type, "image") {
			dbContent += fmt.Sprintf("\n![image](%s)", attach.Url)
		}
	}

	userMsg := &model.AiMessage{
		SessionID:  sessionID,
		Role:       "user",
		Content:    dbContent, 
		CreateTime: time.Now(),
	}
	// Estimate tokens (improved estimation welcome)
	estimatedTokens := len(req.Content) / 4
	if estimatedTokens < 1 { estimatedTokens = 1 }
	userMsg.PromptTokens = estimatedTokens
	userMsg.TotalTokens = estimatedTokens
	
	if err := mysql.AiMessageRepo.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %s", err.Error())
	}

	// 7. Init MCP Tools & Debug Logic
	var tools []llm.Tool
	mcpManager := NewMcpManager()
	
	// Determine config to use
	var configToUse string
	var requestedTools []string
	var enableAllTools bool
	var useProfile bool

	if req.McpProfile != nil && req.McpProfile.InstanceId != "" {
		useProfile = true
		// Load instance config
		instance, err := GInstanceBiz.GetInstance(req.McpProfile.InstanceId)
		if err != nil {
			mcpManager.Close()
			return nil, fmt.Errorf("failed to load mcp instance: %v", err)
		}
		if instance != nil {
			configToUse = string(instance.SourceConfig)
			// Parse tool filters
			requestedTools = req.McpProfile.IncludeTools
			enableAllTools = req.McpProfile.EnableAll
		}
	} else {
		// Fallback to session/request overrides
		configToUse = string(session.ToolsConfig)
		if req.Tools != "" {
			configToUse = req.Tools
		}
		enableAllTools = true // Default to all if not using profile (legacy behavior)
	}

	systemInstruction := ""

	if len(configToUse) > 0 && configToUse != "{}" && configToUse != "null" {
		if err := mcpManager.Initialize(ctx, configToUse); err != nil {
			mcpManager.Close()
			return nil, fmt.Errorf("failed to init mcp tools: %v", err)
		}
		
		allTools, err := mcpManager.GetTools(ctx)
		if err != nil {
			mcpManager.Close()
			return nil, fmt.Errorf("failed to get tools: %v", err)
		}

		if useProfile && !enableAllTools && len(requestedTools) > 0 {
			// Filter tools
			toolMap := make(map[string]llm.Tool)
			for _, t := range allTools {
				toolMap[t.Function.Name] = t
			}

			var filteredTools []llm.Tool
			var missingTools []string

			for _, name := range requestedTools {
				if t, ok := toolMap[name]; ok {
					filteredTools = append(filteredTools, t)
				} else {
					missingTools = append(missingTools, name)
				}
			}
			tools = filteredTools

			// Inject warning for missing tools
			if len(missingTools) > 0 {
				systemInstruction = fmt.Sprintf("\n[System Warning]: The following requested tools are unavailable or not found: %s. You cannot use them.", strings.Join(missingTools, ", "))
			}
		} else {
			tools = allTools
		}
	}

	// Inject system instruction if present
	if systemInstruction != "" {
		// Prepend system message or append to last user message?
		// Usually system message at start is best.
		sysMsg := llm.Message{
			Role:    "system",
			Content: systemInstruction,
		}
		// Insert at beginning of context (messages 0 is usually system if present, otherwise prepend)
		// Simple prepend here
		messages = append([]llm.Message{sysMsg}, messages...)
	}

	// 8. Create Stream Loop
	outCh := make(chan llm.StreamResponse)

	go func() {
		defer close(outCh)
		defer mcpManager.Close()

		currentMessages := messages
		maxTurns := 10 // Safety limit for tool loops

		for turn := 0; turn < maxTurns; turn++ {
			reqChat := llm.ChatRequest{
				Model:       session.ModelName,
				Messages:    currentMessages,
				Stream:      true,
				Tools:       tools,
				Temperature: float32(session.Temperature),
			}

			stream, err := provider.StreamChat(ctx, reqChat)
			if err != nil {
				outCh <- llm.StreamResponse{Error: err}
				return
			}

			var accumulatedContent string
			accumulatedToolCalls := make(map[int]*llm.ToolCall)
			var finalUsage *llm.Usage // 累积 Token 统计信息

			for resp := range stream {
				if resp.Error != nil {
					outCh <- resp
					return
				}

				if resp.Usage != nil {
					finalUsage = resp.Usage // 记录最后的 Usage
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
				// 添加 Token 统计
				if finalUsage != nil {
					asstMsg.PromptTokens = finalUsage.PromptTokens
					asstMsg.CompletionTokens = finalUsage.CompletionTokens
					asstMsg.TotalTokens = finalUsage.TotalTokens
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
			// 添加 Token 统计
			if finalUsage != nil {
				dbAsstMsg.PromptTokens = finalUsage.PromptTokens
				dbAsstMsg.CompletionTokens = finalUsage.CompletionTokens
				dbAsstMsg.TotalTokens = finalUsage.TotalTokens
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

// SessionUsage 会话的 Token 使用统计
type SessionUsage struct {
	SessionID        int64 `json:"sessionId"`
	TotalMessages    int   `json:"totalMessages"`
	PromptTokens     int   `json:"promptTokens"`
	CompletionTokens int   `json:"completionTokens"`
	TotalTokens      int   `json:"totalTokens"`
}

// GetSessionUsage 获取会话的 Token 使用统计
func (b *AiSessionBiz) GetSessionUsage(ctx context.Context, sessionID int64) (*SessionUsage, error) {
	// 获取会话的所有消息
	messages, err := mysql.AiMessageRepo.FindBySessionID(ctx, sessionID, 10000) // 使用大数字获取所有消息
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	usage := &SessionUsage{
		SessionID: sessionID,
	}

	// 累计统计
	for _, msg := range messages {
		usage.TotalMessages++
		usage.PromptTokens += msg.PromptTokens
		usage.CompletionTokens += msg.CompletionTokens
		usage.TotalTokens += msg.TotalTokens
	}

	return usage, nil
}

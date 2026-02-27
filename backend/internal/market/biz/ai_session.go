package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	llm "github.com/kymo-mcp/mcpcan/pkg/llm_adapter"
	"github.com/mark3labs/mcp-go/mcp"
)

// normalizeToolsConfig 标准化 MCP 配置 JSON
// 兼容以下输入：空字符串、合法 JSON 对象、被 double-encode 的 JSON 字符串
func normalizeToolsConfig(raw string) json.RawMessage {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return json.RawMessage("{}")
	}

	// 尝试直接解析为 JSON
	if json.Valid([]byte(raw)) {
		return json.RawMessage(raw)
	}

	// 可能是被 double-encode 的字符串（如 "\"{ ... }\"" ）
	var unquoted string
	if err := json.Unmarshal([]byte(raw), &unquoted); err == nil {
		if json.Valid([]byte(unquoted)) {
			return json.RawMessage(unquoted)
		}
	}

	// 无法解析，返回空对象
	log.Printf("[normalizeToolsConfig] invalid JSON, fallback to {}: %s", raw)
	return json.RawMessage("{}")
}

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
		log.Printf("Invalid model access id %d provided when creating session: %v", req.ModelAccessID, err)
		return nil, fmt.Errorf("invalid model access id %d: %w", req.ModelAccessID, err)
	}

	// 校验并标准化 ToolsConfig JSON
	toolsConfig := normalizeToolsConfig(req.ToolsConfig)

	session := &model.AiSession{
		UserID:        userID,
		Name:          req.Name,
		ModelAccessID: req.ModelAccessID,
		ModelName:     req.ModelName, // 模型名称在会话中指定
		Temperature:   float64(req.Temperature),
		SystemPrompt:  req.SystemPrompt,
		MaxContext:    int(req.MaxContext),
		ToolsConfig:   toolsConfig,
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
			log.Printf("Invalid model access id %d provided when updating session %d: %v", req.ModelAccessID, req.Id, err)
			return nil, fmt.Errorf("invalid model access id %d: %w", req.ModelAccessID, err)
		}
		session.ModelAccessID = req.ModelAccessID
	}
	if req.ToolsConfig != "" {
		session.ToolsConfig = normalizeToolsConfig(req.ToolsConfig)
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
	// Note: using a large limit to get all messages, or better, use FindBySessionId without limit if available
	messages, err := mysql.AiMessageRepo.FindBySessionId(ctx, id, 10000)
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
	return mysql.AiMessageRepo.FindBySessionIdPaged(ctx, sessionID, page, pageSize)
}

// Chat prepares the chat stream and saves the user message
func (b *AiSessionBiz) Chat(ctx context.Context, req *pb.ChatRequest) (<-chan llm.StreamResponse, error) {
	sessionID := req.SessionId
	// 1. Load Session
	session, err := mysql.AiSessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found")
	}

	// 2. Load Model Access
	modelAccess, err := mysql.AiModelAccessRepo.FindByID(ctx, session.ModelAccessID)
	if err != nil {
		// Log the specific ID that was not found for debugging
		log.Printf("Model access config not found for ID: %d, Session ID: %d", session.ModelAccessID, sessionID)
		return nil, fmt.Errorf("model access config not found for ID %d: %w", session.ModelAccessID, err)
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
		providerType = llm.ProviderType(strings.ToLower(modelAccess.Provider))
	}

	provider, err := llm.NewProvider(providerType, llm.ProviderConfig{
		BaseURL:  modelAccess.BaseUrl,
		APIKey:   modelAccess.ApiKey,
		ProxyURL: llm.GlobalProxyURL,
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

	// 构建 ToolCallID -> FunctionName 映射（用于历史 tool 消息恢复函数名）
	toolCallNameMap := make(map[string]string)
	for _, msg := range historyMessages {
		if msg.ToolCalls != "" && msg.ToolCalls != "[]" && msg.ToolCalls != "null" {
			var toolCalls []llm.ToolCall
			if err := json.Unmarshal([]byte(msg.ToolCalls), &toolCalls); err == nil {
				for _, tc := range toolCalls {
					if tc.ID != "" && tc.Function.Name != "" {
						toolCallNameMap[tc.ID] = tc.Function.Name
					}
				}
			}
		}
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
			// 从映射中恢复函数名（Google FunctionResponse 需要）
			if name, ok := toolCallNameMap[msg.ToolCallID]; ok {
				m.ToolCallName = name
			}
		}
		// 只透传数据库中实际存储的 reasoning_content
		// omitempty 保证空值不序列化，避免不支持的模型（如 Mistral）报 422
		m.ReasoningContent = msg.ReasoningContent
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

	// 清洗历史消息：去掉开头孤立的 tool 消息、合并连续同 role 消息
	messages = sanitizeMessages(messages)

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
		SessionId:  sessionID,
		Role:       "user",
		Content:    dbContent,
		CreateTime: time.Now(),
	}
	// Estimate tokens (improved estimation welcome)
	estimatedTokens := len(req.Content) / 4
	if estimatedTokens < 1 {
		estimatedTokens = 1
	}
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

	// Debug: log the config being used for MCP
	log.Printf("[MCP Debug] Session %d, configToUse length=%d, value=%s", sessionID, len(configToUse), configToUse)

	if len(configToUse) > 0 && configToUse != "{}" && configToUse != "null" {
		log.Printf("[MCP Debug] Initializing MCP tools...")
		if err := mcpManager.Initialize(ctx, configToUse); err != nil {
			log.Printf("[MCP Error] Failed to init: %v", err)
			mcpManager.Close()
			return nil, fmt.Errorf("failed to init mcp tools: %v", err)
		}

		allTools, err := mcpManager.GetTools(ctx)
		if err != nil {
			mcpManager.Close()
			return nil, fmt.Errorf("failed to get tools: %v", err)
		}

		// Robust Check: If toolsConfig was provided but no tools found, return error
		if len(allTools) == 0 {
			mcpManager.Close()
			return nil, fmt.Errorf("mcp initialization successful but no tools found. please check your mcp status or config")
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
			// Kimi (Moonshot) 模型有温度限制，通常只支持 1.0 (或默认值)
			// 用户反馈报错: [System Error: API returned unexpected status code: 400: invalid temperature: only 1 is allowed for this model]
			temperature := float32(session.Temperature)
			if providerType == llm.ProviderMoonshot {
				temperature = 1.0
			}

			reqChat := llm.ChatRequest{
				Model:       session.ModelName,
				Messages:    currentMessages,
				Stream:      true,
				Tools:       tools,
				Temperature: temperature,
			}

			fmt.Printf("[AiSessionBiz] Start StreamChat. ProviderType: %s, Model: %s\n", providerType, session.ModelName)

			stream, err := provider.StreamChat(ctx, reqChat)
			if err != nil {
				fmt.Printf("[AiSessionBiz] StreamChat failed: %v\n", err)
				outCh <- llm.StreamResponse{Error: err}
				return
			}

			var accumulatedContent string
			var accumulatedReasoning string
			// accumulatedToolCalls 键是 index, 值是 *ToolCall
			accumulatedToolCalls := make(map[int]*llm.ToolCall)
			var finalUsage *llm.Usage
			
			// Stream Loop
			for resp := range stream {
				if resp.Error != nil {
					// If we have accumulated content, save it before returning error
					if accumulatedContent != "" || accumulatedReasoning != "" {
						asstMsg := &model.AiMessage{
							SessionId:        sessionID,
							Role:             "assistant",
							Content:          accumulatedContent,
							ReasoningContent: accumulatedReasoning,
							CreateTime:       time.Now(),
						}
						if finalUsage != nil {
							asstMsg.PromptTokens = finalUsage.PromptTokens
							asstMsg.CompletionTokens = finalUsage.CompletionTokens
							asstMsg.TotalTokens = finalUsage.TotalTokens
							asstMsg.PromptTokens = finalUsage.PromptTokens
							asstMsg.CompletionTokens = finalUsage.CompletionTokens
							asstMsg.TotalTokens = finalUsage.TotalTokens
						}
						mysql.AiMessageRepo.Create(ctx, asstMsg)
					}
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
				
				if resp.ReasoningContent != "" {
					accumulatedReasoning += resp.ReasoningContent
					outCh <- llm.StreamResponse{ReasoningContent: resp.ReasoningContent}
				}

				if len(resp.ToolCalls) > 0 {
					// Forward tool calls to frontend
					outCh <- llm.StreamResponse{
						ToolCalls: resp.ToolCalls,
					}

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
					SessionId:        sessionID,
					Role:             "assistant",
					Content:          accumulatedContent,
					ReasoningContent: accumulatedReasoning,
					CreateTime:       time.Now(),
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

			// Append Assistant message to context
			currentMessages = append(currentMessages, llm.Message{
				Role:             "assistant",
				Content:          accumulatedContent,
				ReasoningContent: accumulatedReasoning,
				ToolCalls:        toolCalls,
			})

			toolCallsJson, _ := json.Marshal(toolCalls)
			// Save Assistant Message (with Tool Calls)
			dbAsstMsg := &model.AiMessage{
				SessionId:        sessionID,
				Role:             "assistant",
				Content:          accumulatedContent,
				ReasoningContent: accumulatedReasoning,
				ToolCalls:        string(toolCallsJson),
				CreateTime:       time.Now(),
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
						switch v := c.(type) {
						case mcp.TextContent:
							texts = append(texts, v.Text)
						case mcp.ImageContent:
							texts = append(texts, fmt.Sprintf("[Image: %s, Data: %s]", v.MIMEType, v.Data))
						case *mcp.TextContent:
							texts = append(texts, v.Text)
						case *mcp.ImageContent:
							texts = append(texts, fmt.Sprintf("[Image: %s, Data: %s]", v.MIMEType, v.Data))
						default:
							// For other types (e.g. EmbeddedResource), usage default JSON marshaling
							if b, err := json.Marshal(c); err == nil {
								texts = append(texts, string(b))
							} else {
								texts = append(texts, fmt.Sprintf("%v", c))
							}
						}
					}
					resultStr = strings.Join(texts, "\n")
					if mcpResult.IsError {
						resultStr = "Tool Error: " + resultStr
					}
				}

				// Append Tool Message
				currentMessages = append(currentMessages, llm.Message{
					Role:         "tool",
					Content:      resultStr,
					ToolCallID:   tc.ID,
					ToolCallName: tc.Function.Name,
				})

				dbToolMsg := &model.AiMessage{
					SessionId:  sessionID,
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
	SessionId        int64 `json:"sessionId"`
	TotalMessages    int   `json:"totalMessages"`
	PromptTokens     int   `json:"promptTokens"`
	CompletionTokens int   `json:"completionTokens"`
	TotalTokens      int   `json:"totalTokens"`
}

// GetSessionUsage 获取会话的 Token 使用统计
func (b *AiSessionBiz) GetSessionUsage(ctx context.Context, sessionID int64) (*SessionUsage, error) {
	// 获取会话的所有消息
	messages, err := mysql.AiMessageRepo.FindBySessionId(ctx, sessionID, 10000) // 使用大数字获取所有消息
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	usage := &SessionUsage{
		SessionId: sessionID,
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

// sanitizeMessages 清洗消息列表，修复截断和格式问题
// 1. 去掉开头孤立的 tool 消息（没有前置 assistant+ToolCalls）
// 2. 去掉孤立的 assistant+ToolCalls（没有后续 tool response）
// 3. 合并连续同 role 消息（Google API 要求 user/model 交替）
func sanitizeMessages(messages []llm.Message) []llm.Message {
	if len(messages) == 0 {
		return messages
	}

	// Step 1: 收集所有有效的 tool call ID（来自 assistant 消息的 ToolCalls）
	// 和所有 tool response 的 ToolCallID，以及 ToolCallID -> Name 映射
	assistantToolCallIDs := make(map[string]bool)
	toolResponseIDs := make(map[string]bool)
	toolCallNames := make(map[string]string) // ID -> Name

	for _, m := range messages {
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				assistantToolCallIDs[tc.ID] = true
				toolCallNames[tc.ID] = tc.Function.Name
			}
		}
		if m.Role == "tool" && m.ToolCallID != "" {
			toolResponseIDs[m.ToolCallID] = true
		}
	}

	// Step 2: 过滤掉孤立消息并恢复 ToolCallName
	var cleaned []llm.Message
	for _, m := range messages {
		switch m.Role {
		case "tool":
			// 只保留有对应 assistant+ToolCalls 的 tool response
			if m.ToolCallID != "" && assistantToolCallIDs[m.ToolCallID] {
				// 关键修复：从 assistant 的 ToolCall 中恢复 Name
				// 因为数据库可能没存 ToolCallName，导致 Google API 报错
				if name, ok := toolCallNames[m.ToolCallID]; ok && m.ToolCallName == "" {
					m.ToolCallName = name
				}
				cleaned = append(cleaned, m)
			} else {
				log.Printf("[sanitizeMessages] dropping orphaned tool message: ToolCallID=%s", m.ToolCallID)
			}
		case "assistant":
			if len(m.ToolCalls) > 0 {
				// 检查是否所有 tool call 都有对应的 response
				hasAllResponses := true
				for _, tc := range m.ToolCalls {
					if !toolResponseIDs[tc.ID] {
						hasAllResponses = false
						break
					}
				}
				if hasAllResponses {
					cleaned = append(cleaned, m)
				} else {
					// 去掉 ToolCalls 保留纯文本内容
					if strings.TrimSpace(m.Content) != "" {
						cleaned = append(cleaned, llm.Message{
							Role:    m.Role,
							Content: m.Content,
						})
					} else {
						log.Printf("[sanitizeMessages] dropping assistant+ToolCalls without all responses")
					}
				}
			} else {
				// 只有当 Content 非空时才保留，否则会导致 API 400
				if strings.TrimSpace(m.Content) != "" {
					cleaned = append(cleaned, m)
				} else {
					log.Printf("[sanitizeMessages] dropping empty assistant message")
				}
			}
		default:
			cleaned = append(cleaned, m)
		}
	}

	// Step 3: 合并连续同 role 消息（Google/GLM 要求 user/model 交替）
	var merged []llm.Message
	for _, m := range cleaned {
		if len(merged) > 0 {
			last := &merged[len(merged)-1]
			// 合并连续的同角色的消息
			if last.Role == m.Role {
				// Tool/Function 消息不合并，因为它们有各自的 ID
				if m.Role == "user" || m.Role == "system" || m.Role == "assistant" {
					// 如果 assistant 这一条或上一条含有 ToolCalls，则不合并（保持结构完整性）
					if m.Role == "assistant" && (len(m.ToolCalls) > 0 || len(last.ToolCalls) > 0) {
						merged = append(merged, m)
						continue
					}
					last.Content += "\n" + m.Content
					continue
				}
			}
		}
		merged = append(merged, m)
	}

	// Step 4: 最终校验与修复 (Final Validation)
	// 确保开头是 User/System，且严格交替
	var final []llm.Message
	if len(merged) > 0 {
		// 1. 确保第一条非 System 消息是 User
		firstNonSystemIdx := 0
		for i, m := range merged {
			if m.Role != "system" {
				firstNonSystemIdx = i
				break
			}
		}

		// 如果第一条有效消息是 Assistant，补一个 User
		if firstNonSystemIdx < len(merged) && merged[firstNonSystemIdx].Role == "assistant" {
			// 在 System 消息之后插入 User
			prefix := merged[:firstNonSystemIdx]
			suffix := merged[firstNonSystemIdx:]
			final = append(final, prefix...)
			final = append(final, llm.Message{
				Role:    "user",
				Content: "...", // Placeholder to satisfy API requirements
			})
			final = append(final, suffix...)
		} else {
			final = merged
		}
	} else {
		final = merged
	}

	return final
}

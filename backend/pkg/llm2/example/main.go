package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kymo-mcp/mcpcan/pkg/llm2"
)

func main() {
	fmt.Println("LLM2 包 - 基于 one-api relay 模式的实现")

	// 示例 1: 创建 OpenAI 提供商
	fmt.Println("\n--- 示例 1: 创建 OpenAI 提供商 ---")
	openaiConfig := llm2.ProviderConfig{
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "your-api-key-here",
	}

	provider, err := llm2.NewProvider(llm2.ProviderOpenAI, openaiConfig)
	if err != nil {
		log.Printf("创建 OpenAI 提供商失败: %v", err)
	} else {
		fmt.Println("OpenAI 提供商创建成功")
		fmt.Printf("提供商类型: %T\n", provider)
	}

	// 示例 2: 创建 MCP 提供商
	fmt.Println("\n--- 示例 2: 创建 MCP 提供商 ---")
	mcpConfig := llm2.ProviderConfig{
		BaseURL: "http://localhost:8080/mcp",
		APIKey:  "mcp-config-key",
	}

	mcpProvider, err := llm2.NewProvider(llm2.ProviderMCP, mcpConfig)
	if err != nil {
		log.Printf("创建 MCP 提供商失败: %v", err)
	} else {
		fmt.Println("MCP 提供商创建成功")
		fmt.Printf("提供商类型: %T\n", mcpProvider)
	}

	// 示例 3: 构造聊天请求
	fmt.Println("\n--- 示例 3: 构造聊天请求 ---")
	req := llm2.ChatRequest{
		Model: "gpt-4o",
		Messages: []llm2.Message{
			{
				Role:    "user",
				Content: "Hello, can you help me with my project?",
			},
		},
		Temperature: 0.7,
		Stream:      true,
	}

	fmt.Printf("请求模型: %s\n", req.Model)
	fmt.Printf("消息数: %d\n", len(req.Messages))
	fmt.Printf("流式: %t\n", req.Stream)

	// 示例 4: 执行流式聊天 (理论上的调用)
	fmt.Println("\n--- 示例 4: 执行流式聊天 (示例) ---")

	// 注意：由于我们没有真实的服务，这里只演示调用结构
	if provider != nil {
		fmt.Println("准备发起聊天请求...")
		stream, err := provider.StreamChat(context.Background(), req)
		if err != nil {
			log.Printf("聊天请求失败: %v", err)
		} else {
			fmt.Println("聊天流已启动")

			// 处理响应流
			go func() {
				count := 0
				for resp := range stream {
					if resp.Error != nil {
						log.Printf("流错误: %v", resp.Error)
						break
					}
					fmt.Printf("收到响应 %d: %s (Tokens: %d)\n", count, resp.Content,
						resp.Usage.TotalTokens)
					count++
					if count >= 3 { // 限制演示输出
						break
					}
				}
			}()
		}
	}

	// 示例 5: 输出支持的提供商类型
	fmt.Println("\n--- 示例 5: 支持的提供商类型 ---")
	providers := []llm2.ProviderType{
		llm2.ProviderOpenAI,
		llm2.ProviderAzureOpenAI,
		llm2.ProviderAnthropic,
		llm2.ProviderGoogle,
		llm2.ProviderMCP,
		llm2.ProviderQwen,
		llm2.ProviderZhipu,
	}

	for _, providerType := range providers {
		fmt.Printf("- %s\n", string(providerType))
	}

	fmt.Println("\n--- 实现特点 ---")
	fmt.Println("✓ 基于 one-api relay 模式的架构")
	fmt.Println("✓ 支持多种 LLM 提供商")
	fmt.Println("✓ MCP 工具集成支持")
	fmt.Println("✓ 统一的适配器接口")
	fmt.Println("✓ 可扩展的设计")
}

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// デフォルトのモデル名を設定
const defaultModel = "gpt-3.5-turbo"

// NewOpenAIClient はOpenAI APIキーとタイムアウトを使用して新しいクライアントを初期化します
func NewOpenAIClient(timeout int) (*openai.Client, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI APIキーが設定されていません")
	}

	// HTTPクライアントの設定（タイムアウト付き）
	httpClient := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// OpenAIクライアントの初期化
	openaiConfig := openai.DefaultConfig(apiKey)
	openaiConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(openaiConfig)

	return client, nil
}

// ExecuteChatCompletion はOpenAI APIにリクエストを送り、アシスタントの応答を取得します
func ExecuteChatCompletion(client *openai.Client, model string, maxTokens *int, conversationHistory []openai.ChatCompletionMessage) (openai.ChatCompletionMessage, error) {
	ctx := context.Background()

	// // MaxTokensをポインタ型に変更
	// var maxTokensPtr *int
	// if max_tokens > 0 {
	// 	maxTokensPtr = &max_tokens
	// }

	// ChatCompletionRequest の作成
	chatRequest := openai.ChatCompletionRequest{
		Model:    model,
		Messages: conversationHistory,
	}

	// MaxTokensが存在する場合に設定
	// if maxTokensPtr != nil {
	// 	chatRequest.MaxTokens = *maxTokensPtr
	// }
	if maxTokens != nil {
		chatRequest.MaxTokens = *maxTokens
	}

	resp, err := client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		return openai.ChatCompletionMessage{}, fmt.Errorf("ChatCompletionエラー: %w", err)
	}

	if len(resp.Choices) == 0 {
		return openai.ChatCompletionMessage{}, fmt.Errorf("ChatCompletionエラー: 返されたChoicesが空です")
	}
	assistantMessage := resp.Choices[0].Message
	return assistantMessage, nil
}

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
func ExecuteChatCompletion(client *openai.Client, model string, max_tokens int, conversationHistory []openai.ChatCompletionMessage) (openai.ChatCompletionMessage, error) {
	ctx := context.Background()

	// MaxTokensをポインタ型に変更
	var maxTokensPtr *int
	if max_tokens > 0 {
		maxTokensPtr = &max_tokens
	}

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:     model,
			Messages:  conversationHistory,
			MaxTokens: *maxTokensPtr,
		},
	)
	if err != nil {
		return openai.ChatCompletionMessage{}, fmt.Errorf("ChatCompletionエラー: %w", err)
	}

	assistantMessage := resp.Choices[0].Message
	return assistantMessage, nil
}

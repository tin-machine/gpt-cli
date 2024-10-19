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

func ExecuteChatCompletion(client *openai.Client, model string, conversationHistory []openai.ChatCompletionMessage) (openai.ChatCompletionMessage, error) {
	ctx := context.Background()

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: conversationHistory,
		},
	)
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}

	assistantMessage := resp.Choices[0].Message
	return assistantMessage, nil
}

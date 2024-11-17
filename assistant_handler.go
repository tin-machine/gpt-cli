package main

import (
	"bufio"
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"os"
	"strings"
)

// assistant.go またはヘルパーモジュールに追加
func StringPtr(s string) *string {
	return &s
}

func Float32Ptr(f float32) *float32 {
	return &f
}

func createNewAssistant(client *openai.Client, options Options) (string, error) {
	ctx := context.Background()

	assistantRequest := openai.AssistantRequest{
		Name:         &options.AssistantName,
		Description:  &options.AssistantDescription,
		Model:        options.Model,
		Instructions: &options.Instruction,
		Tools: []openai.AssistantTool{
			{
				Type: openai.AssistantToolTypeCodeInterpreter,
			},
			{
				Type: openai.AssistantToolTypeFileSearch,
			},
		},
		Temperature: Float32Ptr(float32(options.Temperature)),
		Metadata:    options.Metadata,
	}

	assistant, err := client.CreateAssistant(ctx, assistantRequest)
	if err != nil {
		logger.Error("アシスタントを作成中にエラーが発生しました: %v", err)
		return "", fmt.Errorf("アシスタントを作成中にエラーが発生しました: %w", err)
	}

	logger.Info("アシスタントが作成されました:\nID: %s\nName: %s\n", assistant.ID, *assistant.Name)
	// fmt.Printf("アシスタントが作成されました:\nID: %s\nName: %s\n", assistant.ID, *assistant.Name)
	return assistant.ID, nil
}

func chatWithAssistant(client *openai.Client, assistantID string, options Options) error {
	ctx := context.Background()

	// アシスタントの取得
	assistant, err := client.RetrieveAssistant(ctx, assistantID)
	if err != nil {
		return fmt.Errorf("アシスタント(%s)の取得に失敗しました: %w", assistantID, err)
	}

	// Instructionsのnilチェックを追加
	var instructions string
	if assistant.Instructions != nil {
		instructions = *assistant.Instructions
	} else {
		instructions = "あなたはユーザーを助けるアシスタントです。"
	}

	// メッセージの構築
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: instructions,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: options.UserMessage,
		},
	}

	// チャットのリクエスト
	chatRequest := openai.ChatCompletionRequest{
		Model:    assistant.Model,
		Messages: messages,
	}

	if options.Temperature != 0 {
		chatRequest.Temperature = float32(options.Temperature)
	}

	if assistant.TopP != nil {
		chatRequest.TopP = *assistant.TopP
	}

	// チャットの実行
	chatResponse, err := client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		return fmt.Errorf("チャットの実行に失敗しました: %w", err)
	}

	// Choicesが空でないことを確認
	if len(chatResponse.Choices) == 0 {
		return fmt.Errorf("チャットの実行に失敗しました: 返されたChoicesが空です")
	}

	// 応答の表示
	fmt.Printf("アシスタントの応答:\n%s\n", chatResponse.Choices[0].Message.Content)
	return nil
}

func interactiveChatWithAssistant(client *openai.Client, assistantID string, options Options) error {
	ctx := context.Background()

	// アシスタントの取得
	assistant, err := client.RetrieveAssistant(ctx, assistantID)
	if err != nil {
		return fmt.Errorf("アシスタント(%s)の取得に失敗しました: %w", assistantID, err)
	}

	// Instructionsのnilチェックを追加
	var instructions string
	if assistant.Instructions != nil {
		instructions = *assistant.Instructions
	} else {
		instructions = "あなたはユーザーを助けるアシスタントです。" // デフォルトの指示
	}

	// 会話履歴を初期化
	conversationHistory := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: instructions,
		},
	}

	// インタラクティブなチャットを開始
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("アシスタントとの対話を開始します。（終了するには 'exit' と入力してください）")
	for {
		fmt.Print("あなた: ")
		if !scanner.Scan() {
			break
		}
		userInput := scanner.Text()
		if strings.ToLower(strings.TrimSpace(userInput)) == "exit" {
			fmt.Println("チャットを終了します。")
			break
		}

		// ユーザーメッセージを履歴に追加
		conversationHistory = append(conversationHistory, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: userInput,
		})

		// チャットのリクエストを作成
		chatRequest := openai.ChatCompletionRequest{
			Model:    assistant.Model,
			Messages: conversationHistory,
		}

		if assistant.Temperature != nil {
			chatRequest.Temperature = *assistant.Temperature
		}

		if assistant.TopP != nil {
			chatRequest.TopP = *assistant.TopP
		}

		// チャットの実行
		chatResponse, err := client.CreateChatCompletion(ctx, chatRequest)
		if err != nil {
			return fmt.Errorf("チャットの実行に失敗しました: %w", err)
		}

		// Choicesが空でないことを確認
		if len(chatResponse.Choices) == 0 {
			return fmt.Errorf("チャットの実行に失敗しました: 返されたChoicesが空です")
		}

		assistantReply := chatResponse.Choices[0].Message.Content

		// アシスタントの応答を履歴に追加
		conversationHistory = append(conversationHistory, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: assistantReply,
		})

		// アシスタントの応答を表示
		fmt.Printf("アシスタント: %s\n", assistantReply)
	}

	return nil
}

func handleCreateAssistant(client *openai.Client, options Options) error {
	assistantID, err := createNewAssistant(client, options)
	if err != nil {
		return fmt.Errorf("アシスタントの作成に失敗しました: %v", err)
	}
	fmt.Printf("新しいアシスタントが作成されました。ID: %s\n", assistantID)
	return nil
}

func handleAssistantInteraction(client *openai.Client, options Options) error {
	if options.Message != "" {
		// 単一のメッセージを送信
		err := chatWithAssistant(client, options.AssistantID, options)
		if err != nil {
			return fmt.Errorf("アシスタントとのチャットに失敗しました: %v", err)
		}
		return nil
	} else {
		// 対話モードを開始
		err := interactiveChatWithAssistant(client, options.AssistantID, options)
		if err != nil {
			return fmt.Errorf("アシスタントとのチャットに失敗しました: %v", err)
		}
		return nil
	}
}

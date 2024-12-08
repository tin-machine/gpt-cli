package main

import (
	"bufio"
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"os"
	"strings"
)

// StringPtrは、渡された文字列をポインタ型（*string）に変換して返します。
// これにより、他の関数で文字列を参照できるようになります。
func StringPtr(s string) *string {
	return &s
}

// Float32Ptrは、渡されたfloat32値をポインタ型（*float32）に変換して返します。
// これにより、他の関数でfloat32を参照できるようになります。
func Float32Ptr(f float32) *float32 {
	return &f
}

// createNewAssistantは、新しいアシスタントを作成します。
// この関数は、指定されたオプションを使用してアシスタントの名前、説明、モデル、およびその他の設定を定義します。
// clientはOpenAI APIクライアントであり、Options構造体にはアシスタントのための設定が含まれます。
// アシスタントが正常に作成されると、そのアシスタントのIDを返します。
// エラーが発生した場合は、エラーメッセージを返します。
func createNewAssistant(client *openai.Client, options Options) (string, error) {
	ctx := context.Background()

	// Toolsや関連する設定の宣言
	var vectorStoreID string
	if options.VectorStoreID != "" {
		vectorStoreID = options.VectorStoreID
	} else {
		vectorStoreID = DefaultVectorStoreID // オプションが提供されない場合のデフォルト
	}

	// ベクトルストア名で取得または作成
	if options.VectorStoreName != "" {
		vectorStore, err := GetOrCreateVectorStoreByName(client, options.VectorStoreName)
		if err != nil {
			logger.Error("ベクトルストアの取得/作成中にエラーが発生しました: %v", err)
			return "", fmt.Errorf("ベクトルストアの取得または作成に失敗しました: %w", err)
		}
		vectorStoreID = vectorStore.ID
	} else {
		// デフォルトのベクトルストアIDを使用
		vectorStoreID = DefaultVectorStoreID
	}

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
		ToolResources: &openai.AssistantToolResource{
			FileSearch: &openai.AssistantToolFileSearch{
				VectorStoreIDs: []string{vectorStoreID},
			},
		},
		Temperature: Float32Ptr(float32(options.Temperature)),
		Metadata:    options.Metadata,
	}

	assistant, err := client.CreateAssistant(ctx, assistantRequest)
	if err != nil {
		logger.Error("アシスタントを作成中にエラーが発生しました: %v", err)
		return "", fmt.Errorf("アシスタントの作成に失敗しました。設定を確認するか、再試行してください。詳細: %w", err)
	}

	logger.Info("アシスタントが作成されました:\nID: %s\nName: %s\n", assistant.ID, *assistant.Name)
	return assistant.ID, nil
}

// chatWithAssistantは、指定されたアシスタントに対してユーザーからのメッセージを送信し、アシスタントの応答を表示します。
// clientはOpenAI APIクライアント、assistantIDは対象のアシスタントのIDを示します。
// optionsにはユーザーが入力したメッセージやその他の設定が含まれます。
// アシスタントの応答が表示され, エラーが発生した場合はその内容が返されます。
func chatWithAssistant(client *openai.Client, assistantID string, options Options) error {
	logger.Info("chatWithAssistantです:\nassistantID: %s\nUserMassage: %s\n", assistantID, options.UserMessage)
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

// interactiveChatWithAssistantは、指定されたアシスタントとのインタラクティブな対話を開始します。
// ユーザーが標準入力からメッセージを入力すると、アシスタントに送信され、その応答が表示されます。
// パラメータとしてclinetはOpenAI APIクライアント、assistantIDは対象のアシスタントのID、およびoptionsは設定が含まれます。
// ユーザーが“exit”と入力するまで会話は続きます。
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

func handleCreateAssistant(client *openai.Client, options Options, config Config) error {
	assistantConfig, ok := config.Assistants[options.AssistantOption] // 例えば、コマンドラインオプションとして指定されたassistant名を使う場合
	if !ok {
		return fmt.Errorf("指定されたアシスタントの設定が見つかりません: %s", options.PromptOption)
	}

	assistantID, err := createNewAssistant(client, Options{
		AssistantName:        assistantConfig.Name,
		AssistantDescription: assistantConfig.Description,
		Model:                assistantConfig.Model,
		Instruction:          assistantConfig.Instruction,
		Temperature:          assistantConfig.Temperature,
		VectorStoreName:      assistantConfig.VectorStoreName,
	})
	if err != nil {
		return fmt.Errorf("アシスタントの作成に失敗しました: %v", err)
	}
	fmt.Printf("新しいアシスタントが作成されました。ID: %s\n", assistantID)
	return nil
}

// // handleCreateAssistantは、ユーザーが新しいアシスタントを作成する要求を処理します。
// // この関数では、OpenAI APIクライアントとOptionsが渡されて、アシスタントを作成するための内部関数を呼び出します。
// // アシスタントが作成されると、そのIDがコンソールに表示されます。
// // もしアシスタントの作成に失敗した場合は、エラーメッセージが返されます。
// func handleCreateAssistant(client *openai.Client, options Options) error {
// 	assistantID, err := createNewAssistant(client, options)
// 	if err != nil {
// 		return fmt.Errorf("アシスタントの作成に失敗しました: %v", err)
// 	}
// 	fmt.Printf("新しいアシスタントが作成されました。ID: %s\n", assistantID)
// 	return nil
// }

// handleAssistantInteractionは、アシスタントとのインタラクションを処理します。
// この関数では、ユーザーが送信したメッセージの有無に応じて、アシスタントとの単発チャットまたはインタラクティブなあるいは対話モードを開始します。
// clientはOpenAI APIクライアント、optionsにはユーザーの入力や設定が含まれます。
// いずれかの操作の実行中にエラーが発生した場合は、その内容が返されます。
func handleAssistantInteraction(client *openai.Client, options Options) error {
	var assistantID string
	if options.AssistantID != "" {
		assistantID = options.AssistantID
	} else if options.AssistantName != "" {
		id, err := GetAssistantIDByName(client, options.AssistantName)
		if err != nil {
			return err
		}
		assistantID = id
	} else {
		return fmt.Errorf("アシスタントのIDまたは名前を指定してください (--assistant-id または --assistant-name)")
	}

	// options.Messageをoptions.UserMessageに変更
	if options.UserMessage != "" {
		// 単一のメッセージを送信
		options.AssistantID = assistantID
		err := chatWithAssistant(client, assistantID, options)
		if err != nil {
			return fmt.Errorf("アシスタントとのチャットに失敗しました: %v", err)
		}
		return nil
	} else {
		// 対話モードを開始
		options.AssistantID = assistantID
		err := interactiveChatWithAssistant(client, assistantID, options)
		if err != nil {
			return fmt.Errorf("アシスタントとのチャットに失敗しました: %v", err)
		}
		return nil
	}

	// if options.Message != "" {
	// 	// 単一のメッセージを送信
	// 	err := chatWithAssistant(client, options.AssistantID, options)
	// 	if err != nil {
	// 		return fmt.Errorf("アシスタントとのチャットに失敗しました: %v", err)
	// 	}
	// 	return nil
	// } else {
	// 	// 対話モードを開始
	// 	err := interactiveChatWithAssistant(client, options.AssistantID, options)
	// 	if err != nil {
	// 		return fmt.Errorf("アシスタントとのチャットに失敗しました: %v", err)
	// 	}
	// 	return nil
	// }
}

// GetAssistantIDByName は指定された名前のアシスタントのIDを取得します
func GetAssistantIDByName(client *openai.Client, name string) (string, error) {
	ctx := context.Background()

	// ページネーションのパラメータを初期化
	limit := 100 // 一度に取得するアシスタントの数
	var order *string = nil
	var after *string = nil
	var before *string = nil

	for {
		// アシスタント一覧を取得
		assistantsList, err := client.ListAssistants(ctx, &limit, order, after, before)
		if err != nil {
			return "", fmt.Errorf("アシスタント一覧の取得に失敗しました: %w", err)
		}

		// 名前が一致するアシスタントを検索
		for _, assistant := range assistantsList.Assistants {
			if assistant.Name != nil && *assistant.Name == name {
				return assistant.ID, nil
			}
		}

		// 次のページがあるか確認
		if assistantsList.HasMore && assistantsList.LastID != nil {
			after = assistantsList.LastID
		} else {
			// すべてのアシスタントを検索済み
			break
		}
	}

	return "", fmt.Errorf("指定された名前のアシスタントが見つかりませんでした: %s", name)

	// // アシスタント一覧を取得
	// assistants, err := client.ListAssistants(ctx, openai.Pagination{})
	// if err != nil {
	//   return "", fmt.Errorf("アシスタント一覧の取得に失敗しました: %w", err)
	// }

	// // 名前が一致するアシスタントを検索
	// for _, assistant := range assistants.Assistants {
	//   if assistant.Name != nil && *assistant.Name == name {
	//     return assistant.ID, nil
	//   }
	// }
	// return "", fmt.Errorf("指定された名前のアシスタントが見つかりませんでした: %s", name)
}

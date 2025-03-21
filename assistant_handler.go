package main

import (
	"bufio"
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"os"
	"strings"
	"time"
)

// StringPtr は、渡された文字列をポインタ型（*string）に変換して返します。
// これにより、他の関数で文字列を参照できるようになります。
func StringPtr(s string) *string {
	return &s
}

// Float32Ptr は、渡されたfloat32値をポインタ型（*float32）に変換して返します。
// これにより、他の関数でfloat32を参照できるようになります。
func Float32Ptr(f float32) *float32 {
	return &f
}

// createNewAssistant は、新しいアシスタントを作成します。
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
	} else if options.VectorStoreName != "" {
		vectorStore, err := GetOrCreateVectorStoreByName(client, options.VectorStoreName)
		if err != nil {
			logger.Error("VectorStoreの取得または作成に失敗しました: %v", err)
			return "", fmt.Errorf("VectorStoreの取得または作成に失敗しました: %w", err)
		}
		vectorStoreID = vectorStore.ID
	} else {
		vectorStoreID = DefaultVectorStoreID // オプションが提供されない場合のデフォルト
	}

	// // ベクトルストア名で取得または作成
	// if options.VectorStoreName != "" {
	// 	vectorStore, err := GetOrCreateVectorStoreByName(client, options.VectorStoreName)
	// 	if err != nil {
	// 		logger.Error("ベクトルストアの取得/作成中にエラーが発生しました: %v", err)
	// 		return "", fmt.Errorf("ベクトルストアの取得または作成に失敗しました: %w", err)
	// 	}
	// 	vectorStoreID = vectorStore.ID
	// } else {
	// 	// デフォルトのベクトルストアIDを使用
	// 	vectorStoreID = DefaultVectorStoreID
	// }

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

// chatWithAssistant は、指定されたアシスタントに対してユーザーからのメッセージを送信し、アシスタントの応答を表示します。
// clientはOpenAI APIクライアント、assistantIDは対象のアシスタントのIDを示します。
// optionsにはユーザーが入力したメッセージやその他の設定が含まれます。
// アシスタントの応答が表示され, エラーが発生した場合はその内容が返されます。
func chatWithAssistant(client *openai.Client, assistantID string, options Options) error {
	logger.Info("chatWithAssistantです。次のアシスタントを使用します:\nassistantID: %s\n", assistantID)
	ctx := context.Background()

	// アシスタントの取得
	assistant, err := client.RetrieveAssistant(ctx, assistantID)
	if err != nil {
		return fmt.Errorf("アシスタント(%s)の取得に失敗しました: %w", assistantID, err)
	}

	// メッセージの構築
	messages := []openai.ChatCompletionMessage{
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

// interactiveChatWithAssistant は、指定されたアシスタントとのインタラクティブな対話を開始します。
// ユーザーが標準入力からメッセージを入力すると、アシスタントに送信され、その応答が表示されます。
// パラメータとしてclinetはOpenAI APIクライアント、assistantIDは対象のアシスタントのID、およびoptionsは設定が含まれます。
// ユーザーが“exit”と入力するまで会話は続きます。
func interactiveChatWithAssistant(client *openai.Client, assistantID string, options Options) error {
	logger.Info("interactiveChatWithAssistantです。次のアシスタントを使用します:\nassistantID: %s\n", assistantID)
	ctx := context.Background()

	// 新規スレッド（会話セッション）の作成
	thread, err := client.CreateThread(ctx, openai.ThreadRequest{})
	if err != nil {
		return fmt.Errorf("スレッドの作成に失敗しました: %w", err)
	}
	logger.Info("新しいスレッドを作成しました。ID: %s", thread.ID)

	// ③ インタラクティブなチャットを開始
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

		// ユーザーメッセージをスレッドに追加
		_, err = client.CreateMessage(ctx, thread.ID, openai.MessageRequest{
			Role:    openai.ChatMessageRoleUser,
			Content: userInput,
		})
		if err != nil {
			return fmt.Errorf("ユーザーメッセージの送信に失敗しました: %w", err)
		}

		// ④ アシスタントの実行（Run）の開始
		run, err := client.CreateRun(ctx, thread.ID, openai.RunRequest{
			AssistantID: assistantID,
		})
		if err != nil {
			return fmt.Errorf("アシスタントの実行開始に失敗しました: %w", err)
		}
		logger.Info("Run開始: ID=%s", run.ID)

		// Runが queued または in_progress の間、ポーリングして実行完了を待つ
		for run.Status == openai.RunStatusQueued || run.Status == openai.RunStatusInProgress {
			time.Sleep(1 * time.Second)
			run, err = client.RetrieveRun(ctx, thread.ID, run.ID)
			if err != nil {
				return fmt.Errorf("runの取得に失敗しました: %w", err)
			}
		}
		logger.Info("Run完了: Status=%s", run.Status)

		// ⑤ スレッド内のメッセージ一覧を取得（全件取得または最新件数を指定）
		// ここでは全メッセージを取得して、最後のアシスタントからの応答を表示する例です
		messagesList, err := client.ListMessage(ctx, thread.ID, nil, StringPtr("desc"), nil, nil, nil)
		if err != nil {
			return fmt.Errorf("スレッドメッセージの取得に失敗しました: %w", err)
		}

		// 最新のアシスタントメッセージ（Roleがassistantの最後のメッセージ）を検索
		var assistantReply string
		for _, msg := range messagesList.Messages {
			if msg.Role == openai.ChatMessageRoleAssistant && len(msg.Content) > 0 {
				assistantReply = msg.Content[0].Text.Value
				break
			}
		}
		if assistantReply == "" {
			fmt.Println("アシスタントからの応答が見つかりませんでした。")
		} else {
			fmt.Printf("アシスタント: %s\n", assistantReply)
		}
	}

	return nil
}

func chooseString(cliValue, defaultValue string) string {
	if cliValue != "" {
		return cliValue
	}
	return defaultValue
}

func chooseFloat64(cliValue, defaultValue float64) float64 {
	// 例えば、cliValue が 0 としてしまうと「指定なし」とみなす場合
	if cliValue != 0 {
		return cliValue
	}
	return defaultValue
}

func handleCreateAssistant(client *openai.Client, options Options, config Config) error {
	logger.Info("handleCreateAssistant を実行します")
	logger.Info(options.AssistantOption)
	logger.Info(fmt.Sprintf("%v", config.Assistants))

	// まず --assistant-name がCLIで渡されているか確認
	if options.AssistantName == "" {
		return fmt.Errorf("アシスタント名（--assistant-name）が指定されていません")
	}

	// config.yaml の Assistants から、CLIで渡されたAssistantNameに該当する設定を検索
	assistantConfig, found := config.Assistants[options.AssistantName]

	// if found,各項目はconfig側をデフォルトとする
	// CLIで個別指定されている場合はCLIの値で上書き
	var finalOptions Options
	if found {
		finalOptions = Options{
			AssistantName:        options.AssistantName, // CLIの値は必須
			AssistantDescription: chooseString(options.AssistantDescription, assistantConfig.Description),
			Model:                chooseString(options.Model, assistantConfig.Model),
			Instruction:          chooseString(options.Instruction, assistantConfig.Instruction),
			Temperature:          chooseFloat64(options.Temperature, assistantConfig.Temperature),
			VectorStoreName:      chooseString(options.VectorStoreName, assistantConfig.VectorStoreName),
		}
	} else {
		// 設定が見つからない場合、CLIで必要な項目を全て指定しているかチェック
		missing := []string{}
		if options.AssistantDescription == "" {
			missing = append(missing, "--assistant-description")
		}
		if options.Model == "" {
			missing = append(missing, "--model")
		}
		if options.Instruction == "" {
			missing = append(missing, "--instruction")
		}
		// Temperature==0 なら「指定なし」とみなす（※実際には0を有効な値として利用する場合には別途判定）
		if options.Temperature == 0 {
			missing = append(missing, "--temperature")
		}
		if options.VectorStoreName == "" {
			missing = append(missing, "--vector-store-name")
		}
		if len(missing) > 0 {
			return fmt.Errorf("config.yamlに設定が見つからず、かつCLIで以下の必須オプションが指定されていません: %s", strings.Join(missing, ", "))
		}
		// CLIの入力のみを最終設定とする
		finalOptions = Options{
			AssistantName:        options.AssistantName,
			AssistantDescription: options.AssistantDescription,
			Model:                options.Model,
			Instruction:          options.Instruction,
			Temperature:          options.Temperature,
			VectorStoreName:      options.VectorStoreName,
		}
	}
	// ログ出力（確認用）
	logger.Info("最終的なアシスタント設定：Name: %s, Description: %s, Model: %s, Instruction: %s, Temperature: %f, VectorStoreName: %s",
		finalOptions.AssistantName,
		finalOptions.AssistantDescription,
		finalOptions.Model,
		finalOptions.Instruction,
		finalOptions.Temperature,
		finalOptions.VectorStoreName)

	assistantID, err := createNewAssistant(client, finalOptions)
	if err != nil {
		return fmt.Errorf("アシスタントの作成に失敗しました: %v", err)
	}

	fmt.Printf("新しいアシスタントが作成されました。ID: %s\n", assistantID)
	return nil
}

// handleAssistantInteraction は、アシスタントとのインタラクションを処理します。
// この関数では、ユーザーが送信したメッセージの有無に応じて、アシスタントとの単発チャットまたはインタラクティブなあるいは対話モードを開始します。
// clientはOpenAI APIクライアント、optionsにはユーザーの入力や設定が含まれます。
// いずれかの操作の実行中にエラーが発生した場合は、その内容が返されます。
func handleAssistantInteraction(client *openai.Client, options Options) error {
	ctx := context.Background()
	var assistantID string

	// もしコマンドラインで AssistantID が指定されていなければ、AssistantName から検索する
	if options.AssistantID != "" {
		assistantID = options.AssistantID
	} else if options.AssistantName != "" {
		// 一覧取得（必要数は十分大きな数を指定するか、ページネーションに対応）
		limit := 100
		assistantsList, err := client.ListAssistants(ctx, &limit, nil, nil, nil)
		if err != nil {
			return fmt.Errorf("アシスタント一覧の取得に失敗しました: %w", err)
		}

		// AssistantName で一致するアシスタントを検索
		var found *openai.Assistant
		for _, asst := range assistantsList.Assistants {
			if asst.Name != nil && *asst.Name == options.AssistantName {
				found = &asst
				break
			}
		}
		if found == nil {
			return fmt.Errorf("指定されたアシスタント名 '%s' のアシスタントが見つかりませんでした", options.AssistantName)
		}
		assistantID = found.ID
	} else {
		return fmt.Errorf("アシスタントとの対話を開始するには、--assistant-id もしくは --assistant-name を指定してください")
	}

	if options.Message != "" {
		// 単一のメッセージを送信
		err := chatWithAssistant(client, assistantID, options)
		if err != nil {
			return fmt.Errorf("アシスタントとのチャットに失敗しました: %v", err)
		}
		return nil
	}

	// 対話モードを開始
	err := interactiveChatWithAssistant(client, assistantID, options)
	if err != nil {
		return fmt.Errorf("アシスタントとのチャットに失敗しました: %v", err)
	}
	return nil
}

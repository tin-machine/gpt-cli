package main

import (
	"fmt"
	"log"
)

var Version string
var logger Logger

func main() {
	if err := Run(); err != nil {
		log.Fatalf("プログラムの実行中にエラーが発生しました: %v", err)
	}
}

// Run はプログラムのメイン処理を実行します
func Run() error {

	// コマンドライン引数の解析
	options, err := ParseCommandLineArgs()
	if err != nil {
		return err
	}

	// ロギングの設定
	logger = NewConsoleLogger(options.Debug)

	// バージョン情報の表示
	if options.ShowVersion {
		fmt.Printf("Version: %s\n", Version)
		return nil
	}

	// ロギングの設定
	SetupLogging(options.Debug)

	// 設定ファイルの読み込み
	config, err := LoadConfiguration(options.ConfigPath)
	if err != nil {
		return err
	}

	// ログディレクトリの設定と検証
	err = ConfigureLogDirectory(&options, config)
	if err != nil {
		return err
	}

	// ユーザーメッセージの構築
	err = BuildUserMessage(&options)
	if err != nil {
		return err
	}

	// プロンプト設定の取得
	promptConfig, err := GetPromptConfig(config, options)
	if err != nil {
		return err
	}

	// コンテキストメッセージの作成
	messages, err := CreateMessages(promptConfig)
	if err != nil {
		return fmt.Errorf("メッセージの作成に失敗しました: %w", err)
	}

	// 会話履歴の読み込み
	conversationHistory, err := LoadConversationHistory(options.HistoryFile)
	if err != nil {
		return fmt.Errorf("会話履歴の読み込みに失敗しました: %w", err)
	}

	// show-history オプションの処理
	if options.ShowHistory != "" {
		conversationHistory, err := LoadConversationHistory(options.ShowHistory)
		if err != nil {
			return fmt.Errorf("会話履歴の読み込みに失敗しました: %w", err)
		}
		if len(conversationHistory) == 0 {
			fmt.Println("会話履歴はありません。")
			return nil
		}
		DisplayConversationHistory(conversationHistory)
		return nil
	}

	// 会話履歴に新しいメッセージを追加
	conversationHistory = append(conversationHistory, messages...)

	// OpenAI API クライアントの初期化
	client, err := NewOpenAIClient(options.Timeout)
	if err != nil {
		return err
	}

	// // ツール設定の読み込み
	// toolConfig, err := LoadToolConfig(options.ToolConfigPath)
	// if err != nil {
	//     return err
	// }

	// 各オプションに応じた処理を実行
	if options.CreateAssistant {
		return handleCreateAssistant(client, options)
	}

	if options.AssistantID != "" {
		return handleAssistantInteraction(client, options)
	}

	if len(options.UploadAndAddFiles) > 0 {
		return handleUploadAndAddFiles(client, options)
	}

	if options.UploadFilePath != "" {
		return handleUploadFile(client, options)
	}

	if options.ListFiles {
		return handleListFiles(client)
	}

	if options.DeleteFileID != "" {
		return handleDeleteFile(client, options)
	}

	if options.VectorStoreAction != "" {
		return handleVectorStoreAction(client, options)
	}

	// OpenAI API へのリクエスト
	return handleChatCompletion(client, promptConfig, conversationHistory, options)

	// if options.CreateAssistant {
	// 	// 新しいアシスタントを作成
	// 	assistantID, err := createNewAssistant(client, options)
	// 	if err != nil {
	// 		return fmt.Errorf("アシスタントの作成に失敗しました: %v", err)
	// 	}
	// 	fmt.Printf("新しいアシスタントが作成されました。ID: %s\n", assistantID)
	// 	return nil
	// }

	// if options.AssistantID != "" {
	// 	if options.Message != "" {
	// 		// 単一のメッセージを送信
	// 		err = chatWithAssistant(client, options.AssistantID, options)
	// 		if err != nil {
	// 			return fmt.Errorf("アシスタントとのチャットに失敗しました: %v", err)
	// 		}
	// 		return nil
	// 	} else {
	// 		// 対話モードを開始
	// 		err = interactiveChatWithAssistant(client, options.AssistantID, options)
	// 		if err != nil {
	// 			return fmt.Errorf("アシスタントとのチャットに失敗しました: %v", err)
	// 		}
	// 		return nil
	// 	}
	// 	// オプションが指定されていない場合のメッセージ
	// 	// fmt.Println("新しいアシスタントを作成するには --create-assistant を指定してください。既存のアシスタントと対話するには --assistant-id を指定してください。")
	// 	// return nil
	// }

	// // --upload-and-add-to-vector オプションの処理
	// if len(options.UploadAndAddFiles) > 0 {
	// 	// ファイルをアップロード
	// 	fileIDs, err := UploadFiles(client, options.UploadAndAddFiles, options.UploadPurpose)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// VectorStoreの取得または作成
	// 	vectorStoreName := options.VectorStoreName
	// 	if vectorStoreName == "" {
	// 		vectorStoreName = fmt.Sprintf("Auto-Generated Vector Store %d", time.Now().Unix())
	// 	}
	// 	vectorStore, err := GetOrCreateVectorStore(client, vectorStoreName)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Printf("使用するベクトルストア: ID=%s, Name=%s\n", vectorStore.ID, vectorStore.Name)

	// 	// ファイルをVectorStoreに追加
	// 	err = AddFilesToVectorStore(client, vectorStore.ID, fileIDs)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Printf("ファイルをベクトルストアに追加しました: VectorStoreID=%s\n", vectorStore.ID)

	// 	return nil // 処理が完了したので終了
	// }

	// // ファイルのアップロード
	// if options.UploadFilePath != "" {
	// 	// ファイルの存在チェック
	// 	if _, err := os.Stat(options.UploadFilePath); os.IsNotExist(err) {
	// 		return fmt.Errorf("指定されたファイルが見つかりません: %s", options.UploadFilePath)
	// 	}

	// 	// UploadFile 関数を呼び出す
	// 	uploadedFile, err := UploadFile(client, options.UploadFilePath, options.UploadPurpose)
	// 	if err != nil {
	// 		return fmt.Errorf("ファイルのアップロードに失敗しました: %v", err)
	// 	}
	// 	fmt.Printf("ファイルがアップロードされました。File ID: %s\n", uploadedFile.ID)
	// 	return nil // ファイルのアップロード後にプログラムを終了
	// }

	// // ファイル一覧表示の処理
	// if options.ListFiles {
	// 	files, err := ListUploadedFiles(client)
	// 	if err != nil {
	// 		return fmt.Errorf("ファイル一覧の取得に失敗しました: %v", err)
	// 	}
	// 	fmt.Println("アップロードされたファイル一覧:")
	// 	for _, file := range files.Files {
	// 		fmt.Printf("- ID: %s, 名前: %s, ステータス: %s\n", file.ID, file.FileName, file.Status)
	// 	}
	// 	return nil
	// }

	// // ファイル削除の処理
	// if options.DeleteFileID != "" {
	// 	err := DeleteUploadedFile(client, options.DeleteFileID)
	// 	if err != nil {
	// 		return fmt.Errorf("ファイルの削除に失敗しました: %v", err)
	// 	}
	// 	fmt.Printf("ファイルを削除しました。File ID: %s\n", options.DeleteFileID)
	// 	return nil
	// }

	// // ベクトルストアのアクションを処理
	// if options.VectorStoreAction != "" {
	// 	switch options.VectorStoreAction {
	// 	case "create":
	// 		if options.VectorStoreName == "" {
	// 			return fmt.Errorf("ベクトルストアの名前を指定してください (--vector-store-name)")
	// 		}
	// 		vs, err := CreateVectorStore(client, options.VectorStoreName)
	// 		if err != nil {
	// 			return fmt.Errorf("ベクトルストアの作成に失敗しました: %v", err)
	// 		}
	// 		fmt.Printf("ベクトルストアを作成しました: ID=%s, Name=%s\n", vs.ID, vs.Name)
	// 		return nil
	// 	case "list":
	// 		vsList, err := ListVectorStores(client)
	// 		if err != nil {
	// 			return fmt.Errorf("ベクトルストアの一覧取得に失敗しました: %v", err)
	// 		}
	// 		for _, vs := range vsList {
	// 			fmt.Printf("ID: %s, Name: %s, Status: %s\n", vs.ID, vs.Name, vs.Status)
	// 		}
	// 		return nil
	// 	case "delete":
	// 		if options.VectorStoreID == "" {
	// 			return fmt.Errorf("削除するベクトルストアのIDを指定してください (--vector-store-id)")
	// 		}
	// 		err := DeleteVectorStore(client, options.VectorStoreID)
	// 		if err != nil {
	// 			return fmt.Errorf("ベクトルストアの削除に失敗しました: %v", err)
	// 		}
	// 		fmt.Printf("ベクトルストアを削除しました: ID=%s\n", options.VectorStoreID)
	// 		return nil
	// 	case "add-file":
	// 		if options.VectorStoreID == "" || (options.FileID == "" && len(options.FileIDs) == 0) {
	// 			return fmt.Errorf("ベクトルストアIDとファイルIDを指定してください (--vector-store-id, --file-id または --file-ids)")
	// 		}
	// 		if options.FileID != "" {
	// 			// 単一のファイルIDを処理
	// 			vsFile, err := AddFileToVectorStore(client, options.VectorStoreID, options.FileID)
	// 			if err != nil {
	// 				return fmt.Errorf("ファイルの追加に失敗しました: %v", err)
	// 			}
	// 			fmt.Printf("ファイルをベクトルストアに追加しました: FileID=%s, VectorStoreID=%s\n", vsFile.ID, vsFile.VectorStoreID)
	// 		} else if len(options.FileIDs) > 0 {
	// 			// 複数のファイルIDを処理
	// 			err := AddFilesToVectorStore(client, options.VectorStoreID, options.FileIDs)
	// 			if err != nil {
	// 				return fmt.Errorf("複数ファイルの追加に失敗しました: %v", err)
	// 			}
	// 			fmt.Printf("%d 個のファイルをベクトルストアに追加しました: VectorStoreID=%s\n", len(options.FileIDs), options.VectorStoreID)
	// 		} else {
	// 			return fmt.Errorf("ファイルIDを指定してください (--file-id または --file-ids)")
	// 		}
	// 		return nil
	// 	default:
	// 		return fmt.Errorf("不正なベクトルストアアクションが指定されました: %s", options.VectorStoreAction)
	// 	}
	// }

	// // OpenAI API へのリクエスト
	// assistantMessage, err := ExecuteChatCompletion(client, promptConfig.Model, promptConfig.MaxTokens, conversationHistory)
	// if err != nil {
	// 	return fmt.Errorf("ChatCompletionエラー: %w", err)
	// }

	// // 会話履歴にアシスタントの応答を追加
	// conversationHistory = append(conversationHistory, assistantMessage)

	// // 会話履歴の保存
	// if options.HistoryFile != "" {
	// 	err = SaveConversationHistory(options.HistoryFile, conversationHistory)
	// 	if err != nil {
	// 		return fmt.Errorf("会話履歴の保存に失敗しました: %w", err)
	// 	}
	// }

	// // 標準出力に結果を表示
	// fmt.Println(assistantMessage.Content)

	// return nil
}

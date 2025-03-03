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

	// 会話履歴の読み込み
	conversationHistory, err := LoadConversationHistory(options.HistoryFile)
	if err != nil {
		return fmt.Errorf("会話履歴の読み込みに失敗しました: %w", err)
	}

	// OpenAI API クライアントの初期化
	client, err := NewOpenAIClient(options.Timeout)
	if err != nil {
		return err
	}

	// ファイルアップロードとベクトルストア追加
	if len(options.UploadAndAddFiles) > 0 {
		err := handleUploadAndAddFiles(client, options)
		if err != nil {
			return fmt.Errorf("ファイルのアップロードまたは追加に失敗しました: %v", err)
		}
	}

	// ファイルアップロードとベクトルストア追加
	if options.AssistantID != "" {
		err := handleUploadAndAddFiles(client, options)
		if err != nil {
			return fmt.Errorf("ファイルのアップロードまたは追加に失敗しました: %v", err)
		}
	}

	// アシスタントの作成または取得
	// AssistantOption( -aオプションが)指定されている場合、アシスタントを作成または取得
	if options.AssistantName != "" {
		// アシスタントを作成または取得
		err := handleCreateAssistant(client, options, config)
		if err != nil {
			return fmt.Errorf("アシスタントの作成に失敗しました: %v", err)
		}

		// アシスタントとの対話を開始
		err = handleAssistantInteraction(client, options)
		if err != nil {
			return fmt.Errorf("アシスタントとの対話に失敗しました: %v", err)
		}

		return nil
	}

	// // アシスタントの作成
	// if options.CreateAssistant {
	// 	err := handleCreateAssistant(client, options, config)
	// 	if err != nil {
	// 		return fmt.Errorf("アシスタントの作成に失敗しました: %v", err)
	// 	}
	// }

	// 他のオプションに応じた処理
	if options.UploadFilePath != "" {
		err := handleUploadFile(client, options)
		if err != nil {
			return fmt.Errorf("ファイルのアップロードに失敗しました: %v", err)
		}
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

	if options.DeleteFileName != "" {
		return handleDeleteFile(client, options)
	}

	// コンテキストメッセージの作成
	messages, err := CreateMessages(promptConfig)
	if err != nil {
		return fmt.Errorf("メッセージの作成に失敗しました: %w", err)
	}

	// デフォルトプロンプトを設定
	if promptConfig.System == "" && promptConfig.User == "" {
		promptConfig = GetDefaultPromptConfig()
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

	// OpenAI API へのリクエスト
	if options.UserMessage != "" {
		return handleChatCompletion(client, promptConfig, conversationHistory, options)
	}

	// どの条件にも一致しない場合のデフォルトの戻り値
	return nil
}

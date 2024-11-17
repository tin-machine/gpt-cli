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

	// デフォルトプロンプトを設定
	if promptConfig.System == "" && promptConfig.User == "" {
		promptConfig = GetDefaultPromptConfig()
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
}

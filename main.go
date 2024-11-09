package main

import (
	"fmt"
	"log"
)

var Version string

func main() {
	if err := Run(); err != nil {
		log.Fatalf("エラーが発生しました: %v", err)
	}
}

// Run はプログラムのメイン処理を実行します
func Run() error {

	// コマンドライン引数の解析
	options, err := ParseCommandLineArgs()
	if err != nil {
		return err
	}

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

	// OpenAI API へのリクエスト
	assistantMessage, err := ExecuteChatCompletion(client, promptConfig.Model, conversationHistory)
	if err != nil {
		return fmt.Errorf("ChatCompletionエラー: %w", err)
	}

	// 会話履歴にアシスタントの応答を追加
	conversationHistory = append(conversationHistory, assistantMessage)

	// 会話履歴の保存
	if options.HistoryFile != "" {
		err = SaveConversationHistory(options.HistoryFile, conversationHistory)
		if err != nil {
			return fmt.Errorf("会話履歴の保存に失敗しました: %w", err)
		}
	}

	// 標準出力に結果を表示
	fmt.Println(assistantMessage.Content)

	return nil
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var Version string

func main() {
	if err := Run(); err != nil {
		log.Fatalf("エラーが発生しました: %v", err)
	}
}

func Run() error {
	// コマンドラインオプションの定義
	promptOption := flag.String("p", "", "config.yamlにあるプロンプトを選択")
	systemMessage := flag.String("s", "", "Systemのメッセージを変更")
	userMessage := flag.String("u", "", "Userのメッセージを変更")
	imageList := flag.String("i", "", "画像ファイルをカンマ区切りで")
	configPath := flag.String("c", "", "設定ファイルのパスを指定")
	model := flag.String("m", "", "使用するモデルを指定")
	debug := flag.Bool("d", false, "デバッグモードを有効にする")
	showVersion := flag.Bool("version", false, "バージョン情報を表示")
	collectFiles := flag.Bool("collect", false, "現在のディレクトリ内のファイルをUserメッセージに追加")
	historyFile := flag.String("history", "", "会話履歴の保存ファイルを指定（拡張子は不要）")
	timeout := flag.Int("t", 60, "タイムアウト時間（秒）を指定")
	fileList := flag.String("f", "", "読み込むファイルのパスをカンマ区切りで指定")
	showHistory := flag.String("show-history", "", "会話履歴を表示")

	// フラグの解析
	flag.Parse()

	// バージョン情報の表示
	if *showVersion {
		fmt.Printf("Version: %s\n", Version)
		return nil
	}

	// ログレベルの設定
	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(0)
		log.SetOutput(os.Stderr)
	}

	// 設定ファイルのパスを取得
	configFilePath, err := GetConfigFilePath(*configPath)
	if err != nil {
		return err
	}

	// 設定ファイルの読み込み
	config, err := LoadConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("設定ファイルが読み込めません: %w", err)
	}

	// フラグ以外の引数を取得
	args := flag.Args()
	if len(args) > 0 {
		// 最後の引数をユーザープロンプトとして設定
		*userMessage += " " + args[len(args)-1]
	}

	// プロンプトの設定取得
	promptConfig, err := GetPromptConfig(config, *promptOption, *systemMessage, *userMessage, *showHistory, *model)
	if err != nil {
		return err
	}

	// 画像リストの処理
	if *imageList != "" {
		promptConfig.Attachments = SplitImageList(*imageList)
	}

	// -collect オプションが指定された場合、ファイルを収集
	if *collectFiles {
		filesContent, err := CollectFiles(".")
		if err != nil {
			return fmt.Errorf("ファイルの収集に失敗しました: %w", err)
		}
		promptConfig.User += "\n\n" + filesContent
	}

	// -f オプションが指定された場合、ファイルを読み込む
	if *fileList != "" {
		filesContent, err := ReadFiles(*fileList)
		if err != nil {
			return fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
		}
		promptConfig.User += "\n\n" + filesContent
	}

	// 会話履歴の読み込み
	conversationHistory, err := LoadConversationHistory(*historyFile)
	if err != nil {
		return fmt.Errorf("会話履歴の読み込みに失敗しました: %w", err)
	}

	if *showHistory != "" {
		conversationHistory, err := LoadConversationHistory(*showHistory)
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

	// メッセージの作成
	messages, err := CreateMessages(promptConfig)
	if err != nil {
		return fmt.Errorf("メッセージの作成に失敗しました: %w", err)
	}

	// 会話履歴にメッセージを追加
	conversationHistory = append(conversationHistory, messages...)

	// OpenAI API クライアントの初期化
	client, err := NewOpenAIClient(*timeout)
	if err != nil {
		return fmt.Errorf("OpenAIクライアントの初期化に失敗しました: %w", err)
	}

	// OpenAI API へのリクエスト
	assistantMessage, err := ExecuteChatCompletion(client, promptConfig.Model, conversationHistory)
	if err != nil {
		return fmt.Errorf("ChatCompletionエラー: %w", err)
	}

	// 会話履歴にアシスタントの応答を追加
	conversationHistory = append(conversationHistory, assistantMessage)

	// 会話履歴の保存
	if *historyFile != "" {
		err = SaveConversationHistory(*historyFile, conversationHistory)
		if err != nil {
			return fmt.Errorf("会話履歴の保存に失敗しました: %w", err)
		}
	}

	// 標準出力に結果を表示
	fmt.Println(assistantMessage.Content)

	return nil
}

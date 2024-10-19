package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

var Version string

func main() {
	// コマンドラインオプションの定義
	promptOption := flag.String("p", "", "config.yamlにあるプロンプトを選択")
	outputFile := flag.String("o", "", "出力するファイルを指定")
	systemMessage := flag.String("s", "", "Systemのメッセージを変更")
	userMessage := flag.String("u", "", "Userのメッセージを変更")
	imageList := flag.String("i", "", "画像ファイルをカンマ区切りで")
	configPath := flag.String("c", "", "設定ファイルのパスを指定")
	debug := flag.Bool("d", false, "デバッグモードを有効にする")
	showVersion := flag.Bool("version", false, "バージョン情報を表示")
	collectFiles := flag.Bool("collect", false, "現在のディレクトリ内のファイルをUserメッセージに追加")
	historyFile := flag.String("history", "", "会話履歴の保存ファイルを指定（拡張子は不要）")
	flag.Parse()

	// バージョン情報の表示
	if *showVersion {
		fmt.Printf("Version: %s\n", Version)
		return
	}

	// デバッグ用の出力関数
	debugPrintf := func(format string, args ...interface{}) {
		if *debug {
			log.Printf(format, args...)
		}
	}

	log.SetOutput(os.Stderr)
	debugPrintf("Version: %s\n", Version)

	// デフォルトの設定ファイルパス
	defaultConfigPath := filepath.Join(os.Getenv("HOME"), ".config", "gpt-cli", "config.yaml")
	if *configPath == "" {
		*configPath = defaultConfigPath
	}

	// 設定ファイルの読み込み
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("設定ファイルが読み込めません。-c <config.yaml> とパスを指定するか、%sに設置してください: %v", defaultConfigPath, err)
	}

	debugPrintf("Config: %v\n", config)

	var promptConfig Prompt
	if *promptOption != "" {
		// プロンプトの選択
		var ok bool
		promptConfig, ok = config.Prompts[*promptOption]
		if !ok {
			log.Fatalf("プロンプトオプション %s は設定ファイルに定義されていません", *promptOption)
		}
	} else if *systemMessage == "" && *userMessage == "" {
		log.Fatalf("-p(プロンプトの指定)が無い場合は-s(システムメッセージ)か-u(ユーザーメッセージ)の指定が必要です")
	}

	// コマンドライン引数からSystemメッセージを設定
	if *systemMessage != "" {
		promptConfig.System = *systemMessage
	}

	// コマンドライン引数からUserメッセージを設定
	if *userMessage != "" {
		promptConfig.User = *userMessage
	}

	// ここで、-collect オプションが指定された場合のみ CollectFiles を実行
	if *collectFiles {
		// 現在のディレクトリ内のファイルを収集
		filesContent, err := CollectFiles(".")
		if err != nil {
			log.Fatalf("ファイルの収集に失敗しました: %v", err)
		}
		// Userメッセージにファイル内容を追加
		promptConfig.User += "\n\n" + filesContent
	}

	// 画像リストの処理
	if *imageList != "" {
		promptConfig.Attachments = strings.Split(*imageList, ",")
	}

	debugPrintf("Prompt config: %v\n", promptConfig)

	// 会話履歴の初期化
	var conversationHistory []openai.ChatCompletionMessage

	// 履歴ファイルが指定されている場合は読み込む
	if *historyFile != "" {
		// 拡張子がない場合は .json を追加
		if filepath.Ext(*historyFile) == "" {
			*historyFile += ".json"
		}
		history, err := LoadConversationHistory(*historyFile)
		if err == nil {
			conversationHistory = history
		} else if !os.IsNotExist(err) {
			log.Fatalf("会話履歴の読み込みに失敗しました: %v", err)
		}
	}

	// システムメッセージを履歴に追加（最初に一度だけ）
	if promptConfig.System != "" && len(conversationHistory) == 0 {
		systemMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: promptConfig.System,
		}
		conversationHistory = append(conversationHistory, systemMessage)
	}

	// ユーザーメッセージを履歴に追加
	if promptConfig.User != "" {
		userMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: promptConfig.User,
		}
		conversationHistory = append(conversationHistory, userMessage)
	}

	// HTTPクライアントの設定（タイムアウト付き）
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	// OpenAIクライアントの初期化
	openaiConfig := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	openaiConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(openaiConfig)

	// デフォルトのモデル設定
	if promptConfig.Model == "" {
		promptConfig.Model = "gpt-4o"
	}

	// コンテキストにタイムアウトを設定
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// OpenAI APIへのリクエスト
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    promptConfig.Model,
			Messages: conversationHistory, // 会話の履歴全体を送信
		},
	)
	if err != nil {
		log.Fatalf("ChatCompletionエラー: %v", err)
	}

	// モデルからの応答を履歴に追加
	assistantMessage := resp.Choices[0].Message
	conversationHistory = append(conversationHistory, assistantMessage)

	debugPrintf("Response: %v\n", assistantMessage.Content)

	// 会話履歴を保存
	if *historyFile != "" {
		err = SaveConversationHistory(*historyFile, conversationHistory)
		if err != nil {
			log.Fatalf("会話履歴の保存に失敗しました: %v", err)
		}
	}

	// 出力ファイルの設定
	var outputFileName string
	if *outputFile != "" {
		outputFileName = *outputFile
	} else {
		dirName := fmt.Sprintf("%v", time.Now().Unix())
		err = os.Mkdir(dirName, 0755)
		if err != nil {
			log.Fatalf("ディレクトリの作成に失敗しました: %v", err)
		}
		outputFileName = fmt.Sprintf("%s/conversation.txt", dirName)
	}

	// 会話内容の保存
	err = SaveConversation(outputFileName, assistantMessage.Content)
	if err != nil {
		log.Fatalf("会話内容のファイル保存に失敗しました: %v", err)
	}
}

// CollectFilesは指定したディレクトリ内のファイル名と内容を収集します（.gitディレクトリを除く）
func CollectFiles(dir string) (string, error) {
	var builder strings.Builder

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// .gitディレクトリをスキップ
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// ファイルの場合、名前と内容を取得
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			builder.WriteString(fmt.Sprintf("ファイル名: %s\n内容:\n%s\n\n", path, string(content)))
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

// SaveConversationHistoryは会話履歴をファイルに保存します
func SaveConversationHistory(filename string, history []openai.ChatCompletionMessage) error {
	data, err := json.Marshal(history)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadConversationHistoryはファイルから会話履歴を読み込みます
func LoadConversationHistory(filename string) ([]openai.ChatCompletionMessage, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var history []openai.ChatCompletionMessage
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, err
	}
	return history, nil
}

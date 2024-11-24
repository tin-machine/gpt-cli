package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Options はコマンドライン引数から取得するオプションを保持します
type Options struct {
	PromptOption         string
	SystemMessage        string
	UserMessage          string
	ImageList            string
	ConfigPath           string
	Model                string
	Debug                bool
	ShowVersion          bool
	CollectFiles         bool
	HistoryFile          string
	ListFiles            bool
	Timeout              int
	FileList             string
	ShowHistory          string
	VectorStoreAction    string
	VectorStoreName      string
	VectorStoreID        string
	FileID               string
	FileIDsStr           string
	FileIDs              []string
	UploadFilePath       string
	UploadPurpose        string
	DeleteFileID         string
	UploadAndAddFilesStr string
	UploadAndAddFiles    []string
	CreateAssistant      bool
	AssistantID          string
	Message              string
	AssistantName        string
	AssistantDescription string
	Instruction          string
	FilePath             string
	ToolConfigPath       string
	Temperature          float64
	MaxTokens            *int
	Metadata             map[string]interface{}
	Attachments          []string
	Tools                []string
	Args                 []string
}

// ParseCommandLineArgs はコマンドライン引数を解析します
func ParseCommandLineArgs() (Options, error) {
	var options Options

	flag.StringVar(&options.PromptOption, "p", "", "config.yamlにあるプロンプトを選択")
	flag.StringVar(&options.SystemMessage, "s", "", "Systemのメッセージを変更")
	flag.StringVar(&options.UserMessage, "u", "", "Userのメッセージを変更")
	flag.StringVar(&options.ImageList, "i", "", "画像ファイルをカンマ区切りで")
	flag.StringVar(&options.ConfigPath, "c", "", "設定ファイルのパスを指定")
	flag.StringVar(&options.Model, "model", "gpt-4o-mini", "使用するモデルを指定")
	flag.BoolVar(&options.Debug, "d", false, "デバッグモードを有効にする")
	flag.BoolVar(&options.ShowVersion, "version", false, "バージョン情報を表示")
	flag.BoolVar(&options.CollectFiles, "collect", false, "現在のディレクトリ内のファイルをUserメッセージに追加")
	flag.StringVar(&options.HistoryFile, "history", "", "会話履歴の保存ファイルを指定（拡張子は不要）")
	flag.IntVar(&options.Timeout, "t", 60, "タイムアウト時間（秒）を指定")
	flag.StringVar(&options.FileList, "f", "", "読み込むファイルのパスをカンマ区切りで指定")
	flag.StringVar(&options.ShowHistory, "show-history", "", "会話履歴を表示")
	flag.StringVar(&options.VectorStoreName, "vector-store-name", "", "作成するベクトルストアの名前を指定")
	flag.StringVar(&options.VectorStoreAction, "vector-store-action", "", "ベクトルストアのアクションを指定（create, list, delete, add-file）")
	flag.StringVar(&options.VectorStoreID, "vector-store-id", "", "操作するベクトルストアのIDを指定")
	flag.StringVar(&options.ToolConfigPath, "tool-config", "", "ツールの設定ファイルのパスを指定")
	flag.StringVar(&options.FileID, "file-id", "", "ベクトルストアに追加するファイルのIDを指定")
	flag.StringVar(&options.FileIDsStr, "file-ids", "", "ベクトルストアに追加するファイルのIDをカンマ区切りで指定")
	flag.StringVar(&options.UploadFilePath, "upload-file", "", "OpenAIにアップロードするファイルのパスを指定")
	flag.StringVar(&options.UploadPurpose, "upload-purpose", "fine-tune", "ファイルのアップロード目的を指定（例: fine-tune, answers）")
	flag.BoolVar(&options.ListFiles, "list-files", false, "アップロードしたファイルの一覧を表示")
	flag.StringVar(&options.DeleteFileID, "delete-file", "", "削除するファイルのIDを指定")
	flag.StringVar(&options.UploadAndAddFilesStr, "upload-and-add-to-vector", "", "アップロードするファイルのパスをカンマ区切りで指定し、自動的にベクトルストアに追加")
	flag.StringVar(&options.AssistantID, "assistant-id", "", "操作するアシスタントのIDを指定")
	flag.StringVar(&options.AssistantName, "assistant-name", "MyAssistant", "アシスタントの名前を指定")
	flag.StringVar(&options.AssistantDescription, "assistant-description", "これはアシスタントの説明です。", "アシスタントの説明を指定")
	flag.StringVar(&options.Instruction, "instruction", "あなたはユーザーを助けるアシスタントです。", "アシスタントへの指示を指定")
	flag.StringVar(&options.FilePath, "file-path", "", "アップロードするファイルのパスを指定")
	flag.StringVar(&options.UserMessage, "user-message", "", "ユーザーからのメッセージを指定")
	flag.Float64Var(&options.Temperature, "temperature", 0.7, "モデルの温度パラメータを指定")
	flag.BoolVar(&options.CreateAssistant, "create-assistant", false, "新しいアシスタントを作成する")
	flag.StringVar(&options.Message, "message", "", "アシスタントに送信するメッセージを指定")
	// flag.IntVar(&options.MaxTokens, "max-tokens", 16384, "Max tokens to generate in the completion")

	// MaxTokensのフラグを設定
	options.MaxTokens = nil
	flag.Func("max-tokens", "Max tokens to generate in the completion", func(s string) error {
		value, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		options.MaxTokens = &value
		return nil
	})

	flag.Parse()

	options.Args = flag.Args()

	// アップロードするファイルのリストをパース
	if options.UploadAndAddFilesStr != "" {
		options.UploadAndAddFiles = strings.Split(options.UploadAndAddFilesStr, ",")
		for i := range options.UploadAndAddFiles {
			options.UploadAndAddFiles[i] = strings.TrimSpace(options.UploadAndAddFiles[i])
		}
	}

	// FileIDsStrを分割してFileIDsに設定
	if options.FileIDsStr != "" {
		options.FileIDs = strings.Split(options.FileIDsStr, ",")
		for i := range options.FileIDs {
			options.FileIDs[i] = strings.TrimSpace(options.FileIDs[i])
		}
	}

	// // コマンドライン引数から取得した max-tokens 値を Options 構造体に設定
	// if options.MaxTokens <= 0 {
	// 	options.MaxTokens = 200 // デフォルト値などを適切に設定
	// }

	return options, nil
}

// SetupLogging はロギングの設定を行います
func SetupLogging(debug bool) {
	if debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(0)
		log.SetOutput(os.Stderr)
	}
}

// BuildUserMessage はユーザーメッセージを構築します
func BuildUserMessage(options *Options, input io.Reader) error {
	// フラグ以外の引数をユーザーメッセージに追加
	if len(options.Args) > 0 {
		options.UserMessage += " " + strings.Join(options.Args, " ")
	}

	// 標準入力からのデータを取得
	if inputAvailable() {
		reader := bufio.NewReader(input)
		inputData, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("標準入力の読み込みに失敗しました: %w", err)
		}
		trimmedInput := strings.TrimSpace(string(inputData))
		if trimmedInput != "" {
			options.UserMessage += " " + trimmedInput
		}
	}
	return nil
}

// ConfigureLogDirectory はログディレクトリの設定と検証を行います
func ConfigureLogDirectory(options *Options, config Config) error {
	// デフォルトのログディレクトリを設定
	logDir := GetLogDirectory(config)
	if err := EnsureDirectory(logDir); err != nil {
		return fmt.Errorf("ログディレクトリの作成に失敗しました: %w", err)
	}

	// historyFileのフルパスをLogDirに基づいて設定
	if options.HistoryFile != "" {
		options.HistoryFile = filepath.Join(logDir, options.HistoryFile)
	}

	// ログファイル名を自動生成
	if options.HistoryFile == "" && config.AutoSaveLogs {
		options.HistoryFile = filepath.Join(logDir, fmt.Sprintf("log_%s.json", time.Now().Format("20060102_150405.000")))
	}

	// showHistoryのフルパスをLogDirに基づいて設定
	if options.ShowHistory != "" {
		options.ShowHistory = filepath.Join(logDir, options.ShowHistory)
	}

	return nil
}

func (o Options) String() string {
	var sb strings.Builder
	sb.WriteString("Options:\n")
	sb.WriteString(fmt.Sprintf("  PromptOption: %s\n", o.PromptOption))
	sb.WriteString(fmt.Sprintf("  SystemMessage: %s\n", o.SystemMessage))
	sb.WriteString(fmt.Sprintf("  UserMessage: %s\n", o.UserMessage))
	sb.WriteString(fmt.Sprintf("  Model: %s\n", o.Model))
	sb.WriteString(fmt.Sprintf("  Debug: %t\n", o.Debug))
	sb.WriteString(fmt.Sprintf("  AssistantID: %s\n", o.AssistantID))
	sb.WriteString(fmt.Sprintf("  Temperature: %f\n", o.Temperature))
	// sb.WriteString(fmt.Sprintf("  MaxTokens: %d\n", o.MaxTokens))
	sb.WriteString(fmt.Sprintf("  ImageList: %s\n", o.ImageList))
	sb.WriteString(fmt.Sprintf("	ConfigPath: %s\n", o.ConfigPath))
	sb.WriteString(fmt.Sprintf("	ShowVersion: %t\n", o.ShowVersion))
	sb.WriteString(fmt.Sprintf("	HistoryFile: %s\n", o.HistoryFile))
	sb.WriteString(fmt.Sprintf("	ListFiles: %t\n", o.ListFiles))
	sb.WriteString(fmt.Sprintf("	Timeout: %d\n", o.Timeout))
	sb.WriteString(fmt.Sprintf("	FileList: %s\n", o.FileList))
	sb.WriteString(fmt.Sprintf("	ShowHistory: %s\n", o.ShowHistory))
	sb.WriteString(fmt.Sprintf("	VectorStoreAction: %s\n", o.VectorStoreAction))
	sb.WriteString(fmt.Sprintf("	VectorStoreName: %s\n", o.VectorStoreName))
	sb.WriteString(fmt.Sprintf("	VectorStoreID: %s\n", o.VectorStoreID))
	sb.WriteString(fmt.Sprintf("	FileID: %s\n", o.FileID))
	sb.WriteString(fmt.Sprintf("	FileIDsStr: %s\n", o.FileIDsStr))
	sb.WriteString(fmt.Sprintf("  FileIDs: %s\n", strings.Join(o.FileIDs, ", ")))
	sb.WriteString(fmt.Sprintf("	UploadFilePath: %s\n", o.UploadFilePath))
	sb.WriteString(fmt.Sprintf("	UploadPurpose: %s\n", o.UploadPurpose))
	sb.WriteString(fmt.Sprintf("	DeleteFileID: %s\n", o.DeleteFileID))
	sb.WriteString(fmt.Sprintf("	UploadAndAddFilesStr: %s\n", o.UploadAndAddFilesStr))
	sb.WriteString(fmt.Sprintf("	UploadAndAddFiles: %s\n", strings.Join(o.UploadAndAddFiles, ", ")))
	sb.WriteString(fmt.Sprintf("	CreateAssistant: %t\n", o.CreateAssistant))
	sb.WriteString(fmt.Sprintf("	Message: %s\n", o.Message))
	sb.WriteString(fmt.Sprintf("	AssistantName: %s\n", o.AssistantName))
	sb.WriteString(fmt.Sprintf("	AssistantDescription: %s\n", o.AssistantDescription))
	sb.WriteString(fmt.Sprintf("	Instruction: %s\n", o.Instruction))
	sb.WriteString(fmt.Sprintf("	FilePath: %s\n", o.FilePath))
	sb.WriteString(fmt.Sprintf("	ToolConfigPath: %s\n", o.ToolConfigPath))
	sb.WriteString(fmt.Sprintf("	Attachments: %s\n", o.Attachments))
	sb.WriteString(fmt.Sprintf("	Tools: %s\n", strings.Join(o.Tools, ", ")))
	sb.WriteString(fmt.Sprintf("	Args: %s\n", strings.Join(o.Args, ", ")))
	if o.MaxTokens != nil {
		sb.WriteString(fmt.Sprintf("  MaxTokens: %d\n", *o.MaxTokens))
	} else {
		sb.WriteString("  MaxTokens: <nil>\n")
	}

	return sb.String()
}

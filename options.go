// 新しいファイル `options.go` を追加し、コマンドラインオプションの解析と関連する処理をまとめました。

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options はコマンドライン引数から取得するオプションを保持します
type Options struct {
	PromptOption  string
	SystemMessage string
	UserMessage   string
	ImageList     string
	ConfigPath    string
	Model         string
	Debug         bool
	ShowVersion   bool
	CollectFiles  bool
	HistoryFile   string
	Timeout       int
	FileList      string
	ShowHistory   string
	Args          []string
}

// ParseCommandLineArgs はコマンドライン引数を解析します
func ParseCommandLineArgs() (Options, error) {
	var options Options

	flag.StringVar(&options.PromptOption, "p", "", "config.yamlにあるプロンプトを選択")
	flag.StringVar(&options.SystemMessage, "s", "", "Systemのメッセージを変更")
	flag.StringVar(&options.UserMessage, "u", "", "Userのメッセージを変更")
	flag.StringVar(&options.ImageList, "i", "", "画像ファイルをカンマ区切りで")
	flag.StringVar(&options.ConfigPath, "c", "", "設定ファイルのパスを指定")
	flag.StringVar(&options.Model, "m", "", "使用するモデルを指定")
	flag.BoolVar(&options.Debug, "d", false, "デバッグモードを有効にする")
	flag.BoolVar(&options.ShowVersion, "version", false, "バージョン情報を表示")
	flag.BoolVar(&options.CollectFiles, "collect", false, "現在のディレクトリ内のファイルをUserメッセージに追加")
	flag.StringVar(&options.HistoryFile, "history", "", "会話履歴の保存ファイルを指定（拡張子は不要）")
	flag.IntVar(&options.Timeout, "t", 60, "タイムアウト時間（秒）を指定")
	flag.StringVar(&options.FileList, "f", "", "読み込むファイルのパスをカンマ区切りで指定")
	flag.StringVar(&options.ShowHistory, "show-history", "", "会話履歴を表示")

	flag.Parse()

	options.Args = flag.Args()

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
func BuildUserMessage(options *Options) error {
	// フラグ以外の引数をユーザーメッセージに追加
	if len(options.Args) > 0 {
		options.UserMessage += " " + strings.Join(options.Args, " ")
	}

	// 標準入力からのデータを取得
	if inputAvailable() {
		reader := bufio.NewReader(os.Stdin)
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
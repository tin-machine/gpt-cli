package main

import (
	"context"
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
	promptOption := flag.String("p", "", "config.yamlにあるプロンプトを選択")
	outputFile := flag.String("o", "", "出力するファイルを指定")
	systemMessage := flag.String("s", "", "Systemのメッセージを変更")
	userMessage := flag.String("u", "", "Userのメッセージを変更")
	imageList := flag.String("i", "", "画像ファイルをカンマ区切りで")
	configPath := flag.String("c", "", "設定ファイルのパスを指定")
	debug := flag.Bool("d", false, "デバッグモードを有効にする")
	showVersion := flag.Bool("version", false, "バージョン情報を表示")
	flag.Parse()

	// バージョンを表示
	if *showVersion {
		fmt.Printf("Version: %s\n", Version)
		return
	}

	// -dオプションの時に出力されるデバック用出力
	debugPrintf := func(format string, args ...interface{}) {
		if *debug {
			log.Printf(format, args...)
		}
	}

	log.SetOutput(os.Stderr)
	debugPrintf("Version: %s\n", Version)

	defaultConfigPath := filepath.Join(os.Getenv("HOME"), ".config", "gpt-cli", "config.yaml")
	if *configPath == "" {
		*configPath = defaultConfigPath
	}

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("設定ファイルが読み込めません main.go (%s): %v", *configPath, err)
	}

	// config, err := LoadConfig("config.yaml")
	// if err != nil {
	// 	log.Fatalf("Failed to read config.yaml: %v", err)
	// }

	debugPrintf("Config: %v\n", config)

	var promptConfig Prompt
	if *promptOption != "" {
		var ok bool
		promptConfig, ok = config.Prompts[*promptOption]
		if !ok {
			log.Fatalf("Prompt option %s is not defined in the config file", *promptOption)
		}
	} else if *systemMessage == "" && *userMessage == "" {
		log.Fatalf("-p(プロンプトの指定)が無い場合は-s(システムプロンプト)か-u(ユーザープロンプト)の指定が必要です")
	}

	// Systemのメッセージをコマンドラインのものに修正
	if *systemMessage != "" {
		promptConfig.System = *systemMessage
	}

	// Userのメッセージをコマンドラインのものに修正
	if *userMessage != "" {
		promptConfig.User = *userMessage
	}

	// カンマで文字列を分割
	if *imageList != "" {
		promptConfig.Attachments = strings.Split(*imageList, ",")
	}

	debugPrintf("Prompt config: %v\n", promptConfig)

	messages, err := CreateMessages(promptConfig)
	if err != nil {
		log.Fatalf("Failed to create messages: %v", err)
	}

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	openaiConfig := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	openaiConfig.HTTPClient = httpClient
	client := openai.NewClientWithConfig(openaiConfig)

	// https://github.com/sashabaranov/go-openai/blob/03851d20327b7df5358ff9fb0ac96f476be1875a/completion.go#L25
	// デフォルトのモデルは gpt-4o とする
	if promptConfig.Model == "" {
		promptConfig.Model = "gpt-4o"
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    promptConfig.Model,
			Messages: messages,
		},
	)

	if err != nil {
		log.Fatalf("ChatCompletion error: %v", err)
	}

	debugPrintf("Response: %v\n", resp.Choices[0].Message.Content)

	var outputFileName string
	if *outputFile != "" {
		outputFileName = *outputFile
	} else {
		dirName := fmt.Sprintf("%v", time.Now().Unix())
		err = os.Mkdir(dirName, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory: %v\n", err)
		}
		outputFileName = fmt.Sprintf("%s/conversation.txt", dirName)
	}

	err = SaveConversation(outputFileName, resp.Choices[0].Message.Content)
	if err != nil {
		log.Fatalf("Failed to write conversation to file: %v\n", err)
	}
}

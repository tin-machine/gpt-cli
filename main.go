package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

var Version string

func main() {
	promptOption := flag.String("p", "", "config.yamlにあるプロンプトを選択")
	outputFile := flag.String("o", "", "出力するファイルを指定")
	debug := flag.Bool("d", false, "デバッグモードを有効にする")
	showVersion := flag.Bool("version", false, "バージョン情報を表示")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Version: %s\n", Version)
		return
	}

	debugPrintf := func(format string, args ...interface{}) {
		if *debug {
			log.Printf(format, args...)
		}
	}

	log.SetOutput(os.Stderr)
	debugPrintf("Version: %s\n", Version)

	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config.yaml: %v", err)
	}

	debugPrintf("Config: %v\n", config)

	promptConfig, ok := config.Prompts[*promptOption]
	if !ok {
		log.Fatalf("Prompt option %s is not defined in the config file", *promptOption)
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

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4o,
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

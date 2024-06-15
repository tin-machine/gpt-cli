package main

import (
	"context"
	"flag"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

// PromptMapping is a struct to hold the yaml configuration
type PromptMapping struct {
	Prompts map[string]string `yaml:"prompts"`
}

func main() {
	// コマンドラインオプションの設定
	promptOption := flag.String("p", "", "config.yamlにあるプロンプトを選択")
	addMessageFile := flag.String("m", "", "追加するメッセージをファイルで指定")
	outputFile := flag.String("o", "", "出力するファイルを指定")
	debug := flag.Bool("d", false, "デバッグモードを有効にする")
	flag.Parse()

	// デバッグメッセージの関数
	debugPrintf := func(format string, args ...interface{}) {
		if *debug {
			fmt.Printf(format, args...)
		}
	}

	debugPrintf("addMessageFile: %v\n", *addMessageFile)

	// config.yamlの読み込み
	yamlFile, err := ioutil.ReadFile("config.yaml")
	debugPrintf("config.yaml content:\n%s\n", string(yamlFile))

	if err != nil {
		fmt.Printf("yamlFile.Get err #%v ", err)
	}

	// yamlのパース
	var promptMapping PromptMapping
	err = yaml.Unmarshal(yamlFile, &promptMapping)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	debugPrintf("Prompt map: %v\n", promptMapping)

	// ベースとなるpromptの読み込み
	prompt, ok := promptMapping.Prompts[*promptOption]
	if !ok {
		fmt.Printf("Prompt option %s is not defined in the config file", *promptOption)
		return
	}

	addMessage, err := ioutil.ReadFile(*addMessageFile)
	if err != nil {
		fmt.Printf("Failed to read additional message file: %v\n", err)
		return
	}
	prompt = prompt + string(addMessage)

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4, // Change this to GPT4 as per your requirement
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	// 結果のプリント
	debugPrintf("Prompt map: %v\n", resp.Choices[0].Message.Content)

	// 保存ファイル名の指定がある場合、それを使用
	var outputFileName string
	if *outputFile != "" {
		outputFileName = *outputFile
	} else {
		// 指定がない場合はunix timeでディレクトリを作成
		dirName := fmt.Sprintf("%v", time.Now().Unix())
		err = os.Mkdir(dirName, 0755)
		if err != nil {
			fmt.Printf("Failed to create directory: %v\n", err)
			return
		}
		outputFileName = fmt.Sprintf("%s/conversation.txt", dirName)
	}

	// 指定されたファイル名またはデフォルトのファイル名で会話を保存
	err = ioutil.WriteFile(outputFileName, []byte(resp.Choices[0].Message.Content), 0644)
	if err != nil {
		fmt.Printf("Failed to write conversation to file: %v\n", err)
		return
	}
}

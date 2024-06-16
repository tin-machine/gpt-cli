package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v2"
)

// PromptMapping is a struct to hold the yaml configuration
type PromptMapping struct {
	Prompts map[string]string `yaml:"prompts"`
}

// Function to convert image to base64
func imageToBase64(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size = fileInfo.Size()
	buf := make([]byte, size)
	file.Read(buf)

	return base64.StdEncoding.EncodeToString(buf), nil
}

func main() {
	// コマンドラインオプションの設定
	promptOption := flag.String("p", "", "config.yamlにあるプロンプトを選択")
	addMessageFile := flag.String("m", "", "追加するメッセージをファイルで指定")
	outputFile := flag.String("o", "", "出力するファイルを指定")
	imageFiles := flag.String("images", "", "カンマ区切りの画像ファイルのリスト")
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

	// 画像ファイルをbase64に変換
	imageList := []string{}
	if *imageFiles != "" {
		imagePaths := strings.Split(*imageFiles, ",")
		for _, imagePath := range imagePaths {
			base64Image, err := imageToBase64(imagePath)
			if err != nil {
				fmt.Printf("Failed to convert image to base64: %v\n", err)
				return
			}
			imageList = append(imageList, base64Image)
		}
	}

	// メッセージの作成
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	for _, base64Image := range imageList {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("![image](data:image/png;base64,%s)", base64Image),
		})
	}

	// カスタムHTTPクライアントの作成
	httpClient := &http.Client{
		Timeout: 60 * time.Second, // タイムアウトを60秒に設定
	}

	// タイムアウトを伸ばしたHTTPクライアントを使うため
	// NewClientWithConfig(config)の形で利用する
	// https://pkg.go.dev/github.com/sashabaranov/go-openai#ClientConfig
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.HTTPClient = &http.Client{
		HTTPClient: httpClient,
	}
	//	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	// OpenAIクライアントの作成
	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4o, // Change this to GPT4 as per your requirement
			Messages: messages,
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

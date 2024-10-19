package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// SplitImageList は画像ファイルのリストを分割します
func SplitImageList(imageList string) []string {
	return strings.Split(imageList, ",")
}

// CollectFiles は指定したディレクトリ内のファイル名と内容を収集します（.gitディレクトリを除く）
func CollectFiles(dir string) (string, error) {
	var builder strings.Builder

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
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

// CreateMessages はプロンプト設定からメッセージを作成します
func CreateMessages(promptConfig Prompt) ([]openai.ChatCompletionMessage, error) {
	var messages []openai.ChatCompletionMessage

	// システムメッセージの追加
	if promptConfig.System != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: promptConfig.System,
		})
	}

	// ユーザーメッセージの追加
	if promptConfig.User != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: promptConfig.User,
		})
	}

	// 添付ファイル（画像）の処理
	for _, attachmentPath := range promptConfig.Attachments {
		base64Image, mimeType, err := imageToBase64(attachmentPath)
		if err != nil {
			return nil, fmt.Errorf("画像のエンコードに失敗しました: %v", err)
		}

		// 画像データをメッセージに追加
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("画像ファイル: %s\nデータ: data:%s;base64,%s", attachmentPath, mimeType, base64Image),
		})
	}

	return messages, nil
}

// imageToBase64 は画像ファイルをBase64エンコードします
func imageToBase64(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	// ファイル拡張子からMIMEタイプを推測
	ext := strings.ToLower(filepath.Ext(path))
	mimeType := ""
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	default:
		return "", "", fmt.Errorf("サポートされていない画像形式: %s", ext)
	}

	base64Image := base64.StdEncoding.EncodeToString(data)
	return base64Image, mimeType, nil
}

// SaveConversationHistory は会話履歴をファイルに保存します
func SaveConversationHistory(filename string, history []openai.ChatCompletionMessage) error {
	if filepath.Ext(filename) == "" {
		filename += ".json"
	}
	data, err := json.Marshal(history)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0600)
}

// LoadConversationHistory はファイルから会話履歴を読み込みます
func LoadConversationHistory(filename string) ([]openai.ChatCompletionMessage, error) {
	if filename == "" {
		return []openai.ChatCompletionMessage{}, nil
	}
	if filepath.Ext(filename) == "" {
		filename += ".json"
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []openai.ChatCompletionMessage{}, nil
		}
		return nil, err
	}
	var history []openai.ChatCompletionMessage
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, err
	}
	return history, nil
}

// SaveOutput はアシスタントの応答をファイルに保存します
func SaveOutput(outputFile, content string) error {
	var outputFileName string
	if outputFile != "" {
		outputFileName = outputFile
	} else {
		dirName := fmt.Sprintf("%v", time.Now().Unix())
		err := os.Mkdir(dirName, 0700)
		if err != nil {
			return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
		}
		outputFileName = filepath.Join(dirName, "conversation.txt")
	}

	return os.WriteFile(outputFileName, []byte(content), 0600)
}

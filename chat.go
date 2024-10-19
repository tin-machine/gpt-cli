package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

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

// 画像ファイルをBase64エンコードする関数
func imageToBase64(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	// 画像のデコード
	img, format, err := image.Decode(file)
	if err != nil {
		return "", "", err
	}

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	case "png":
		err = png.Encode(&buf, img)
	default:
		return "", "", fmt.Errorf("サポートされていない画像形式: %s", format)
	}
	if err != nil {
		return "", "", err
	}

	// Base64エンコード
	base64Image := base64.StdEncoding.EncodeToString(buf.Bytes())
	mimeType := "image/" + format

	return base64Image, mimeType, nil
}

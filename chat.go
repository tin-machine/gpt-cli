package main

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func CreateMessages(promptConfig Prompt) ([]openai.ChatCompletionMessage, error) {
	// messages := []openai.ChatCompletionMessage{
	// 	{
	// 		Role:    openai.ChatMessageRoleSystem,
	// 		Content: promptConfig.System,
	// 	},
	// 	{
	// 		Role:    openai.ChatMessageRoleUser,
	// 		Content: promptConfig.User,
	// 	},
	// }

	// この次に-mオプションがあった場合は追加のメッセージを追加する
	// promptConfig.System
	// promptConfig.User
	// 上記が存在したら、配列に追加する

	var messages []openai.ChatCompletionMessage

  if promptConfig.System != "" {
    messages = append(messages, openai.ChatCompletionMessage{
      Role:    "system",
        Content: promptConfig.System,
    })
  }

  if promptConfig.User != "" {
    messages = append(messages, openai.ChatCompletionMessage{
      Role:    "user",
      Content: promptConfig.User,
    })
  }

	// 添付ファイルの処理
	for _, attachmentPath := range promptConfig.Attachments {
		base64Image, mimeType, err := imageToBase64(attachmentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to convert image to base64: %v", err)
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL:    fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image),
						Detail: openai.ImageURLDetailLow,
					},
				},
			},
		})
	}

	return messages, nil
}

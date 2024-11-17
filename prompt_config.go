package main

import (
	"fmt"
)

// GetPromptConfig はプロンプトの設定を取得します
func GetPromptConfig(config Config, options Options) (Prompt, error) {
	var promptConfig Prompt

	if options.PromptOption != "" {
		var ok bool
		promptConfig, ok = config.Prompts[options.PromptOption]
		if !ok {
			return promptConfig, fmt.Errorf("プロンプトオプション %s は設定ファイルに定義されていません", options.PromptOption)
		}
	}

	// コマンドライン引数からの上書き
	if options.SystemMessage != "" {
		promptConfig.System = options.SystemMessage
	}
	if options.UserMessage != "" {
		promptConfig.User = options.UserMessage
	}
	if options.Model != "" {
		promptConfig.Model = options.Model
	}
	if len(options.Attachments) > 0 {
		promptConfig.Attachments = options.Attachments
	}

	// デフォルトのモデル設定
	if promptConfig.Model == "" {
		promptConfig.Model = defaultModel
	}

	// 画像リストの処理
	if options.ImageList != "" {
		promptConfig.Attachments = SplitImageList(options.ImageList)
	}

	// -collect オプションが指定された場合、ファイルを収集
	if options.CollectFiles {
		filesContent, err := CollectFiles(".")
		if err != nil {
			return promptConfig, fmt.Errorf("ファイルの収集に失敗しました: %w", err)
		}
		promptConfig.User += "\n\n" + filesContent
	}

	// -f オプションが指定された場合、ファイルを読み込む
	if options.FileList != "" {
		filesContent, err := ReadFiles(options.FileList)
		if err != nil {
			return promptConfig, fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
		}
		promptConfig.User += "\n\n" + filesContent
	}

	// ツール設定のマージ
	if len(options.Tools) > 0 {
		promptConfig.Tools = append(promptConfig.Tools, options.Tools...)
	}

	return promptConfig, nil
}

// GetDefaultPromptConfig はデフォルトのプロンプトの設定を取得します
func GetDefaultPromptConfig() Prompt {
	return Prompt{
		Model:       "gpt-3.5-turbo",
		System:      "あなたはユーザーを助けるアシスタントです。",
		User:        "ユーザーからのメッセージがまだありません。",
		MaxTokens:   150,
		Attachments: []string{},
		Tools:       []string{},
	}
}

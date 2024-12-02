package main

import (
	"fmt"
)

const (
	DefaultVectorStoreName = "Default Vector Store"
	DefaultVectorStoreID   = "default-vector-store-id"
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

	// ベクトルストア設定を取得
	if options.VectorStoreAction != "" {
		if vectorStoreConfig, exist := config.VectorStores[options.VectorStoreAction]; exist {
			options.VectorStoreName = vectorStoreConfig.Name
			options.VectorStoreID = vectorStoreConfig.ID
		} else {
			// 設定が見つからない場合はエラーを出すか、デフォルト値を使用
			options.VectorStoreName = DefaultVectorStoreName
			options.VectorStoreID = DefaultVectorStoreID
			logger.Info("ベクトルストア設定が見つからなかったため、デフォルト設定を使用します。")
		}
	}

	// コマンドライン引数からの上書き
	if options.MaxTokens != nil {
		promptConfig.MaxTokens = options.MaxTokens
	}
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
		Attachments: []string{},
		Tools:       []string{},
	}
}

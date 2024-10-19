package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Prompt struct {
	Model       string   `yaml:"model"`
	System      string   `yaml:"system"`
	User        string   `yaml:"user"`
	Attachments []string `yaml:"attachments"`
}

type Config struct {
	Prompts map[string]Prompt `yaml:"prompts"`
}

// LoadConfig は指定されたファイルパスから設定を読み込みます
func LoadConfig(filePath string) (Config, error) {
	var config Config

	// ファイルパスを正規化
	cleanPath := filepath.Clean(filePath)

	yamlFile, err := os.ReadFile(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, fmt.Errorf("設定ファイルが見つかりませんでした (%s): %w", cleanPath, err)
		}
		return config, fmt.Errorf("設定ファイルの読み込みに失敗しました (%s): %w", cleanPath, err)
	}

	// YAMLファイルをパース
	err = yaml.UnmarshalStrict(yamlFile, &config)
	if err != nil {
		return config, fmt.Errorf("設定ファイルの解析に失敗しました (%s): %w", cleanPath, err)
	}

	return config, nil
}

// GetConfigFilePath は設定ファイルのパスを取得します
func GetConfigFilePath(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}

	// ユーザーの設定ディレクトリを取得
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("ユーザー設定ディレクトリの取得に失敗しました: %w", err)
	}
	defaultConfigPath := filepath.Join(configDir, "gpt-cli", "config.yaml")
	return defaultConfigPath, nil
}

// GetPromptConfig はプロンプトの設定を取得します
func GetPromptConfig(config Config, promptOption, systemMessage, userMessage, model string) (Prompt, error) {
	var promptConfig Prompt

	if promptOption != "" {
		var ok bool
		promptConfig, ok = config.Prompts[promptOption]
		if !ok {
			return promptConfig, fmt.Errorf("プロンプトオプション %s は設定ファイルに定義されていません", promptOption)
		}
	} else if systemMessage == "" && userMessage == "" {
		return promptConfig, fmt.Errorf("-p(プロンプトの指定)が無い場合は-s(システムメッセージ)か-u(ユーザーメッセージ)の指定が必要です")
	}

	// コマンドライン引数からの上書き
	if systemMessage != "" {
		promptConfig.System = systemMessage
	}
	if userMessage != "" {
		promptConfig.User = userMessage
	}
	if model != "" {
		promptConfig.Model = model
	}

	// デフォルトのモデル設定
	if promptConfig.Model == "" {
		promptConfig.Model = defaultModel
	}

	return promptConfig, nil
}

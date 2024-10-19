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

// SaveConversation は指定されたファイルパスに会話内容を保存します
func SaveConversation(filePath string, content string) error {
	// ファイルパスを正規化
	cleanPath := filepath.Clean(filePath)

	err := os.WriteFile(cleanPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("会話内容の保存に失敗しました (%s): %w", cleanPath, err)
	}

	return nil
}

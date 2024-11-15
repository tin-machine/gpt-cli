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
	MaxTokens   int      `yaml:"maxTokens"`
	Attachments []string `yaml:"attachments"`
	Tools       []string `yaml:"tools"`
}

type Config struct {
	LogDir       string            `yaml:"logDir"`
	AutoSaveLogs bool              `yaml:"autoSaveLogs"`
	Prompts      map[string]Prompt `yaml:"prompts"`
}

// LoadConfig は指定されたファイルパスから設定を読み込みます
func LoadConfig(filePath string) (Config, error) {
	var config Config

	if filePath == "" {
		return config, fmt.Errorf("設定ファイルのパスが指定されていません")
	}

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

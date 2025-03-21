package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Promptはプロンプト設定の構造体で、以下のフィールドを含みます:
// - Model: 使用するAIモデル名
// - System: システムメッセージ（アシスタントへの指示）
// - User: ユーザーからのメッセージ
// - MaxTokens: 最大トークン数
// - Attachments: 添付ファイル名のリスト
// - Tools: 使用するツール名のリスト
type Prompt struct {
	Model       string   `yaml:"model"`
	System      string   `yaml:"system"`
	User        string   `yaml:"user"`
	MaxTokens   *int     `yaml:"maxTokens"`
	Attachments []string `yaml:"attachments"`
	Tools       []string `yaml:"tools"`
}

type VectorStoreConfig struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type AssistantConfig struct {
	Name            string  `yaml:"name"`
	Description     string  `yaml:"description"`
	Model           string  `yaml:"model"`
	Instruction     string  `yaml:"instruction"`
	Temperature     float64 `yaml:"temperature"`
	VectorStoreName string  `yaml:"vectorStoreName"`
}

// Configはアプリケーション全体の設定を保持するための構造体で、以下のフィールドを含みます:
// - LogDir: ログファイルを保存するディレクトリ
// - AutoSaveLogs: ログの自動保存を有効にするかどうか
// - Prompts: プロンプト名とその内容のマッピング
type Config struct {
	LogDir       string                       `yaml:"logDir"`
	AutoSaveLogs bool                         `yaml:"autoSaveLogs"`
	Prompts      map[string]Prompt            `yaml:"prompts"`
	VectorStores map[string]VectorStoreConfig `yaml:"vectorStores"`
	Assistants   map[string]AssistantConfig   `yaml:"assistants"`
}

// LoadConfigは、指定されたファイルパスから設定を読み込む関数です。
// 設定ファイルがYAML形式であり、内容がConfig構造体にマッピングされます。
// 引数filePathは設定ファイルの場所を指します。
// 成功すると、読み込まれたConfigが返され、読み込みに失敗した場合はエラーメッセージが返されます。
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
	decoder := yaml.NewDecoder(bytes.NewReader(yamlFile))
	decoder.KnownFields(true)
	err = decoder.Decode(&config)
	if err != nil {
		return config, fmt.Errorf("設定ファイルの解析に失敗しました (%s): %w", cleanPath, err)
	}

	return config, nil
}

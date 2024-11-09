package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// LoadConfiguration は設定ファイルを読み込みます
func LoadConfiguration(configPath string) (Config, error) {
	// デフォルト設定の定義
	defaultConfig := Config{
		LogDir:       "",
		AutoSaveLogs: false,
		Prompts:      make(map[string]Prompt),
	}

	// 設定ファイルのパスを取得
	configFilePath, err := GetConfigFilePath(configPath)
	if err != nil {
		log.Printf("設定ファイルのパスを取得できませんでした (%v)。デフォルト設定を使用します。", err)
		return defaultConfig, nil
	}

	// 設定ファイルの読み込み
	config, err := LoadConfig(configFilePath)
	if err != nil {
		log.Printf("設定ファイルが読み込めません (%v)。デフォルト設定を使用します。", err)
		return defaultConfig, nil
	}

	return config, nil
}

// GetConfigFilePath は設定ファイルのパスを取得します
func GetConfigFilePath(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}

	// 環境変数をチェック
	if envConfigPath := os.Getenv("GPT_CLI_CONFIG_PATH"); envConfigPath != "" {
		return envConfigPath, nil
	}

	// ユーザーの設定ディレクトリを取得
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("ユーザー設定ディレクトリの取得に失敗しました: %w", err)
	}
	defaultConfigPath := filepath.Join(configDir, "gpt-cli", "config.yaml")
	return defaultConfigPath, nil
}

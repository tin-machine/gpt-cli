package main

import (
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v2"
)

type ToolConfig struct {
    Tools map[string]interface{} `yaml:"tools"`
}

func LoadToolConfig(filePath string) (ToolConfig, error) {
    var config ToolConfig

    if filePath == "" {
        return config, fmt.Errorf("ツール設定ファイルのパスが指定されていません")
    }

      cleanPath := filepath.Clean(filePath)

    yamlFile, err := os.ReadFile(cleanPath)
    if err != nil {
        if os.IsNotExist(err) {
            return config, fmt.Errorf("ツール設定ファイルが見つかりませんでした (%s): %w", cleanPath, err)
        }
        return config, fmt.Errorf("ツール設定ファイルの読み込みに失敗しました (%s): %w", cleanPath, err)
    }

    err = yaml.UnmarshalStrict(yamlFile, &config)
    if err != nil {
        return config, fmt.Errorf("ツール設定ファイルの解析に失敗しました (%s): %w", cleanPath, err)
    }

    return config, nil
}

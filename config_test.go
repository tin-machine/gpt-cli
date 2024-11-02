package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	os.WriteFile("config.yaml", []byte(`
prompts:
  testPrompt:
    model: "gpt-3.5-turbo"
    system: "system message"
    user: "user message"
    attachments: []
`), 0644)
	defer os.Remove("config.yaml")

	config, err := LoadConfig("config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig() エラー: %v", err)
	}

	if _, ok := config.Prompts["testPrompt"]; !ok {
		t.Errorf("LoadConfig() は 'testPrompt' を読み取れませんでした")
	}
}

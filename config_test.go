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
    maxTokens: 150
    attachments: []
    tools: []
`), 0644)
	defer os.Remove("config.yaml")

	config, err := LoadConfig("config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig() エラー: %v", err)
	}

	if prompt, ok := config.Prompts["testPrompt"]; !ok {
		t.Errorf("LoadConfig() は 'testPrompt' を読み取れませんでした")
	} else {
		if prompt.Model != "gpt-3.5-turbo" {
			t.Errorf("LoadConfig() は model を正しく読み取れませんでした")
		}
		if prompt.System != "system message" {
			t.Errorf("LoadConfig() は system を正しく読み取れませんでした")
		}
		if prompt.User != "user message" {
			t.Errorf("LoadConfig() は user を正しく読み取れませんでした")
		}
		if prompt.MaxTokens == nil || *prompt.MaxTokens != 150 {
			t.Errorf("LoadConfig() は maxTokens を正しく読み取れませんでした")
		}
		if len(prompt.Attachments) != 0 {
			t.Errorf("LoadConfig() は attachments を正しく読み取れませんでした")
		}
		if len(prompt.Tools) != 0 {
			t.Errorf("LoadConfig() は tools を正しく読み取れませんでした")
		}
	}
}

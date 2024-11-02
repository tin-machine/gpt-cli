package main

import (
	"os"
	"testing"
)

func TestNewOpenAIClient(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "dummy_key")

	_, err := NewOpenAIClient(30)
	if err != nil {
		t.Fatalf("NewOpenAIClient() エラー: %v", err)
	}
}

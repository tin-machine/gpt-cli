package main

import (
	"bytes"
	"testing"
)

func TestBuildUserMessageWithStdin(t *testing.T) {
	// 標準入力を模擬するデータ
	inputData := "This is a test input from stdin"
	mockInput := bytes.NewBufferString(inputData) // bytes.Bufferはio.Readerを実装している

	// Optionsの準備
	options := Options{
		UserMessage: "",
	}

	// BuildUserMessageを呼び出し
	err := BuildUserMessage(&options, mockInput)
	if err != nil {
		t.Fatalf("BuildUserMessage() エラー: %v", err)
	}
	// 結果の確認
	if options.UserMessage != " "+inputData {
		t.Errorf("UserMessageが正しく構築されていません。Expected: '%s', got: '%s'", inputData, options.UserMessage)
	}
}

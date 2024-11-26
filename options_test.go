package main

import (
	"os"
	"testing"
)

func TestBuildUserMessageWithStdin(t *testing.T) {
	// 標準入力を模擬するデータ
	inputData := "This is a test input from stdin"
	// mockInput := bytes.NewBufferString(inputData) // bytes.Bufferはio.Readerを実装している

	// テスト用に一時的なReader，Writerを取得する．
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}

	// os.Stdin，os.Stdoutを取得したReader，Writerと差し替え，テスト終了時に復元する．
	osStdin, osStdout := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = r, w
	defer func() { os.Stdin, os.Stdout = osStdin, osStdout }()

	// Optionsの準備
	options := Options{
		UserMessage: "",
	}

	input := []byte(inputData)
	// 標準出力へ書き込む．
	if n, err := os.Stdout.Write(input); err != nil {
		t.Errorf("input is %v bytes, but only %v byte written", len(input), n)
		return
	}

	// BuildUserMessageを呼び出し
	msg_err := BuildUserMessage(&options)
	if msg_err != nil {
		t.Fatalf("BuildUserMessage() エラー: %v", err)
	}
	// 結果の確認
	if options.UserMessage != " "+inputData {
		t.Errorf("UserMessageが正しく構築されていません。Expected: '%s', got: '%s'", inputData, options.UserMessage)
	}
}

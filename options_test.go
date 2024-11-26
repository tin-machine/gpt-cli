package main

import (
	"os"
	"testing"
)

func TestBuildUserMessageWithStdin(t *testing.T) {
	// 標準入力を模擬するデータ
	inputData := "This is a test input from stdin"

	// stdinの差し替えの準備
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // テスト終了後に元に戻す

	// 標準入力用のパイプを作成
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	// 作成したパイプを標準入力としてセット
	os.Stdin = r

	// エラーチャネルを作成
	errChan := make(chan error, 1)

	// テストで使用するデータを、バックグラウンドで標準入力に書き込む
	go func() {
		defer w.Close()
		if _, err := w.Write([]byte(inputData)); err != nil {
			errChan <- err
		}
	}()

	// Optionsの準備
	options := Options{
		UserMessage: "",
	}

	// BuildUserMessageを呼び出し
	err = BuildUserMessage(&options)
	if err != nil {
		t.Fatalf("BuildUserMessage() エラー: %v", err)
	}

	// エラーチャネルをチェック
	select {
	case err := <-errChan:
		t.Fatalf("goroutine 内でエラーが発生しました: %v", err)
	default:
		// エラーなし
	}

	// 結果の確認
	expected := " " + inputData
	if options.UserMessage != expected {
		t.Errorf("UserMessageが正しく構築されていません。Expected: '%s', got: '%s'", inputData, options.UserMessage)
	}
}

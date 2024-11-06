// main_test.go
package main

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	// 環境変数やフラグの準備
	os.Setenv("OPENAI_API_KEY", "dummy_key")

	// コマンドライン引数をモック
	os.Args = []string{"cmd", "-version"}

	// メイン関数のテスト実行
	err := Run()
	if err != nil {
		t.Errorf("Run() でエラーが発生しました: %v", err)
	}
}

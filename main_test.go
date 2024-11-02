// main_test.go
package main

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	// 環境変数やフラグの準備
	os.Setenv("OPENAI_API_KEY", "dummy_key")
	os.Args = []string{"cmd", "-version"}

	if err := Run(); err != nil {
		t.Fatalf("Run() エラー: %v", err)
	}
}

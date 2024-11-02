package main

import (
	"os"
	"testing"
)

func TestCollectFiles(t *testing.T) {
	dir := "./test_dir"
	os.Mkdir(dir, 0755)
	defer os.RemoveAll(dir)

	err := os.WriteFile(dir+"/file.txt", []byte("test data"), 0644)
	if err != nil {
		t.Fatalf("ファイル作成エラー: %v", err)
	}

	content, err := CollectFiles(dir)
	if err != nil {
		t.Errorf("CollectFiles() エラー: %v", err)
	}

	if content == "" {
		t.Errorf("CollectFiles() は空の文字列を返しました")
	}
}

func TestReadFiles(t *testing.T) {
	err := os.WriteFile("file.txt", []byte("test data"), 0644)
	if err != nil {
		t.Fatalf("ファイル作成エラー: %v", err)
	}
	defer os.Remove("file.txt")

	content, err := ReadFiles("file.txt")
	if err != nil {
		t.Errorf("ReadFiles() エラー: %v", err)
	}

	if content == "" {
		t.Errorf("ReadFiles() は空の文字列を返しました")
	}
}

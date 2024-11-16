package main

import (
	"context"
	"os"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

// MockOpenAIClient はモック用のOpenAIクライアントです。
type TestMockOpenAIClient struct{}

func (m *TestMockOpenAIClient) CreateFile(ctx context.Context, req openai.FileRequest) (openai.File, error) {
	return openai.File{
		ID:       "mock-file-id",
		FileName: req.FileName,
		Purpose:  req.Purpose,
	}, nil
}

func (m *TestMockOpenAIClient) ListFiles(ctx context.Context) (openai.FilesList, error) {
	return openai.FilesList{
		Files: []openai.File{
			{
				ID:       "file1",
				FileName: "test1.txt",
				Status:   "uploaded",
			},
			{
				ID:       "file2",
				FileName: "test2.txt",
				Status:   "uploaded",
			},
		},
	}, nil
}

func (m *TestMockOpenAIClient) DeleteFile(ctx context.Context, fileID string) error {
	return nil
}

func TestNewOpenAIClient(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "dummy_key")

	client, err := NewOpenAIClient(30)
	if err != nil {
		t.Fatalf("NewOpenAIClient() エラー: %v", err)
	}

	// クライアントが正しく初期化されたことを確認
	if client == nil {
		t.Fatalf("NewOpenAIClient() はnilを返しました")
	}
}

func TestUploadFile(t *testing.T) {
	// モッククライアントを使用
	client := &TestMockOpenAIClient{}

	// テスト用ファイルを作成
	filePath := "test_upload.txt"
	err := os.WriteFile(filePath, []byte("test data"), 0644)
	if err != nil {
		t.Fatalf("ファイル作成エラー: %v", err)
	}
	defer os.Remove(filePath)

	uploadedFile, err := UploadFile(client, filePath, "fine-tune")
	if err != nil {
		t.Fatalf("UploadFile() エラー: %v", err)
	}

	if uploadedFile.ID != "mock-file-id" {
		t.Errorf("期待されるFile IDは 'mock-file-id' ですが、実際は '%s' です", uploadedFile.ID)
	}
}

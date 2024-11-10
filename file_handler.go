package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	openai "github.com/sashabaranov/go-openai"
)

// UploadFile はファイルをOpenAIにアップロードします
func UploadFile(client *openai.Client, filePath string, purpose string) (*openai.File, error) {
	ctx := context.Background()

	// ファイルの存在確認
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("指定されたファイルが見つかりません: %s", filePath)
	}

	// ファイル名を取得
	fileName := filepath.Base(filePath)

	// FileRequest を初期化
	req := openai.FileRequest{
		FileName: fileName,
		FilePath: filePath,
		Purpose:  purpose,
	}

	uploadedFile, err := client.CreateFile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ファイルのアップロードに失敗しました: %v", err)
	}

	return &uploadedFile, nil
}

// ListUploadedFiles はアップロードされたファイルの一覧を取得します
func ListUploadedFiles(client *openai.Client) (openai.FilesList, error) {
	ctx := context.Background()
	files, err := client.ListFiles(ctx)
	if err != nil {
		return openai.FilesList{}, err
	}
	return files, nil
}

// DeleteUploadedFile は指定されたIDのファイルを削除します
func DeleteUploadedFile(client *openai.Client, fileID string) error {
	ctx := context.Background()
	err := client.DeleteFile(ctx, fileID)
	if err != nil {
		return err
	}
	return nil
}

// UploadFiles は複数のファイルをOpenAIにアップロードし、そのファイルIDのスライスを返します
func UploadFiles(client *openai.Client, filePaths []string, purpose string) ([]string, error) {
	var fileIDs []string
	for _, filePath := range filePaths {
		// ファイルの存在確認
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("指定されたファイルが見つかりません: %s", filePath)
		}

		// ファイルをアップロード
		uploadedFile, err := UploadFile(client, filePath, purpose)
		if err != nil {
			return nil, fmt.Errorf("ファイルのアップロードに失敗しました (%s): %v", filePath, err)
		}
		fmt.Printf("ファイルがアップロードされました。File ID: %s\n", uploadedFile.ID)
		// ファイルIDを収集
		fileIDs = append(fileIDs, uploadedFile.ID)
	}
	return fileIDs, nil
}

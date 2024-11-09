package main

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

// CreateVectorStore は新しいベクトルストアを作成します
func CreateVectorStore(client *openai.Client, name string) (*openai.VectorStore, error) {
	req := openai.VectorStoreRequest{
		Name: name,
	}
	ctx := context.Background()
	vs, err := client.CreateVectorStore(ctx, req)
	if err != nil {
		return nil, err
	}
	return &vs, nil
}

// ListVectorStores は既存のベクトルストアを一覧表示します
func ListVectorStores(client *openai.Client) ([]openai.VectorStore, error) {
	ctx := context.Background()
	resp, err := client.ListVectorStores(ctx, openai.Pagination{})
	if err != nil {
		return nil, err
	}
	return resp.VectorStores, nil
}

// DeleteVectorStore は指定したIDのベクトルストアを削除します
func DeleteVectorStore(client *openai.Client, vectorStoreID string) error {
	ctx := context.Background()
	_, err := client.DeleteVectorStore(ctx, vectorStoreID)
	return err
}

// AddFileToVectorStore はファイルをベクトルストアに追加します
func AddFileToVectorStore(client *openai.Client, vectorStoreID string, fileID string) (*openai.VectorStoreFile, error) {
	req := openai.VectorStoreFileRequest{
		FileID: fileID,
	}
	ctx := context.Background()
	vsFile, err := client.CreateVectorStoreFile(ctx, vectorStoreID, req)
	if err != nil {
		return nil, err
	}
	return &vsFile, nil
}

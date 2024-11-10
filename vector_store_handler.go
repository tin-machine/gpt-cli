package main

import (
	"context"
	"fmt"

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

// // AddFilesToVectorStore は複数のファイルをベクトルストアに追加します
// func AddFilesToVectorStore(client *openai.Client, vectorStoreID string, fileIDs []string) error {
//     ctx := context.Background()
//     req := openai.VectorStoreFileBatchRequest{
//         FileIDs: fileIDs,
//     }
//     vsFileBatch, err := client.CreateVectorStoreFileBatch(ctx, vectorStoreID, req)
//     if err != nil {
//         return fmt.Errorf("ファイルバッチの作成に失敗しました: %v", err)
//     }
//     fmt.Printf("ファイルバッチが作成されました。Batch ID: %s, Status: %s\n", vsFileBatch.ID, vsFileBatch.Status)
//     return nil
// }

// AddFilesToVectorStore は複数のファイルをベクトルストアに追加します
func AddFilesToVectorStore(client *openai.Client, vectorStoreID string, fileIDs []string) error {
	for _, fileID := range fileIDs {
		_, err := AddFileToVectorStore(client, vectorStoreID, fileID)
		if err != nil {
			return fmt.Errorf("ファイルID %s の追加に失敗しました: %v", fileID, err)
		}
	}
	return nil
}

// GetOrCreateVectorStore は指定された名前のVectorStoreを取得または作成します
func GetOrCreateVectorStore(client *openai.Client, name string) (*openai.VectorStore, error) {
	// 既存のVectorStoreを一覧取得
	vsList, err := ListVectorStores(client)
	if err != nil {
		return nil, fmt.Errorf("ベクトルストアの一覧取得に失敗しました: %v", err)
	}

	// 名前が一致するVectorStoreを検索
	for _, vs := range vsList {
		if vs.Name == name {
			return &vs, nil
		}
	}

	// 見つからない場合は新規作成
	vs, err := CreateVectorStore(client, name)
	if err != nil {
		return nil, fmt.Errorf("ベクトルストアの作成に失敗しました: %v", err)
	}
	return vs, nil
}

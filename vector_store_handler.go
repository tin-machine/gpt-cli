package main

import (
	"context"
	"fmt"
	"path/filepath"

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

// ベクトルストアIDで取得する関数を追加
func GetVectorStoreByID(client *openai.Client, vectorStoreID string) (*openai.VectorStore, error) {
    ctx := context.Background()

    // OpenAIクライアントを使用してベクトルストアを取得
    vectorStore, err := client.RetrieveVectorStore(ctx, vectorStoreID)
    if err != nil {
        return nil, fmt.Errorf("ベクトルストアの取得に失敗しました（ID: %s）: %w", vectorStoreID, err)
    }

    return &vectorStore, nil
}


// GetVectorStore はベクトルストアを取得または作成します
func GetVectorStore(client *openai.Client, options Options) (*openai.VectorStore, error) {
    if options.VectorStoreID != "" {
        // IDでベクトルストアを取得
        vs, err := GetVectorStoreByID(client, options.VectorStoreID)
        if err != nil {
            return nil, err
        }
        return vs, nil
    } else if options.VectorStoreName != "" {
        // 名前でベクトルストアを取得または作成
        vs, err := GetOrCreateVectorStore(client, options.VectorStoreName)
        if err != nil {
            return nil, err
        }
        return vs, nil
    } else {
        // デフォルトのベクトルストアを使用またはエラーを返す
        return nil, fmt.Errorf("ベクトルストアのIDまたは名前を指定してください")
    }
}

func createAndAttachVectorStore(client *openai.Client, assistantID string, options Options) error {
    ctx := context.Background()

    // ベクトルストアを作成
    vectorStoreRequest := openai.VectorStoreRequest{
		    Name: options.VectorStoreName,
    }

    vectorStore, err := client.CreateVectorStore(ctx, vectorStoreRequest)
    if err != nil {
        return fmt.Errorf("ベクトルストアの作成に失敗しました: %w", err)
    }

    fmt.Printf("ベクトルストアが作成されました:\nID: %s\nName: %s\n", vectorStore.ID, vectorStore.Name)

    // アシスタントのツールリソースにベクトルストアを追加
    assistantRequest := openai.AssistantRequest{
        ToolResources: &openai.AssistantToolResource{
            FileSearch: &openai.AssistantToolFileSearch{
                VectorStoreIDs: []string{vectorStore.ID},
            },
        },
    }

    // アシスタントを更新
    _, err = client.ModifyAssistant(ctx, assistantID, assistantRequest)
    if err != nil {
        return fmt.Errorf("アシスタントの更新に失敗しました: %w", err)
    }

    fmt.Println("アシスタントにベクトルストアが関連付けられました。")

    return nil
}

func uploadAndAttachFile(client *openai.Client, assistantID string, options Options) error {
    if options.FilePath == "" {
        fmt.Println("ファイルパスが指定されていないため、ファイルのアップロードと追加をスキップします。")
        return nil
    }

    ctx := context.Background()

    fileRequest := openai.FileRequest{
        FileName: filepath.Base(options.FilePath),
        FilePath: options.FilePath,
        Purpose:  "fine-tune",
    }

    file, err := client.CreateFile(ctx, fileRequest)
    if err != nil {
        return fmt.Errorf("ファイルのアップロードに失敗しました: %w", err)
    }
    fmt.Printf("ファイルがアップロードされました:\nID: %s\n", file.ID)

    assistantFileRequest := openai.AssistantFileRequest{
        FileID: file.ID,
    }

    _, err = client.CreateAssistantFile(ctx, assistantID, assistantFileRequest)
    if err != nil {
        return fmt.Errorf("アシスタントへのファイル追加に失敗しました: %w", err)
    }

    fmt.Println("アシスタントにファイルが追加されました。")
    return nil
}

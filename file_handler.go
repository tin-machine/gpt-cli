package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient インターフェースを定義
type OpenAIClient interface {
	CreateFile(ctx context.Context, req openai.FileRequest) (openai.File, error)
	ListFiles(ctx context.Context) (openai.FilesList, error)
	DeleteFile(ctx context.Context, fileID string) error
}

// UploadFileは指定されたファイルをOpenAI APIにアップロードする関数です。
// 引数clientはOpenAI APIのクライアント、filePathはアップロードするファイルのパス、purposeはファイルの用途です。
// 成功した場合はアップロードされたファイルの情報を含むopenai.File構造体が返されます。
// エラーが発生した場合は、そのエラーメッセージが返されます。
func UploadFile(client OpenAIClient, filePath string, purpose string) (*openai.File, error) {
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

// ListUploadedFilesは、ユーザーがOpenAIにアップロードしたファイルの一覧を取得する関数です。
// clientはOpenAI APIクライアントであり、レスポンスにはファイルの詳細が含まれます。
// 成功した場合はファイルのリストが返されますが、API呼び出しに失敗した場合はエラーが返されます。
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

// handleUploadAndAddFilesは、ユーザー指定のファイルをOpenAIにアップロードし、そのファイルをベクトルストアに追加します。
// 引数clientはOpenAI APIクライアント、optionsにはアップロード対象のファイルや追加に関する設定が含まれます。
// 成功した場合は、アップロード結果の詳細が表示され、エラーが発生した場合はエラーメッセージが返されます。
func handleUploadAndAddFiles(client *openai.Client, options Options) error {
	// ファイルをアップロード
	fileIDs, err := UploadFiles(client, options.UploadAndAddFiles, options.UploadPurpose)
	if err != nil {
		return err
	}

	// VectorStoreの取得または作成
	vectorStoreName := options.VectorStoreName
	if vectorStoreName == "" {
		vectorStoreName = fmt.Sprintf("Auto-Generated Vector Store %d", time.Now().Unix())
	}
	vectorStore, err := GetOrCreateVectorStore(client, vectorStoreName)
	if err != nil {
		return err
	}
	fmt.Printf("使用するベクトルストア: ID=%s, Name=%s\n", vectorStore.ID, vectorStore.Name)

	// ファイルをVectorStoreに追加
	err = AddFilesToVectorStore(client, vectorStore.ID, fileIDs)
	if err != nil {
		return err
	}
	fmt.Printf("ファイルをベクトルストアに追加しました: VectorStoreID=%s\n", vectorStore.ID)

	return nil // 処理が完了したので終了
}
func handleUploadFile(client *openai.Client, options Options) error {
	// ファイルの存在チェックはUploadFile関数内で行われています
	// UploadFile 関数を呼び出す
	uploadedFile, err := UploadFile(client, options.UploadFilePath, options.UploadPurpose)
	if err != nil {
		return fmt.Errorf("ファイルのアップロードに失敗しました: %v", err)
	}
	fmt.Printf("ファイルがアップロードされました。File ID: %s\n", uploadedFile.ID)
	return nil // ファイルのアップロード後にプログラムを終了
}

func handleListFiles(client *openai.Client) error {
	files, err := ListUploadedFiles(client)
	if err != nil {
		return fmt.Errorf("ファイル一覧の取得に失敗しました: %v", err)
	}
	fmt.Println("アップロードされたファイル一覧:")
	for _, file := range files.Files {
		fmt.Printf("- ID: %s, 名前: %s, ステータス: %s\n", file.ID, file.FileName, file.Status)
	}
	return nil
}

// func handleDeleteFile(client *openai.Client, options Options) error {
// 	err := DeleteUploadedFile(client, options.DeleteFileID)
// 	if err != nil {
// 		return fmt.Errorf("ファイルの削除に失敗しました: %v", err)
// 	}
// 	fmt.Printf("ファイルを削除しました。File ID: %s\n", options.DeleteFileID)
// 	return nil
// }

func handleVectorStoreAction(client *openai.Client, options Options) error {
	switch options.VectorStoreAction {
	case "create":
		if options.VectorStoreName == "" {
			return fmt.Errorf("ベクトルストアの名前を指定してください (--vector-store-name)")
		}
		vs, err := CreateVectorStore(client, options.VectorStoreName)
		if err != nil {
			return fmt.Errorf("ベクトルストアの作成に失敗しました: %v", err)
		}
		fmt.Printf("ベクトルストアを作成しました: ID=%s, Name=%s\n", vs.ID, vs.Name)
		return nil
	case "list":
		vsList, err := ListVectorStores(client)
		if err != nil {
			return fmt.Errorf("ベクトルストアの一覧取得に失敗しました: %v", err)
		}
		for _, vs := range vsList {
			fmt.Printf("ID: %s, Name: %s, Status: %s\n", vs.ID, vs.Name, vs.Status)
		}
		return nil
	case "delete":
		if options.VectorStoreID == "" {
			return fmt.Errorf("削除するベクトルストアのIDを指定してください (--vector-store-id)")
		}
		err := DeleteVectorStore(client, options.VectorStoreID)
		if err != nil {
			return fmt.Errorf("ベクトルストアの削除に失敗しました: %v", err)
		}
		fmt.Printf("ベクトルストアを削除しました。ID: %s\n", options.VectorStoreID)
		return nil
	case "add-file":
		if options.VectorStoreID == "" || (options.FileID == "" && len(options.FileIDs) == 0) {
			return fmt.Errorf("ベクトルストアIDとファイルIDを指定してください (--vector-store-id, --file-id または --file-ids)")
		}
		if options.FileID != "" {
			// 単一のファイルIDを処理
			vsFile, err := AddFileToVectorStore(client, options.VectorStoreID, options.FileID)
			if err != nil {
				return fmt.Errorf("ファイルの追加に失敗しました: %v", err)
			}
			fmt.Printf("ファイルをベクトルストアに追加しました: FileID=%s, VectorStoreID=%s\n", vsFile.ID, vsFile.VectorStoreID)
		} else if len(options.FileIDs) > 0 {
			// 複数のファイルIDを処理
			err := AddFilesToVectorStore(client, options.VectorStoreID, options.FileIDs)
			if err != nil {
				return fmt.Errorf("複数ファイルの追加に失敗しました: %v", err)
			}
			fmt.Printf("%d 個のファイルをベクトルストアに追加しました: VectorStoreID=%s\n", len(options.FileIDs), options.VectorStoreID)
		} else {
			return fmt.Errorf("ファイルIDを指定してください (--file-id または --file-ids)")
		}
		return nil
	default:
		return fmt.Errorf("不正なベクトルストアアクションが指定されました: %s", options.VectorStoreAction)
	}
}

// ファイル名でファイルを削除するためのヘルパー関数
func DeleteFilesByName(client *openai.Client, pattern string) error {
    files, err := ListUploadedFiles(client)
    if err != nil {
        return fmt.Errorf("ファイル一覧の取得に失敗しました: %w", err)
    }

    var errors []error
    for _, file := range files.Files {
        match, err := filepath.Match(pattern, file.FileName)
        if err != nil {
            return fmt.Errorf("パターンのマッチングに失敗しました: %w", err)
        }
        if match {
            err := DeleteUploadedFile(client, file.ID)
            if err != nil {
                errors = append(errors, fmt.Errorf("ファイルの削除に失敗しました。File ID: %s, エラー: %w", file.ID, err))
            } else {
                fmt.Printf("ファイルを削除しました。Name: %s, File ID: %s\n", file.FileName, file.ID)
            }
        }
    }
    if len(errors) > 0 {
        return fmt.Errorf("いくつかのファイルが削除できませんでした: %v", errors)
    }
    return nil
}

// handleDeleteFile関数を修正して、名前による削除をサポートする
func handleDeleteFile(client *openai.Client, options Options) error {
    if options.DeleteFileID != "" {
        err := DeleteUploadedFile(client, options.DeleteFileID)
        if err != nil {
            return fmt.Errorf("ファイルの削除に失敗しました: %v", err)
        }
        fmt.Printf("ファイルを削除しました。File ID: %s\n", options.DeleteFileID)
        return nil
    }

    if options.DeleteFileName != "" {
        err := DeleteFilesByName(client, options.DeleteFileName)
        if err != nil {
            return fmt.Errorf("ファイルの名前による削除に失敗しました: %v", err)
        }
    }
    return nil
}


package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// SplitImageListは、カンマで区切られた画像ファイル名のリストを分割し、スライス（配列）として返します。
// 引数imageListはカンマ区切りの文字列です。
func SplitImageList(imageList string) []string {
	return strings.Split(imageList, ",")
}

// CollectFilesは、指定されたディレクトリ内のすべてのファイル名とその内容を収集します。
// 収集した内容は、ファイル名と内容のペアとして文字列として返されます。
// 引数dirは検索開始のディレクトリです。
// .gitディレクトリはスキップされます。
func CollectFiles(dir string) (string, error) {
	var builder strings.Builder

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// .gitディレクトリをスキップ
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// ファイルの場合、名前と内容を取得
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			builder.WriteString(fmt.Sprintf("ファイル名: %s\n内容:\n%s\n\n", path, string(content)))
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

// ReadFilesは、コンマで区切られたファイル名リストからファイルを読み込み、その内容を結合して返します。
// 引数fileListは読み込むファイルのパスを示します。
// 成功した場合は内容が連結された文字列、エラーが発生した場合はそのエラーメッセージを返します。
func ReadFiles(fileList string) (string, error) {
	var builder strings.Builder
	files := strings.Split(fileList, ",")

	for _, filePath := range files {
		// グロブパターンを処理
		if matches, err := filepath.Glob(filePath); err == nil {
			for _, match := range matches {
				if err := readAndAppendFileContents(match, &builder); err != nil {
					return "", fmt.Errorf("ファイルの読み込みに失敗しました (%s): %w", match, err)
				}
			}
		} else {
			// グロブ処理が失敗した場合、通常のファイルパスとみなして読み込む
			if err := readAndAppendFileContents(filePath, &builder); err != nil {
				return "", fmt.Errorf("ファイルの読み込みに失敗しました (%s): %w", filePath, err)
			}
		}
	}
	return builder.String(), nil
}

// readAndAppendFileContents は指定されたファイルの内容をリストに追加します。
func readAndAppendFileContents(filePath string, builder *strings.Builder) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// fmt.Fprintfを使って直接builderに書き込み
	_, err = fmt.Fprintf(builder, "ファイル名: %s\n内容:\n%s\n\n", filePath, string(content))
	return err
}

// CreateMessages はプロンプト設定からメッセージを作成します
func CreateMessages(promptConfig Prompt) ([]openai.ChatCompletionMessage, error) {
	var messages []openai.ChatCompletionMessage

	// システムメッセージの追加
	if promptConfig.System != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: promptConfig.System,
		})
	}

	// ユーザーメッセージの追加
	if promptConfig.User != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: promptConfig.User,
		})
	}

	// 添付ファイル（画像）の処理
	for _, attachmentPath := range promptConfig.Attachments {
		base64Image, mimeType, err := imageToBase64(attachmentPath)
		if err != nil {
			return nil, fmt.Errorf("画像のエンコードに失敗しました: %v", err)
		}

		// 画像データをメッセージに追加
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("画像ファイル: %s\nデータ: data:%s;base64,%s", attachmentPath, mimeType, base64Image),
		})
	}

	return messages, nil
}

// imageToBase64 は画像ファイルをBase64エンコードします
func imageToBase64(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// return "", "", err
		return "", "", fmt.Errorf("ファイルを読み込む際にエラーが発生しました: %w", err)
	}

	// ファイル拡張子からMIMEタイプを推測
	ext := strings.ToLower(filepath.Ext(path))
	mimeType := getMimeType(ext)

	if mimeType == "" {
		return "", "", fmt.Errorf("サポートされていない画像形式: %s", ext)
	}

	base64Image := base64.StdEncoding.EncodeToString(data)
	return base64Image, mimeType, nil
}

// getMimeType はファイル拡張子に基づいてMIMEタイプを返します
func getMimeType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	default:
		return ""
	}
}

// SaveConversationHistory は会話履歴をファイルに保存します
func SaveConversationHistory(filename string, history []openai.ChatCompletionMessage) error {
	if filepath.Ext(filename) == "" {
		filename += ".json"
	}
	data, err := json.Marshal(history)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0600)
}

// LoadConversationHistory はファイルから会話履歴を読み込みます
func LoadConversationHistory(filename string) ([]openai.ChatCompletionMessage, error) {
	if filename == "" {
		return []openai.ChatCompletionMessage{}, nil
	}
	if filepath.Ext(filename) == "" {
		filename += ".json"
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []openai.ChatCompletionMessage{}, nil
		}
		return nil, err
	}
	var history []openai.ChatCompletionMessage
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, err
	}
	return history, nil
}

// DisplayConversationHistory は会話履歴をMarkdown形式で表示します
func DisplayConversationHistory(history []openai.ChatCompletionMessage) {
	c := cases.Title(language.Und) // 言語を指定（ここでは未指定）
	for _, message := range history {
		role := c.String(message.Role)
		fmt.Printf("### %s\n\n%s\n\n", role, message.Content)
	}
}

// GetLogDirectory は設定ファイルや環境変数に基づいてログの保存ディレクトリを取得します
func GetLogDirectory(config Config) string {
	if config.LogDir != "" {
		return config.LogDir
	}
	// 環境変数XDG_DATA_HOMEを確認
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, "gpt-cli", "logs")
	}
	// デフォルトのログパスを使用
	return filepath.Join(os.Getenv("HOME"), ".local", "share", "gpt-cli", "logs")
}

// EnsureDirectory は指定されたディレクトリパスが存在するか確認し、存在しない場合は作成します
func EnsureDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}

// inputAvailable は標準入力が利用可能かどうかを判断します
func inputAvailable() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func handleChatCompletion(client *openai.Client, promptConfig Prompt, conversationHistory []openai.ChatCompletionMessage, options Options) error {
	// -dオプションが有効な場合、Optionsの内容を出力
	if options.Debug {
		logger.Debug("現在のオプション内容:\n%s", options.String())
	}

	// OpenAI API へのリクエスト
	assistantMessage, err := ExecuteChatCompletion(client, promptConfig.Model, promptConfig.MaxTokens, conversationHistory)
	if err != nil {
		return fmt.Errorf("ChatCompletionエラー: %w", err)
	}

	// 会話履歴にアシスタントの応答を追加
	conversationHistory = append(conversationHistory, assistantMessage)

	// 会話履歴の保存
	if options.HistoryFile != "" {
		err = SaveConversationHistory(options.HistoryFile, conversationHistory)
		if err != nil {
			return fmt.Errorf("会話履歴の保存に失敗しました: %w", err)
		}
	}

	// 標準出力に結果を表示
	fmt.Println(assistantMessage.Content)

	return nil
}

// RecursiveGlob は再帰的なグロブパターンを展開します
func RecursiveGlob(pattern string) ([]string, error) {
	var matches []string

	// パターンを分割
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return nil, fmt.Errorf("パターンの形式が正しくありません: %s", pattern)
	}

	// ベースディレクトリを取得
	baseDir := parts[0]
	if baseDir == "" {
		baseDir = "."
	}

	subPattern := parts[1]

	// ベースディレクトリから再帰的に探索
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリはスキップ
		if info.IsDir() {
			return nil
		}

		// サブパターンにマッチするか確認
		match, err := filepath.Match(subPattern, filepath.Base(path))
		if err != nil {
			return err
		}
		if match {
			matches = append(matches, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return matches, nil
}

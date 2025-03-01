# 概要

ChatGPTに度々問い合わせを行っているのですが、
viで書いたテキストを貼り付ける事が多いので、CLIから問い合わせできたら楽だな、という発想のコマンドです。

# インストール

```
go install github.com/tin-machine/gpt-cli@latest
```

[環境変数 OPEN_API_KEY にChatGPTのAPIキーを設定してください](https://github.com/tin-machine/gpt-cli/blob/21c4889a98cda54f3dc222bf32c00f02e26a11f0/openai_client.go#L17)

```bash
export OPENAI_API_KEY="your-api-key-here"
```

# 使い方

## シンプルな使い方

- **単純に聞くだけの場合**:
```
gpt-cli "こんにちは！"
```

- **標準入力から**:
```
echo "こんにちは！" | gpt-cli
```

- **システムプロンプトとユーザープロンプトを指定している場合**:
```
gpt-cli -s "なるべく陽気に答えてください" -u "こんにちは"
```

- **会話ログを-histroyで保存しつつ会話**:
```
gpt-cli -p prompt4 -history gpt-cli改修 -u "何か改修できる点を教えてください"
```

- **会話ログを表示**:
```
gpt-cli -show-history gpt-cli改修
```

- **下に記載しているconfig.yamlを設定している場合**:
```
gpt-cli -p prompt1
```

- **ファイルをユーザープロンプトに追加して会話**:
```
gpt-cli -p prompt4 -history gpt-cli改修 -f main.go,config.go,utils.go -u "何か改修できる点を教えてください"
```

## Assistant APIを使う

OpenAIの管理画面のStorageでは[File]と[Vector store]という機能があります。
ChatGPTのAssistant APIを使う際には、これらの機能を使ってファイルをアップロードし、ベクトルストアに追加する必要があります。

- **ファイルをアップロードする例**:

```bash
gpt-cli --upload-file "path/to/file.txt" --upload-purpose "assistants"
```

- ファイルを削除する例:

```bash
gpt-cli -delete-file '*.go'
```

- **ベクトルストアをする例**:

```bash
gpt-cli --vector-store-name "my_vector_store" --vector-store-action create
```

- --vector-store-actionのオプション
  - create
  - list
  - delete
  - add-file

- **ベクトルストア作って中にファイルを追加する例**:
ファイルをアップロードしてからVectore storesに自動的に入れる場合

```bash
gpt-cli --upload-and-add-to-vector '*.go' --vector-store-name "my_vector_store"
```

- **ベクトルストアの一覧を表示する例**:

```bash
gpt-cli --vector-store-action list
```

- ベクトルストアを削除する例:

```bash
gpt-cli --vector-store-action delete --vector-store-id <ベクトルストアのID>
```

- **複数ファイルをVector-storeにアップロードする例**:
  - --upload-and-add-to-vector: ファイルをカンマ区切り
  - --vector-store-name: ベクトルストアの名前
  - --upload-purpose: アップロード目的を引数に取る
    - assistants: アシスタント
    - assistants_output: アシスタントの結果
    - batch: バッチ処理
    - fine-tune: ファインチューニング
    - fine-tune-results: ファインチューニング結果を置く

```bash
gpt-cli --upload-and-add-to-vector assistant_handler.go,config.go,config_loader.go,file_handler.go,main.go,openai_client.go,options.go,prompt_config.go,tool_config.go,utils.go,vector_store_handler.go -vector-store-name add-option -upload-purpose assistants
```

- **アシスタントを作成する例**:

```bash
gpt-cli --create-assistant --assistant-name "MyAssistant" --assistant-description "これはテスト用のアシスタントです。" --user-message "あなたはユーザーを助けるフレンドリーなアシスタントです。"
```

- アシスタント作成時、vectore-storeにアップロードしたファイルを追加する例:

```bash
gpt-cli --create-assistant --assistant-name "MyAssistant" --upload-and-add-to-vector '*.go' --vector-store-name "my_vector_store"
```

- **アシスタントと対話する例**:

```bash
gpt-cli --assistant-id "assistant_id" --message "こんにちは！"
```

- アシスタントの設定をconfig.yamlで行う:

config.yamlからassistantsの設定を探してアシスタントを操作する

```
gpt-cli -a myassistant1
```

config.yamlの設定は下記の通り:
```
vectorStores:
  myVectorStore:
    name: "my_vectore_store"

ssistants:
  myassistant1:  # アシスタントのキー名、コマンドラインから指定する際はこのキーを使用します
    name: "My First Assistant"  # アシスタントの名前
    description: "This is a description of my first assistant."  # アシスタントの説明
    model: "gpt-3.5-turbo"  # 使用するAIモデル
    instruction: "あなたはユーザーを助けるアシスタントです。"  # アシスタントに与える初期の指示
    temperature: 0.7  # モデルの応答の多様性を制御するパラメータ
    vectorStoreName: "my_vector_store"  # 関連付けたいベクトルストアの名前
```

# オプション

取り得るオプションですが
- `-s`: システムプロンプトで文字列を与えたい(config.yamlを上書き)
- `-u`: ユーザープロンプトで文字列を与えたい(config.yamlを上書き)
- `-p`: config.yaml内でベースとなるプロンプトを選択
  - `-p` が無い場合は `-s` か `-u` が必要
- `-i`: イメージファイルをカンマ区切りで与える
- `-c`: config.yamlのパスを指定
- `-m`: モデルを指定
- `-d`: デバックモード
- `-v`: バージョン
- `-collect`: 現在のディレクトリ内のファイルをUserプロンプトに追加
- `-history`: 会話履歴の保存ファイルを指定（拡張子は不要）
- `-t`: タイムアウト時間（秒）を指定
- `-f`: 読み込むファイルのパスをカンマ区切りで指定
- `-show-history`: 会話ログを表示
- `--vector-store-action`: ベクトルストアのアクションを指定（`create`、`list`、`delete`、`add-file`）。
- `--vector-store-name`: 作成または操作するベクトルストアの名前を指定。
- `--vector-store-id`: 操作するベクトルストアのIDを指定。
- `--upload-file`: OpenAIにアップロードするファイルのパスを指定。
- `--upload-purpose`: ファイルのアップロード目的を指定（例：`fine-tune`, `assistants`, `batch`）。
- `--list-files`: アップロードしたファイルの一覧を表示。
- `--delete-file`: 削除するファイルのIDを指定。
- `--upload-and-add-to-vector`: ファイルをアップロードし、ベクトルストアに追加。
- `--assistant-id`: 操作するアシスタントのIDを指定。
- `--create-assistant`: 新しいアシスタントを作成。
- `--assistant-name`: アシスタントの名前を指定。
- `--assistant-description`: アシスタントの説明を指定。
- `--instruction`: アシスタントへの指示を指定。
- `--message`: アシスタントに送信するメッセージを指定。
- `--file-id`: ベクトルストアに追加するファイルのIDを指定。
- `--file-ids`: ベクトルストアに追加するファイルのIDをカンマ区切りで指定。

# config.yamlのサンプル

~/.config/gpt-cli/config.yaml に配置してください


autoSaveLogs が true の場合、会話ログを保存します。
autoSaveLogsがtrueでlogDirを指定しない場合、[環境変数XDG_DATA_HOMEが設定sれている場合は$XDG_DATA_HOME/gpt-cli/、設定されていない場合は$HOME/.local/share/gpt-cli/に保存します。](https://github.com/tin-machine/gpt-cli/blob/c683710784958f33760741fabf3ce4cdbfc76607/utils.go#L183)
この保存ファイル名は-histroyで指定したファイル名で変更できます。会話の文脈を繋げたい場合は -history で指定した方が会話が繋がります。

```
# logDir: "<会話ログを保存するディレクトリ>"
autoSaveLogs: true # 自動的に会話ログを保存するか
prompts:
  prompt1:
    model: gpt-4o
    maxTokens: 16384
    system: |
      "深呼吸し順番にゆっくり考えてみてください。出力するの最終的なものだけにしてください(途中の考えてほしい段階は出力しないでください)"
      "次の文章はランダムなタスクになっています。文章すべて最後まで読み込み"
      "タスクを分析、分解し、依存関係を考えて私が次に行うべき重要なタスクを3個ピックアップ、そのタスクを行う事で可能になる未来を考え楽しくなるように話をふくらませるようにしてください。"
      "最終的にボリュームのある一つの文章にまとめて300文字以内で出力してください。"
    user: |
      -uオプションで必要なソースコードを上書きする
  prompt2:
    model: gpt-4o
    system: |
      "ずんだもんっぽい文章にしてください、その際、変換するなどの作業工程は省き、喋っているように内容だけ出力してください"
    user: |
      -uオプションで必要なソースコードを上書きする
  prompt3:
    model: gpt-4o
    system: |
      "動作確認です。添付ファイルは存在するでしょうか?またどんなファイルでしょうか?画像ファイルの場合、内容について日本語で表現して欲しいです"
    user: |
      -uオプションで必要なソースコードを上書きする
  prompt4:
    model: gpt-4o
    system: |
      "ソースコードのレビューをお願いします。何か改善点を日本語で複数あげてほしいです。"
    user: |
      -uオプションで必要なソースコードを上書きする
  prompt5:
    model: gpt-4o
    system: |
      次のプログラムの改修点をいくつかあげてください。
      複数ある場合、ファイルごとに指摘事項をあげてください。
      ソースコードを出力する場合、ソースコード全てを出力してください。
    user: |
      -uオプションで必要なソースコードを上書きする
```

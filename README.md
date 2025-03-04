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

- 単純に聞く
```
gpt-cli "こんにちは！"
```

- モデルを指定

```
gpt-cli -m gpt-4o "こんにちは！"
```

- 標準入力から

```
echo "こんにちは！" | gpt-cli
```

- システムプロンプトとユーザープロンプトを指定している場合

```
gpt-cli -s "なるべく陽気に答えてください" -u "こんにちは"
```

- 会話ログを-histroyで保存しつつ会話
```
gpt-cli -p prompt4 -history gpt-cli改修 -u "何か改修できる点を教えてください"
```

- 会話ログ

```
gpt-cli -show-history gpt-cli改修
```

- 下に記載しているconfig.yamlを設定している設定から選択する

```
gpt-cli -p prompt1
```

config.yaml に -p で指定した設定が無い場合、-u か -s での指定が必要です。

- ファイルをユーザープロンプトに追加して会話
```
gpt-cli -p prompt4 -history gpt-cli改修 -f main.go,config.go,utils.go -u "何か改修できる点を教えてください"
```

## Assistant APIを使う

ChatGPTのAssistant APIからファイルを扱う場合、一旦、ファイルをStorage->Fileにアップロードし、更にStorage->Vectore storesにに追加する必要があります。

コストがかかります。 最初の1GBはフリーですが、それ以降は$0.1/MBです。
[料金](https://openai.com/pricing)のAssistant API を参照

### アシスタントを作成する
カレントディレクトリの *.go を<Vectore_store_name>にアップロード、<アシスタント名>アシスタントを作成する例です。

```bash
gpt-cli --model "gpt-4o" \
        --assistant-description "これはテスト用のアシスタントです。"  \
        --instruction "あなたはユーザーを助けるフレンドリーなアシスタントです。" \
        --upload-and-add-to-vector '*.go' \
        --assistant-name "<アシスタント名>" \
        --vector-store-name "<Vectore_store_name>"
```

- `--assistant-description`: アシスタントの説明を指定。
- `--instruction`: アシスタントへの指示を指定。

### 既に作成ずみアシスタントと対話

```bash
gpt-cli --assistant-id "assistant_id" --message "こんにちは！"
```

### Storage->Files の操作

ファイルをアップロード

```
gpt-cli --upload-file <ファイルのパス> 
```

ファイルをStorage->Filesから削除する例:

```bash
gpt-cli -delete-file '*.go'
```

ファイルのリスト

```
gpt-cli --list-files
```


### Storage->Vectore stores の操作

ファイルをVectore storesを作成してアップロード

```bash
gpt-cli --vector-store-name "20250303-2" --upload-file "path/to/file.txt" --upload-purpose "assistants"
```

- --upload-purpose: アップロード目的を引数に取る
  - assistants: アシスタント
  - assistants_output: アシスタントの結果
  - batch: バッチ処理
  - fine-tune: ファインチューニング
  - fine-tune-results: ファインチューニング結果を置く

- ベクトルストア作って中にファイルを追加する
ファイルをアップロードしてからVectore storesに自動的に入れる場合

```bash
gpt-cli --upload-and-add-to-vector '*.go' --vector-store-name "my_vector_store"
```

ベクトルストアの作成だけする

```bash
gpt-cli --vector-store-name "my_vector_store" --vector-store-action create
```

ベクトルストアの一覧を表示

```bash
gpt-cli --vector-store-action list
```

ベクトルストアを削除

```bash
gpt-cli --vector-store-action delete --vector-store-id <ベクトルストアのID>
```

ベクターストアに追加

```bash
gpt-cli --vector-store-action add-file --file-ids <ファイルのIDをカンマ区切り> --vector-store-id <ベクトルストアのID>
```

### アシスタントの設定をconfig.yamlで行う

config.yamlからassistantsの設定を探してアシスタントを操作する

```
gpt-cli -a myassistant1
```

config.yamlの設定

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

他に取り得るオプションですが
- `-c`: config.yamlのパスを指定
- `-d`: デバックモード
- `-v`: バージョン
- `-collect`: 現在のディレクトリ内のファイルをUserプロンプトに追加
- `-t`: タイムアウト時間（秒）を指定

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

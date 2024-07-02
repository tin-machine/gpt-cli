# 概要

ChatGPTに度々問い合わせを行っているのですが、
viで書いたテキストを貼り付ける事が多いので、CLIから問い合わせできたら楽だな、という発想のコマンドです。

# インストール

```
go install github.com/tin-machine/gpt-cli@latest
```

# 使い方

```
gpt-cli -p prompt1 -m add-text.txt -o output.txt -d
```

# オプション

取り得るオプションですが
`-p`: config.yaml内でベースとなるプロンプトを選択
`-s`: システムプロンプトで文字列を与えたい(config.yamlを上書き)
`-u`: ユーザープロンプトで文字列を与えたい(config.yamlを上書き)
`-i`: イメージファイルをカンマ区切りで与える(config.yamlを上書き)
`-a`: 追加する文章へのパスを指定します。
`-o`: 出力するファイル名
`-d`: デバックモード

`-p` での指定が無い場合、`-s` か `-u` が必要。

# config.yamlのサンプル

現状は下記ファイルがあるディレクトリで実行するようになっています。

```
prompts:
  prompt3:
    system: |
      添付ファイルの動物を当ててください。
    user: |
      どんな動物でしょうか?
    attachments:
      - dog.jpg
      - cat.jpeg
      - human.png
  prompt4:
    system: |
      次の文章を構成してください
    user: |
      おはようございます、今日も良い天気ですね
```

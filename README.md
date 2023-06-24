# knoQ

イベント管理システム

## docs

[wiki](https://github.com/traPtitech/knoQ/wiki)

## 開発

### 必要要件

- go
- make
- docker
- docker-compose

### サーバーの起動

```bash
docker compose up --build
```

`http://localhost:6006`に knoQ が起動します。
`http://localhost:8000`に phpmyadmin が起動します。

現在、ログインできるのは traP ユーザーのみです。

### 環境変数の設定・追加のファイル

knoQ の全ての機能を動作させるためには、追加の情報が必要です。

| 名前                  | 種類   | デフォルト                                  | 説明                                             |
|---------------------|------|----------------------------------------|------------------------------------------------|
| SESSION_KEY         | 環境変数 | `random32wordsXXXXXXXXXXXXXXXXXXX`     | session を暗号化するもの                               |
| TRAQ_CALENDARID     | 環境変数 |                                        | 進捗部屋の提供元（公開されている google calendar の id なら何でもいい） |
| CLIENT_ID           | 環境変数 | `aYj6mwyLcpBIrxZZD8jkCzH3Gsdqc9DJqle2` | 認証に必要                                          |
| WEBHOOK_ID          | 環境変数 |                                        | Bot 情報                                         |
| WEBHOOK_SECRET      | 環境変数 |                                        | Bot 情報                                         |
| CHANNEL_ID          | 環境変数 |                                        | Bot の送信先チャンネル (deprecated)                     |
| DAILY_CHANNEL_ID    | 環境変数 |                                        | Bot が毎日定時に投稿する先のチャンネル                          |
| ACTIVITY_CHANNEL_ID | 環境変数 |                                        | Bot が都度送信するチャンネル                               |
| TOKEN_KEY           | 環境変数 | `random32wordsXXXXXXXXXXXXXXXXXXX`     | Token を暗号化する。長さ 32 文字のランダム文字列。存在しない場合はエラー。     |
| KNOQ_VERSION        | 環境変数 | UNKNOWN                                | knoQ のバージョン (github actions でイメージ作成時に指定)       |
| KNOQ_REVISION       | 環境変数 | UNKNOWN                                | git の sha1 (github actions でイメージ作成時に指定)        |
| DEVELOPMENT         | 環境変数 |                                        | 開発時かどうか                                        |
| service.json        | ファイル | 空のファイル                                 | google calendar api に必要（権限は必要なし）               |

### テスト

```bash
go test ./...
```

### コード生成

```bash
go generate ./...
```

## コードフォーマット

```bash
golangci-lint run --fix ./...
```

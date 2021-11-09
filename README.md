# knoQ

イベント管理システム

## docs

[wiki](https://github.com/traPtitech/knoQ/wiki)

## 開発

### 必要要件

- go 1.16
- make
- docker
- docker-compose

### サーバーの起動

```
> cd ./development
> make init
> docker-compose up -d
```

`http://localhost:6006`に knoQ が起動します。
また、`http://localhost:8000`に phpmyadmin が起動します。

現在、ログインできるのは traP ユーザーのみです。

### 環境変数の設定・追加のファイル

knoQ の全ての機能を動作させるためには、追加の情報が必要です。

| 名前            | 種類     | デフォルト                             | 説明                                                                       |
| --------------- | -------- | -------------------------------------- | -------------------------------------------------------------------------- |
| SESSION_KEY     | 環境変数 | `random32wordsXXXXXXXXXXXXXXXXXXX`     | session を暗号化するもの                                                   |
| TRAQ_CALENDARID | 環境変数 |                                        | 進捗部屋の提供元（公開されている google calendar の id なら何でもいい）    |
| CLIENT_ID       | 環境変数 | `aYj6mwyLcpBIrxZZD8jkCzH3Gsdqc9DJqle2` | 認証に必要                                                                 |
| WEBHOOK_ID      | 環境変数 |                                        | Bot 情報                                                                   |
| WEBHOOK_SECRET  | 環境変数 |                                        | Bot 情報                                                                   |
| CHANNEL_ID      | 環境変数 |                                        | Bot の送信先チャンネル                                                     |
| TOKEN_KEY       | 環境変数 | `random32wordsXXXXXXXXXXXXXXXXXXX`     | Token を暗号化する。長さ 32 文字のランダム文字列。存在しない場合はエラー。 |
| DEVELOPMENT     | 環境変数 |                                        | 開発時かどうか                                                             |
| service.json    | ファイル | 空のファイル                           | google calendar api に必要（権限は必要なし）                               |

### テスト

#### テスト環境の構築

テストするために、db(`localhost:3306`), traQ(`localhost:3000`)を起動します。

```
> cd ./development/test
> docker-compose up -d
```

#### 実行

```
> go test ./infra...
```

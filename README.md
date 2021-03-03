# knoQ
イベント管理システム
## 開発
### 必要要件
- go 1.15
- make
- docker
- docker-compose

### サーバーの起動
```
> cd development
> make init
> docker-compose up -d
```
`http://localhost:6006`にknoQが起動します。
また、`http://localhost:8000`にphpmyadminが起動します。

### 環境変数の設定・追加のファイル
knoQの全ての機能を動作させるためには、追加の情報が必要です。

| 名前 | 種類 | デフォルト | 説明 |
| - | - | - | - |
| SESSION_KEY | 環境変数 | `random32wordsXXXXXXXXXXXXXXXXXXX` | sessionを暗号化するもの |
| TRAQ_CALENDARID | 環境変数 | | 進捗部屋の提供元（公開されているgoogle calendarのidなら何でもいい） |
| CLIENT_ID | 環境変数 | `aYj6mwyLcpBIrxZZD8jkCzH3Gsdqc9DJqle2` | 認証に必要 |
| WEBHOOK_ID | 環境変数 | | Bot情報 |
| WEBHOOK_SECRET | 環境変数| | Bot情報 |
| CHANNEL_ID | 環境変数 | | Botの送信先チャンネル |
| service.json | ファイル | 空のファイル | google calendar apiに必要（権限は必要なし） |

### テスト
#### テスト環境の構築
テストするために、db(`localhost:3306`), traQ(`localhost:3000`)を起動します。
```
> cd development/test
> docker-compose up -d
```

#### 実行
```
> go test ./repository
```

# gh-image-fetch

GitHub の issue / PR に添付された画像・ファイル(`https://github.com/user-attachments/assets/<uuid>` 形式の URL)を、
gh CLI の認証情報を使ってダウンロードする [gh extension](https://cli.github.com/manual/gh_extension) です。

トークンの保存・管理は一切行いません。[go-gh](https://github.com/cli/go-gh) 経由で
gh CLI 自身の認証(`GH_TOKEN` / `GH_HOST` 環境変数、または `gh auth login` で保存済みの OAuth トークン)を
そのまま利用します。

## 必要条件

- [GitHub CLI (gh)](https://cli.github.com/) がインストールされ、`gh auth login` 済みであること

## インストール

```sh
gh extension install kentem-re-hiramatsu/gh-image-fetch
```

ソースからインストールする場合(要 Go 1.26+):

```sh
git clone https://github.com/kentem-re-hiramatsu/gh-image-fetch
cd gh-image-fetch
gh extension install .
```

## 使い方

```sh
# URL 全体を指定してファイルパスに保存
gh image-fetch download https://github.com/user-attachments/assets/<uuid> ./screenshot.png

# UUID のみ指定
gh image-fetch download <uuid> ./screenshot.png

# 保存先に既存ディレクトリを指定すると、<uuid> + Content-Type から推測した拡張子で保存される
gh image-fetch download <uuid> ./images/
```

- 保存先に同名ファイルがある場合は**上書き**されます。
- `..` を含む保存先パスはパストラバーサル対策のため拒否されます。

## エラーメッセージ

| 状況 | 表示 |
| --- | --- |
| gh 未認証 / トークン期限切れ | `gh auth login` での再認証を案内 |
| アセットが存在しない / アクセス権なし (404) | アセット不存在またはリポジトリへのアクセス権不足の可能性を案内 |
| レート制限 (403/429) | リトライ時期を案内 |

## private リポジトリの添付について

public / private どちらのリポジトリの添付も取得できます(private は実環境で動作確認済み)。
ただし private の添付を取得するには、gh のトークンが対象リポジトリへのアクセス権を
持っている必要があります。権限がない場合は 404 が返ります。

## 手動テスト手順

public リポジトリの実際の添付 URL で動作確認します。

1. ブラウザで任意の public リポジトリの issue を開き、画像添付を含むものを探す
   (例: [cli/cli の issues](https://github.com/cli/cli/issues) からスクリーンショット付きの issue)。
2. issue 本文の画像を右クリック →「画像アドレスをコピー」で
   `https://github.com/user-attachments/assets/<uuid>` 形式の URL を取得する。
3. ダウンロードを実行する:

   ```sh
   gh image-fetch download "<コピーしたURL>" ./test.png
   ```

   `Saved ./test.png (NNN bytes)` と表示されることを確認する。
4. 保存されたファイルが画像として開けることを確認する。
5. ディレクトリ指定の動作確認:

   ```sh
   mkdir -p ./out
   gh image-fetch download "<コピーしたURL>" ./out/
   ```

   `./out/<uuid>.png` (拡張子は Content-Type による) が生成されることを確認する。
6. エラー系の確認:

   ```sh
   # 存在しない UUID → 404 メッセージ
   gh image-fetch download 00000000-0000-4000-8000-000000000000 ./x.png

   # 不正な入力 → パースエラー
   gh image-fetch download not-a-uuid ./x.png

   # パストラバーサル → 拒否
   gh image-fetch download "<コピーしたURL>" ../evil.png
   ```

## 開発

```sh
go build ./...
go test ./...
```

## スコープ外(今後のフェーズ)

- issue / PR 本文からの添付 URL 自動抽出
- GitHub Enterprise (`GH_HOST`) 対応(内部的に host はパラメータ化済みで拡張可能)

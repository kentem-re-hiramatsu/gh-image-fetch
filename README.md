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

リリースバイナリは、`v*` タグの push を契機に GitHub Actions
([cli/gh-extension-precompile](https://github.com/cli/gh-extension-precompile))が自動生成します。

ソースからインストールする場合(要 Go 1.26+):

```sh
git clone https://github.com/kentem-re-hiramatsu/gh-image-fetch
cd gh-image-fetch
go build -o gh-image-fetch.exe .   # macOS/Linux は -o gh-image-fetch
gh extension install .
```

## 使い方(基本の3ステップ)

### ① 画像の URL をコピーする

GitHub の issue / PR を開き、本文に貼られている画像を右クリック →「画像アドレスをコピー」。
次のような URL が取れます:

```
https://github.com/user-attachments/assets/27ecac64-b73f-4ad7-ac47-a4071db12c76
```

### ② 最初に1回だけ、デフォルト保存フォルダーを設定する

```sh
gh image-fetch config set dir C:\Users\me\Pictures\gh-attachments
```

フォルダーが無ければダウンロード時に自動作成されます。設定内容は
`gh image-fetch config get dir` で確認できます。

### ③ ダウンロードする

```sh
gh image-fetch download "<コピーした画像のURL>"
```

②のフォルダーに `<日時>-<ID先頭8桁><拡張子>`(例: `20260722-093015-27ecac64.png`)という
名前で保存されます。時系列で並び、連続ダウンロードでも名前が衝突しません。

## 応用: 保存先をその場で指定する

第2引数に保存先を渡すと、そのときだけデフォルトフォルダー以外に保存できます。

```sh
# ファイルパスを指定 → その名前で保存
gh image-fetch download "<画像のURL>" ./screenshot.png

# 既存ディレクトリを指定 → その中に <ID> + 推測した拡張子で保存
gh image-fetch download "<画像のURL>" ./images/
```

- URL の代わりに、URL 末尾の ID 部分(UUID)だけを渡すこともできます(例: `gh image-fetch download 27ecac64-b73f-4ad7-ac47-a4071db12c76`)。
- 保存先に同名ファイルがある場合は**上書き**されます。
- `..` を含む保存先パスはパストラバーサル対策のため拒否されます。

## デフォルト保存フォルダーの詳細

- 設定は `%AppData%\gh-image-fetch\config.json`(OS の標準設定ディレクトリ)に保存されます。認証情報は保存しません。
- 環境変数 **`GH_IMAGE_FETCH_DIR`** が設定されている場合はそちらが**優先**されます(一時的な切り替えや CI 用)。
- 未設定のまま保存先を省略するとエラーになり、設定方法が案内されます。

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
6. デフォルト保存フォルダーの動作確認:

   ```sh
   gh image-fetch config set dir ./default-out
   gh image-fetch download "<コピーしたURL>"
   ```

   `./default-out/<日時>-<uuid先頭8桁>.png` が生成されることを確認する。

7. エラー系の確認:

   ```sh
   # 存在しない UUID → 404 メッセージ
   gh image-fetch download 00000000-0000-4000-8000-000000000000 ./x.png

   # 不正な入力 → パースエラー
   gh image-fetch download not-a-uuid ./x.png

   # パストラバーサル → 拒否
   gh image-fetch download "<コピーしたURL>" ../evil.png
   ```

## Claude Code から使う(スキル配布)

[skills/fetch-issue-images/](skills/fetch-issue-images/) に、Claude Code 用のスキル
(issue / PR の添付画像を自動でダウンロードして確認する手順書)を同梱しています。

自分の環境にコピーすると、どのリポジトリで作業していても
「この issue のスクリーンショットを見て」のような依頼で Claude が自動的にこのツールを使うようになります:

```sh
# ユーザーレベル(全プロジェクトで有効)
mkdir -p ~/.claude/skills
cp -r skills/fetch-issue-images ~/.claude/skills/

# または特定プロジェクトのみ(チームで共有する場合はこちらをコミット)
cp -r skills/fetch-issue-images <プロジェクト>/.claude/skills/
```

前提として `gh extension install kentem-re-hiramatsu/gh-image-fetch` が必要です。

## 開発

```sh
go build ./...
go test ./...
```

## スコープ外(今後のフェーズ)

- issue / PR 本文からの添付 URL 自動抽出
- GitHub Enterprise (`GH_HOST`) 対応(内部的に host はパラメータ化済みで拡張可能)

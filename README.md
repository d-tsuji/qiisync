# Qiisync

[![Actions Status](https://github.com/d-tsuji/qiisync/workflows/test/badge.svg)](https://github.com/d-tsuji/qiisync/actions)

Qiisync は Qiita への記事の投稿や更新に便利な CLI クライアントです。

## インストール

### Binary

Binary が必要な場合は [Releases](https://github.com/d-tsuji/qiisync/releases) ページから欲しいバージョンの zip ファイルをダウンロードしてください。
zip ファイルを解凍し、パスが通る場所に Binary を配置します。

### macOS

```
$ brew tap d-tsuji/qiisync
$ brew install qiisync
```

### CentOS

```
$ sudo rpm -ivh https://github.com/d-tsuji/qiisync/releases/download/v0.0.1/qiisync_0.0.1_Tux-64-bit.rpm
```

### Debian, Ubuntu

```
$ wget https://github.com/d-tsuji/qiisync/releases/download/v0.0.1/qiisync_0.0.1_Tux-64-bit.deb
$ sudo dpkg -i qiisync_0.0.1_Tux-64-bit.deb
```

### Golang

```
$ go get -u github.com/d-tsuji/qiisync
```

## 使い方

### 設定

Qiisync を使うためには Qiita の API トークンが必要です。[こちら](https://qiita.com/settings/applications)から取得できます。

次に設定ファイルを書きます。ホームディレクトリ配下の `~/.config/qiisync/config` に、以下のような TOML ファイルを置いてください。

```toml
[qiita]
api_token = "1234567890abcdefghijklmnopqrstuvwxyz1234"

[local]
base_dir = "./testdata/output"
filename_mode = "title"
```

設定ファイルのおける各項目の説明です。

#### [qiita]

| #   | 項目        | 説明                                | デフォルト値 |
| --- | ----------- | ----------------------------------- | ------------ |
| 1   | `api_token` | Qiita の API トークンを設定します。 | <必須>       |

#### [local]

| #   | 項目            | 説明                                                                                                                                                                                               | デフォルト値 |
| --- | --------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| 1   | `base_dir`      | 記事を格納するパスのルートです。                                                                                                                                                                   | <必須>       |
| 2   | `filename_mode` | 記事をローカルに取得する際のファイル名です。"title" か "id" を指定できます。<br>"title" はファイル名に、Qiita の記事のファイル名を、"id" の場合は記事のファイル名に Qiita の記事の ID を用います。 | "title"      |

### 記事の操作

Qiisync では以下の 3 つの操作をサポートしています。

- Qiita から記事のダウンロード
- Qiita へ記事を更新
- Qiita へ記事を投稿

#### 記事のダウンロード (qiisync pull)

```bash
$ qiisync pull
```

設定完了後、上記のコマンドで Qiita の記事を `base_dir` で指定したディレクトリ配下にダウンロードできます。

`base_dir` を `"./testdata/output/pull"` に設定して `qiisync pull` を実行したときは以下のようにダウンロードされます。
`base_dir` 配下に記事を作成した日付ごとにディレクトリが作成されて、その中に記事が保存されます。

```bash
$ ./qiisync pull
     fresh remote=2020-04-14 11:26:38 +0900 JST > local=0001-01-01 00:00:00 +0000 UTC
     store /mnt/c/Users/dramt/go/src/github.com/d-tsuji/qiisync/testdata/output/pull/20200413/改行コードって難しいっ.md
     ...
     fresh remote=2019-12-05 07:01:29 +0900 JST > local=0001-01-01 00:00:00 +0000 UTC
     store /mnt/c/Users/dramt/go/src/github.com/d-tsuji/qiisync/testdata/output/pull/20191124/GoでシンプルなHTTPサーバを自作する.md
     fresh remote=2019-12-10 07:00:25 +0900 JST > local=0001-01-01 00:00:00 +0000 UTC
     store /mnt/c/Users/dramt/go/src/github.com/d-tsuji/qiisync/testdata/output/pull/20191118/GoのFormatterの書式における'+'フラグと独自実装.md
     fresh remote=2019-11-20 10:33:03 +0900 JST > local=0001-01-01 00:00:00 +0000 UTC
     ...
```

`filename_mode` で `"id"` を指定しているとダウンロードしたときのファイル名は以下のようになります。

```bash
$ ./qiisync pull
     fresh remote=2020-04-14 11:26:38 +0900 JST > local=0001-01-01 00:00:00 +0000 UTC
     store /mnt/c/Users/dramt/go/src/github.com/d-tsuji/qiisync/testdata/output/pull/20200413/1234567890abcdefghij.md
```

#### ファイルのフォーマット

ローカルにダウンロードした記事のフォーマットは以下の YAML 形式のメタデータを含んでいます。記事を更新する際に、このメタデータを修正して記事を更新すると、更新した内容が反映されます。なお `ID` と `Author` は更新できません。

```
---
ID: 1234567890abcdefghij
Title: はじめてのGo
Tags: Go,はじめて
Author: Tsuji Daishiro
Private: false
---

## はじめに

...
```

各メタデータの説明です。

| #   | 項目      | 説明                                                                  |
| --- | --------- | --------------------------------------------------------------------- |
| 1   | `ID`      | Qiita 上の記事を一意に特定する ID                                     |
| 2   | `Title`   | Qiita の記事のタイトル                                                |
| 3   | `Tags`    | Qiita 上の記事に付与するタグ                                          |
| 4   | `Author`  | 記事を投稿したユーザ名                                                |
| 5   | `Private` | 記事が限定公開かどうか。true の場合は限定公開、false の場合は一般公開 |

#### 記事の更新 (qiisync update)

```bash
$ qiisync update <filepath>
```

`qiisync update` を実行したときの実行例を記載します。`qiisync pull` でローカルにダウンロードしたメタデータが付与されているファイルを指定します。

```bash
$ qiisync update ./testdata/output/pull/20200423/はじめてのGo.md
      post fresh article ---> https://qiita.com/tutuz/private/1234567890abcdefghij
```

ファイルがリモートの記事よりも新しくない場合は、更新は行われません。ローカルファイルの更新日時と Qiita 上の記事の更新日時を比較して判定します。

```bash
$ qiisync update ./testdata/output/pull/20200423/はじめてのGo.md
           article is not updated. remote=2020-04-23 13:34:50 +0900 JST > local=2020-04-23 13:33:10.8990083 +0900 JST
```

#### 記事の投稿 (qiisync post)

```bash
$ qiisync post <filepath>
```

まだ Qiita に存在しない記事を投稿する場合は `qiisync post` で記事を投稿します。引数に任意のファイルパスを指定します。
投稿に成功するとメタデータが付与されたファイルが `base_dir` で指定したディレクトリ配下にダウンロードされます。以降はダウンロードされたファイルを更新し、`qiisync update` を実行することで Qiita に変更内容を反映することができます。

`qiisync post` を実行したときの実行例を記載します。投稿時に、タイトル、タグ、限定公開にするかどうかを確認します。これらは標準入力から受け取ります。

```
$ ./qiisync post ./testdata/qiita/post/test_article.md

Please enter the "title" of the article you want to post.
はじめてのGo

Please enter the "tag" of the article you want to post.
Tag is like "React,redux,TypeScript" or "Go" or "Python:3.7". To specify more than one, separate them with ",".
Go:1.14

Do you make the article you post private? "true" is private, "false" is public.
true
      post article ---> https://qiita.com/tutuz/items/private/1234567890abcdefghij
     store /mnt/c/Users/dramt/go/src/github.com/d-tsuji/qiisync/testdata/output/pull/20200423/はじめてのGo.md
```

## 制限事項

- Winodws 

Windows 環境でも動作しますが、今のところ Qiisync が Windows の改行コード CRLF(`\r\n`) をサポートしていないため、`qiisync post` でファイルを投稿する際のファイルの改行コードは LF(`\r`) である必要があります。
また、`~/.config/qiisync/config` に記述する `base_dir` も `"testdata\\output\\pull\\"` といったように `\` をエスケープする必要があります。

## ライセンス

このソフトウェアは [MIT](https://github.com/d-tsuji/qiisync/blob/master/LICENSE) ライセンスの下でライセンスされています。

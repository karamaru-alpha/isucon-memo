---
name: isucon-survey
description: ISUCONのベンチ後に競技サーバーで `make slow`(pt-query-digestでスロークエリ集計)と `make kataru`(kataribeでアクセスログ集計)を実行し、`./analyze.sh` で karamaru-alpha/isucon2026 の issue #1 にコメント投稿したうえで、ボトルネックを要約する。「計測して」「集計して」「ログ解析」「isucon-survey」「issueに貼って」などの言葉が出たら使う。
---

# isucon-survey

ベンチ 1 回分のログを集計し、GitHub の issue にスコアと一緒に記録して、次に直すべき箇所を特定する。

## 前提

| 項目 | 値 |
| --- | --- |
| リポジトリ | `karamaru-alpha/isucon2026`（記録先は issue #1） |
| ssh ホスト | `isu1` `isu2` `isu3` |
| サーバー上の作業ディレクトリ | `/home/isucon/webapp` |
| `make slow` | `slow.sh` 経由で `pt-query-digest`（上位 6 クエリ） |
| `make kataru` | nginx の access.log を `kataribe` で集計 |
| `./analyze.sh <タイトル> <スコア>` | 上の 2 つを実行し、結果を `gh issue comment 1` で投稿 |

`/isucon-deploy` でログをローテートしてからベンチを 1 回だけ回した状態で実行する。ローテート後に複数回ベンチを回すと集計が混ざる。

## 引数

- タイトル（何を変えたか。例: `usersにindex追加`）
- スコア（ポータルのベンチ結果）
- `N`: 集計するサーバー。省略したら `1`

タイトルかスコアがない場合は、ユーザーに聞く。スコアは推測で埋めない。

## 手順

### 1. issue #1 の存在確認（初回だけ）

```sh
gh issue view 1 -R karamaru-alpha/isucon2026 --json number,title
```

なければ `gh issue create -R karamaru-alpha/isucon2026 --title "計測ログ" --body "ベンチごとの kataribe / slow-log"` で作成する（番号が 1 になることを確認する）。

### 2. 集計結果の取得（分析用）

```sh
ssh isu${N} "cd /home/isucon/webapp && make kataru"
ssh isu${N} "cd /home/isucon/webapp && make slow"
```

2 つは並列に実行してよい。出力が空なら、ログ設定（nginx の `log_format` / MySQL の `slow_query_log`）が有効になっているか、ベンチ後にローテートしていないかを確認する。

DB を別サーバーに分けている場合、`make slow` は DB サーバーで実行する。

### 3. issue へ投稿

```sh
ssh isu${N} "cd /home/isucon/webapp && ./analyze.sh '<タイトル>' <スコア>"
```

- タイトルはシングルクォートで囲む（タイトル中の `'` は `'\''` にエスケープする）。
- サーバー上の `gh` が未ログインで失敗したら、サーバーで `gh auth login` するようユーザーに案内する。
- `gh issue comment` が出力したコメント URL を控える。
- DB が別サーバーの場合、analyze.sh の slow-log 欄は空になる。手順 2 で DB サーバーから取得した結果を `gh issue comment 1 -R karamaru-alpha/isucon2026 --body-file <file>` で追記する。

### 4. ボトルネックの要約

手順 2 の出力を読み、次の観点でまとめる。

**kataribe（アクセスログ）**

- `Sum`（合計応答時間）の大きい順に上位 3〜5 エンドポイントを挙げる。Count が多いのか、1 回が遅いのか（Avg / Max）を区別する。
- 静的ファイルが上位にいるなら nginx 配信やキャッシュヘッダの余地がある。

**pt-query-digest（スロークエリ）**

- `Rank` 上位（総実行時間順）のクエリについて、`Calls`、`R/Call`、`Rows examine` と `Rows sent` の比を確認する。
  - examine ≫ sent → インデックス不足の可能性が高い
  - Calls が極端に多い → N+1 の可能性が高い
- 上位クエリをローカルの `~/github.com/karamaru-alpha/isucon2026` で grep し、呼び出し元のハンドラーと結び付ける。

**出力フォーマット**

```
## 計測結果: <タイトル> / <スコア>
コメント: <URL>

### 遅いエンドポイント
1. GET /api/xxx  Sum 120s (Count 3000, Avg 0.04s)
...

### 遅いクエリ
1. SELECT ... FROM xxx WHERE ...  全体の45% / Calls 30000 → N+1 疑い (handler: getXxx)
...

### 次に直す候補
1. <施策> — <根拠>（期待効果: 大/中/小）
2. ...
```

前回の計測コメントがあれば `gh api repos/karamaru-alpha/isucon2026/issues/1/comments --jq '.[-2].body'` で取得して比較し、スコアと上位項目の変化も添える。

最後に「`/isucon-tunning` で 1 番目の候補から直すのだ」と次の行動を案内する。

---
name: isucon-tunning
description: ISUCONで計測結果(kataribe / pt-query-digest / pprof)から最大のボトルネックを1つ選び、karamaru-alpha/isucon2026 のGoアプリ・MySQL・nginxをチューニングする。「チューニングして」「ボトルネックを直して」「スコアを上げて」「isucon-tunning」などの言葉が出たら使う。
---

# isucon-tunning

計測データをもとにボトルネックを 1 つずつ潰す。サイクルは次のとおり。

```
/isucon-survey で計測 → ボトルネック特定 → 根拠と改善内容を提案して許可を取る → 1 施策だけ実装 → /pr で draft PR 作成（ここで止まる）
→ PR をマージ → /isucon-deploy → ベンチ → /isucon-survey → 改善なら次へ / 悪化なら revert
```

## 前提

| 項目 | 値 |
| --- | --- |
| ローカルリポジトリ | `~/github.com/karamaru-alpha/isucon2026` |
| 計測ログ | `karamaru-alpha/isucon2026` の issue #1 のコメント |
| ssh ホスト | `isu1` `isu2` `isu3`（作業ディレクトリ `/home/isucon/webapp`） |
| 解法集 | `references/` 配下（isucon-memo の README.md と main.go から抜き出したもの） |

## references

| ファイル | 内容 |
| --- | --- |
| [patterns.md](references/patterns.md) | 症状 → 施策 → 解法の対応表と、よくある事故。最初に読む |
| [go.md](references/go.md) | 汎用キャッシュ `Cacher[T]`、オンメモリ戦略（map / slice / 期限付き）、singleflight、シリアライザ切り替え、http 設定、画像のファイル化、定期処理、UDS、pprof |
| [mysql.md](references/mysql.md) | スロークエリ設定、DB 分離（別インスタンス・外部接続・ユーザー作成）、MariaDB への乗り換え、bulk insert、IN 句、コネクションプール、起動待ち、generated column、UUID の BINARY 化、upsert、trigger、グループ最新行、件数カラム、アドバイザリロック |
| [nginx.md](references/nginx.md) | アクセスログ設定、upstream keepalive、fd 上限、静的ファイルキャッシュ、proxy_cache、X-Accel-Redirect、メソッド / パス / クエリでの振り分け、bot 拒否、gzip / 事前圧縮 |
| [linux.md](references/linux.md) | fd 上限、systemd のユニット定義、再起動試験 |
| [dns.md](references/dns.md) | DNS が出題範囲のとき: PowerDNS の TTL・キャッシュ、dnsdist での水責め対策、DNS サーバー自作 |
| [strategy.md](references/strategy.md) | マニュアルから緩めてよい箇所を探す観点、事前計算、複数台構成の決め方 |

各ファイルの末尾（`<!-- ここから下は ISUCON10〜14 の…調査から追記したもの -->` 以降）と dns.md / strategy.md は、過去回の上位解法・公式講評の調査から追加した解法。patterns.md では `[調査]` 付きで載せている。

## 原則

1. **計測してから直す。** 勘で直さない。根拠になる計測値を必ず示す。
2. **1 ベンチにつき 1 施策。** 複数を同時に入れると、どれが効いたか分からなくなる。
3. **仕様を壊さない。** レスポンスの形・順序・ステータスコードを変えない。速くてもベンチが FAIL すれば 0 点になる。
4. **再起動後も動くこと。** 最終スコアは再起動試験の後に確定する。インデックスやキャッシュの初期化は、`/initialize` かアプリ起動時に必ず実行されるようにする。
5. **ボトルネックは移動する。** 大きく改善したら、必ず計測し直してから次に進む。

## 手順

### 1. 計測データを集める

- 直前に `/isucon-survey` を実行していれば、その結果を使う。
- なければ issue の最新コメントを取得する:
  ```sh
  gh api repos/karamaru-alpha/isucon2026/issues/1/comments --jq '.[-1].body'
  ```
- どの層が詰まっているか分からない場合は、ベンチ中のリソースを確認する（ユーザーにベンチを回してもらいながら実行する）:
  ```sh
  ssh isu1 "top -bn1 | head -20"
  ```
  - `mysqld` が CPU を使い切っている → DB 側（インデックス・N+1・クエリ削減）
  - アプリが CPU を使い切っている → Go 側（pprof でホットスポットを確認。[references/go.md](references/go.md) の「Go - pprof」参照）
  - どちらにも余裕があるのに遅い → ロック待ち・外部呼び出し・コネクションプール・nginx 設定

### 2. ボトルネックを 1 つ選ぶ

「合計時間が大きく、かつ直すコストが小さい」ものを選ぶ。候補が複数あるなら、期待効果と作業量を添えて推奨を 1 つ示す。

計測上の上位が直しにくいときや伸び悩んだときは、[references/strategy.md](references/strategy.md) の観点でマニュアルを読み直し、仕様上緩めてよい箇所（反映の猶予、キャッシュ可否、リクエスト間隔の指定など）がないか探す。

### 3. コードを読む

1 施策 = 1 PR にするため、最新の main から作業を始める。未コミットの差分があれば、何の変更かユーザーに確認してから進める。

```sh
cd ~/github.com/karamaru-alpha/isucon2026
git switch main && git pull --ff-only
grep -rn "<クエリの特徴的な部分>" --include=*.go .
```

- 該当ハンドラーの全体を読み、ループ内のクエリ、トランザクション範囲、ロックを把握する。
- スキーマは `make schema`、データ量は `make data` をサーバーで実行して確認する。
- インデックスを検討するときは、サーバーで `EXPLAIN` を実行する:
  ```sh
  ssh isu1 "cd /home/isucon/webapp && echo 'EXPLAIN <query>' | make sql"
  ```

### 4. 施策を提案して許可を取る

[references/patterns.md](references/patterns.md) で症状に合う施策を選び、「解法」列が指す reference ファイルの節を読む。そのうえで、**実装に入る前に**次の形でユーザーに提案し、進めてよいか確認する。許可が出るまでコードを書き換えない。

```
## ボトルネック
<エンドポイント / クエリ / 関数>

## 根拠となる指標
- kataribe: GET /api/xxx が Sum 120s（全体の 35%）で 1 位、Count 3000、Avg 40ms
- pt-query-digest: `SELECT ... FROM xxx WHERE ...` が Total 80s、Rows examine 50 万 / sent 10
- EXPLAIN: type ALL、key NULL
（pprof・top など、判断に使った計測値をそのまま引用する。推測で数値を書かない）

## 原因の推測
<コードを読んで分かったこと。例: ループ内で 1 件ずつ SELECT している（N+1）、インデックスがない>

## 実装改善内容
- 変更するファイル・関数: `webapp/go/xxx.go` の `getXxx`
- 変更内容: <具体的に。例: (user_id, created_at) の複合インデックスを 0_Schema.sql に追加>
- 参考にする解法: patterns.md → mysql.md「〇〇」
- 再起動・/initialize 時の扱い: <インデックス / キャッシュの初期化経路>

## 期待効果とリスク
- 期待効果: <例: このクエリの Total 80s がほぼ 0 になる見込み>
- リスク: <例: 書き込みが少し遅くなる、キャッシュ不整合で FAIL する可能性>

## 他の候補
- <2 番手の候補と、今回選ばなかった理由>（あれば）
```

- AskUserQuestion ツールで「この施策で進める / 別の候補にする / 中止」を選んでもらう。別の候補を選ばれたら、その候補で提案し直す。
- ユーザーが修正を指示したら、それを反映した内容で実装する。

### 5. 実装する

許可された施策だけを実装する。スニペットは雛形なので、対象アプリの型名・テーブル名・ライブラリに合わせて書き直す。提案と違う変更が必要になったら、勝手に広げずにもう一度確認する。

実装時の注意:

- インデックスは、アプリの初期化で流れる SQL（`webapp/sql/` 配下のスキーマや `init.sh` から呼ばれるもの）に追加する。`mysql` で手で `ALTER` しただけだと、`/initialize` や再起動で消えることがある。
- オンメモリキャッシュを入れたら、`/initialize` でクリア・再構築する。また、書き込みが発生するすべての経路で更新・無効化する。
- アプリサーバーを複数台にしている場合、プロセス内キャッシュは台ごとに食い違う。どの台に書き込みが来ても整合するかを確認する。
- 既存コードのスタイル（sqlx / echo などのライブラリ、命名）に合わせる。

### 6. ローカルで検証する

```sh
cd ~/github.com/karamaru-alpha/isucon2026/<go のディレクトリ>
go build ./... && go vet ./...
```

ビルドが通らない状態ではコミットしない。

### 7. /pr で PR を作成して止まる

改善が終わったら、Skill ツールで `pr` スキルを呼び出して draft PR を作成する。ブランチ作成（main 上なら `perf/karamaru/<施策名>`）・コミット・push は `pr` スキルが行う。

- 1 施策 = 1 PR。別の施策の変更を混ぜない。
- 変更種別は `perf` にする。
- PR 本文の「ポイント」には次を書く:
  - 根拠の計測値（例: `GET /api/xxx` が Sum 120s で 1 位、該当クエリが examine 50 万行 / sent 10 行）
  - 施策の内容と期待効果
  - 再起動・`/initialize` 時の扱い（インデックスやキャッシュの初期化経路）
- PR の URL を報告したら、そこで止まる。デプロイやベンチには進まない。
- 次の行動として「PR をマージしたら `/isucon-deploy` → ベンチ → `/isucon-survey` で効果を確認するのだ」と案内する。
- ベンチ後にスコアが下がった、または FAIL した場合は、その場でデバッグせずに `git revert` してデプロイし直し、原因を落ち着いて調べる。

## 終盤（残り 1 時間を切ったら）

新しい施策は入れず、安定化に切り替える。

- [ ] MySQL のスロークエリログを無効化する（`slow_query_log = OFF`）
- [ ] nginx の `access_log off`
- [ ] Go のデバッグログと pprof を無効化する
- [ ] アプリ・MySQL・nginx が `systemctl is-enabled` で enabled になっている
- [ ] 全台で `sudo reboot` → ベンチが通ることを確認する
- [ ] 起動時にキャッシュがウォームアップされる（初回リクエスト時の遅延読み込みにしない）
- [ ] `git status` がクリーンで、最終コミットが push 済み

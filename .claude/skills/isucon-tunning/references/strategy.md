# 定石・考え方

> ISUCON10〜14 の上位解法・公式講評の調査から追記したもの。ボトルネックを選ぶ前・行き詰まったときに、マニュアルと照らし合わせて確認する観点。

## マニュアルから「緩めてよい箇所」を探す

- 「N秒以内に反映されればよい」→ 非同期化・バルク化・短TTLキャッシュの対象 (11予選 isu_condition, 14 chair/coordinate は3秒以内)
- 「最大1秒古くてよい」→ ticker で事前生成してそのまま返す (10本選 ダッシュボード)
- ベンチが `Cache-Control` / `ETag` / 条件付きGETを尊重するか → 尊重するならヘッダを付けるだけでリクエストが減る (10本選, 11予選, 13)
- レスポンスで次のリクエスト間隔を指定できるか (`retry_after_ms`, `Retry-After`) → 状態に応じて伸ばす (14)
- 意図的な負荷調整ノブ(drop確率, 429 Retry-After) → 定数化して処理能力に合わせて調整する (11予選 dropProbability, 12予選)
- 送り先URLを外部から指定できる仕様(`POST_ISUCONDITION_TARGET_BASE_URL`など) → 書き込み専用サーバーを作れる (11予選)
- 外部APIの冪等キー(`Idempotency-Key`) → 確認APIを省いて単純リトライできる (14 決済)

cf. https://isucon.net/archives/56044867.html (11予選講評), https://isucon.net/archives/58869617.html (14講評)

## 確定したら不変なデータは確定時に計算する

- 終了した大会の請求額、ランキング、累計距離など「書き込み時にしか変わらない値」は読み込み時に集計しない
- 書き込み時に集計テーブル/カラムへ反映し、読み込みは SELECT 1回にする (10本選 leaderboard, 12予選 billing/ranking, 13 total_tips, 14 total_distance)
- 「最新1件」「初回1件」しか使わない履歴テーブルは行ごと消す・縮める (12予選 visit_history 約320万行→約20万行)

cf. https://isucon.net/archives/56842718.html (12予選 Node.jsでの解説)

## 複数台構成の決め方

- JOINとトランザクションの依存関係を最初に確認する。依存がなければテーブル単位でDBを分けられる (10予選 chair / estate)
- ユーザー同士が絡まなければユーザーID単位で水平分割できる (11予選, 12本選)
- 書き込みが多い回は App1台 + DB複数台 が公式想定のことが多い
- 複数台にしたら `/initialize` を全台に配る・アプリ起動時にDBをPingで待つ・`StartLimitBurst` を上げる (再起動試験対策)
- インスタンス間のMTUを9000にしたら MySQL が `Bad handshake` になった例あり。いじらない

cf. https://isucon.net/archives/55025156.html (10予選講評), https://isucon.net/archives/56959385.html (12本選講評)

## アクセスログでクエリパラメータの組み合わせを数える

検索APIのインデックス設計は「実際に来ている条件の組み合わせ」から決める (10予選)

```sh
# /api/estate/search のクエリ文字列のキーの組み合わせを集計
grep 'GET /api/estate/search' /var/log/nginx/access.log \
  | awk '{print $7}' | sed -E 's/^[^?]*\??//; s/=[^&]*//g' | sort | uniq -c | sort -rn
```

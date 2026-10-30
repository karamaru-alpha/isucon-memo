# 症状別チューニングパターン

「解法」列は、同じディレクトリにある reference ファイルの節名を指す。実装する前に該当する節を読むこと。

- [go.md](go.md): Go（キャッシュ、singleflight、シリアライザ、UDS、pprof など）
- [mysql.md](mysql.md): MySQL（bulk insert、IN 句、upsert、DB 分離、UUID の BINARY 化など）
- [nginx.md](nginx.md): nginx（keepalive、静的配信、proxy_cache、X-Accel-Redirect、振り分け、圧縮など）
- [linux.md](linux.md): Linux（fd 上限、systemd）

## DB

| 症状 | 施策 | 解法 |
| --- | --- | --- |
| `Rows examine` ≫ `Rows sent`、EXPLAIN が `type: ALL` | WHERE の等値条件 → 範囲・ORDER BY の順で複合インデックスを張る | mysql.md「generatedColumns」（式を index 化したいとき） |
| `Using filesort` | ORDER BY の列までインデックスに含める | — |
| 同じ形のクエリの Calls が異常に多い（N+1） | `IN (?)` でまとめて取得して map で突き合わせる、または JOIN | mysql.md「IN句」 |
| ループ内の INSERT | バルク INSERT | mysql.md「bulkInsert」 |
| SELECT してから INSERT / UPDATE | `INSERT ... ON DUPLICATE KEY UPDATE` | mysql.md「Upsert」 |
| グループごとの最新 1 件を取得している | サブクエリや最新値テーブルで解消する | mysql.md「Group毎に最新のレコードをSELECTする」 |
| `COUNT(*)` が重い | 件数カラムを持たせて更新時に加算する | mysql.md「個テーブルの個数をサブクエリで持ってきてUpdate」 |
| JSON や式の検索 | generated column + インデックス | mysql.md「generatedColumns」 |
| UUID 文字列の主キーが大きい | `BINARY(16)` などへ縮小する | mysql.md「UUIDをBINARY(16)で格納する」 |
| デッドロック・ロック待ち | トランザクション範囲を縮める、ロック順をそろえる | mysql.md「デッドロックの調査」 / mysql.md「アドバイザリロック」 |
| DB の CPU が張り付いたまま | DB を別インスタンスに分離する | mysql.md「DB用のインスタンスのセットアップ」 |
| `Too many connections` / 接続待ち | `SetMaxOpenConns` / `SetMaxIdleConns` を調整する | mysql.md「コネクションプール」 |

## Go

| 症状 | 施策 | 解法 |
| --- | --- | --- |
| 更新の少ないマスタを毎回 SELECT している | 起動時と `/initialize` でオンメモリに載せる | go.md「汎用キャッシュ Cacher[T]（isucon-memo/main.go）」 / go.md「オンメモリ戦略」 |
| 同じキーへの重い処理が同時に走る | singleflight | go.md「SingleFlight」 |
| JSON エンコードが pprof 上位 | シリアライザを切り替える | go.md「シリアライザを切り替える」 |
| 画像を DB の BLOB から返している | ファイルに書き出して nginx から配信する | go.md「画像をfileで読み書き」 |
| 外部 HTTP 呼び出しが遅い・ポート枯渇 | `http.Client` を使い回し、`MaxIdleConnsPerHost` を広げる | go.md「httpを広げる」 / go.md「httpリクエスト」 |
| 即時に反映しなくてよい集計処理 | ticker で定期的にまとめて処理する | go.md「一定時間毎に処理をする」 |
| nginx ↔ アプリ間の TCP オーバーヘッド | Unix domain socket | go.md「UnixDomainSocket」 |

## nginx / OS

| 症状 | 施策 | 解法 |
| --- | --- | --- |
| 静的ファイルが kataribe の上位にいる | nginx から直接配信し、`expires` を付ける | nginx.md「静的ファイルのクライアントキャッシュ」 |
| 同一レスポンスを何度も返している | proxy_cache | nginx.md「レスポンスキャッシュ(ProxyCache)」 |
| 認証付きのファイル配信 | X-Accel-Redirect | nginx.md「認証付きの静的配信(X-Accel-Redirect)」 |
| upstream への接続を毎回張り直している | upstream keepalive | nginx.md「keepaliveを有効する」 |
| `Too many open files` | ファイルディスクリプタの上限を上げる | nginx.md「ファイル上限を確認・拡張する」 / linux.md「ファイルディスクリプタ上限up」 |
| 転送量が大きい | gzip / 事前圧縮 | nginx.md「圧縮 / 事前圧縮」 |

## よくある事故

- キャッシュを `/initialize` でクリアしていないため、前回ベンチのデータが残って FAIL する。
- 複数台構成で、プロセス内キャッシュが台ごとに食い違う。
- キャッシュにポインタを入れ、呼び出し側が書き換えて汚染される。値で返すか、変更しない前提を守る。
- `/initialize` 内で DROP / CREATE TABLE していて、手で張ったインデックスが消える。
- `innodb_flush_log_at_trx_commit` などの永続性に関わる設定は、レギュレーションを確認してから変える。

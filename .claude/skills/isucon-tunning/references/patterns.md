# 症状別チューニングパターン

「解法」列は、同じディレクトリにある reference ファイルの節名を指す。実装する前に該当する節を読むこと。

- [go.md](go.md): Go（キャッシュ、singleflight、シリアライザ、UDS、pprof など）
- [mysql.md](mysql.md): MySQL（bulk insert、IN 句、upsert、DB 分離、UUID の BINARY 化など）
- [nginx.md](nginx.md): nginx（keepalive、静的配信、proxy_cache、X-Accel-Redirect、振り分け、圧縮など）
- [linux.md](linux.md): Linux（fd 上限、systemd、再起動試験）
- [dns.md](dns.md): DNS（PowerDNS のキャッシュ、dnsdist、DNS サーバー自作）
- [strategy.md](strategy.md): 定石・考え方（マニュアルから緩めてよい箇所を探す、事前計算、複数台構成の決め方）

各 reference の「ここから下は ISUCON10〜14 の…調査から追記したもの」より下と、dns.md / strategy.md は過去回の上位解法の調査から追加したもの。表では解法列に `[調査]` を付けて区別する。

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
| PREPARE / `Close stmt` が pt-query-digest 上位 | DSN に `interpolateParams=true` | [調査] mysql.md「interpolateParams」 |
| 書き込みが多く DB の I/O 待ちが大きい | binlog 無効化・flush を緩める・doublewrite 無効化 | [調査] mysql.md「書き込み系のmy.cnf」 |
| 書き込み時に外部キーのチェックが重い | 外部キーを外して通常インデックスにする | [調査] mysql.md「外部キー制約を外す」 |
| AUTO_INCREMENT の id + 別インデックスで引いている時系列テーブル | 自然キーを複合主キーにする | [調査] mysql.md「自然キーを主キーにする」 |
| `ORDER BY a DESC, id ASC` で filesort | 降順インデックス（5.7 / MariaDB は符号反転カラム） | [調査] mysql.md「降順インデックス」 |
| 関数・LIKE などの重い条件 + ORDER BY LIMIT | サブクエリで並べてから重い条件を評価する | [調査] mysql.md「重い条件はORDER BY LIMITの後に評価させる」 |
| 決まった区分での範囲検索 | 区分 ID カラムにして等値条件にする | [調査] mysql.md「範囲条件を区分IDにする」 |
| 緯度経度の多角形検索・N+1 | POINT + SPATIAL INDEX | [調査] mysql.md「空間インデックス」 |
| SELECT FOR UPDATE → UPDATE、読んで足して書き戻す | 条件付き UPDATE + RowsAffected、相対更新、INSERT ... SELECT | [調査] mysql.md「読んで書くをUPDATE1発にする」 |
| SELECT だけのハンドラで COMMIT が上位 | 不要なトランザクションを外す | [調査] mysql.md「不要なトランザクションを外す」 |
| `deleted_at IS NULL` 付きの検索 | 物理削除にする | [調査] mysql.md「論理削除をやめる」 |
| 大きな OFFSET | keyset pagination | [調査] mysql.md「OFFSETをやめる(keyset pagination)」 |
| 履歴テーブルを LAG / ROW_NUMBER で集計している | 最新値・累積値・状態フラグを親テーブルに持つ | [調査] mysql.md「最新値・累積値を親テーブルに持つ」 |
| ランキング・集計を読み込み時に毎回計算 | 書き込み時に計算して保存する | [調査] mysql.md「書き込み時にランキングを作る」 / strategy.md「確定したら不変なデータは確定時に計算する」 |
| `REPLACE INTO id_generator` などの採番テーブル | UUID / AUTO_INCREMENT / Snowflake | [調査] go.md「ID採番をアプリでする」 |
| SQLite など別エンジンのDBが混ざっている | MySQL に移す | [調査] mysql.md「SQLiteからMySQLへ移す」 |
| DB 1 台では書き込みが捌けない（ユーザー間の依存なし） | ユーザー ID 単位のシャーディング | [調査] go.md「ユーザーID単位のシャーディング」 / strategy.md「複数台構成の決め方」 |

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
| 大きなボトルネックを潰し終えて、Go の CPU 使用率が高い | `GOAMD64=v3` でビルドする（効果は数 % 程度か誤差。必ずベンチで確かめる） | go.md「CPU の命令セットレベルを上げる（GOAMD64）」 |
| echo の Logger / Debug、stdout への大量ログ | ログとデバッグモードを止める | [調査] go.md「デバッグモード・ログ出力を止める」 |
| 高頻度 POST の INSERT が重い（即時に読まれない） | channel に溜めて非同期バルク INSERT | [調査] go.md「非同期でバルクINSERTする」 / strategy.md「マニュアルから「緩めてよい箇所」を探す」 |
| 引数が定数の重い集計が毎回走る | 短 TTL キャッシュ + singleflight | [調査] go.md「マイクロキャッシュ(短TTL + SingleFlight)」 |
| 「N 秒古くてよい」レスポンスを毎回生成・圧縮 | ticker で圧縮済みバイト列を作っておく | [調査] go.md「圧縮済みレスポンスを事前生成する」 |
| ポーリング・ロングポーリングが多い | `retry_after_ms` を伸ばす、channel 通知、SSE | [調査] go.md「状態変化をchannelで通知する」 / go.md「Server-Sent Events」 |
| flock・アプリ全体の Mutex で直列化 | キー単位のロック、DB トランザクション | [調査] go.md「キー単位のロック」 |
| 画像などを毎回フルで返している | ETag / If-None-Match で 304 | [調査] go.md「ETag / If-None-Matchで304を返す」 |
| `exec.Command` で zip / cp などを呼んでいる、`io.ReadAll` で大きなファイル | `archive/zip` の Store、ハードリンク、`io.Copy` | [調査] go.md「外部コマンドを呼ばない・ファイルはストリームで扱う」 |
| pprof で mallocgc / GC が上位 | sync.Pool、GOGC | [調査] go.md「アロケーション・GCを減らす」 |
| 大きな施策を終えて Go の CPU が高い | PGO ビルド | [調査] go.md「PGO」 |
| スロークエリがどのハンドラのものか分からない | SQL に呼び出し元コメントを付ける | [調査] go.md「SQLに呼び出し元をコメントで付ける」 |

## nginx / OS

| 症状 | 施策 | 解法 |
| --- | --- | --- |
| 静的ファイルが kataribe の上位にいる | nginx から直接配信し、`expires` を付ける | nginx.md「静的ファイルのクライアントキャッシュ」 |
| 同一レスポンスを何度も返している | proxy_cache | nginx.md「レスポンスキャッシュ(ProxyCache)」 |
| 認証付きのファイル配信 | X-Accel-Redirect | nginx.md「認証付きの静的配信(X-Accel-Redirect)」 |
| upstream への接続を毎回張り直している | upstream keepalive | nginx.md「keepaliveを有効する」 |
| `Too many open files` | ファイルディスクリプタの上限を上げる | nginx.md「ファイル上限を確認・拡張する」 / linux.md「ファイルディスクリプタ上限up」 |
| 転送量が大きい | gzip / 事前圧縮 | nginx.md「圧縮 / 事前圧縮」 |
| SPA のパスがアプリに来ている | nginx で `try_files $uri /index.html` + expires | [調査] nginx.md「SPAのindex.htmlを返す」 |
| error.log に `client request body is buffered to a temporary file` | `client_body_buffer_size`、`proxy_request_buffering off` | [調査] nginx.md「リクエストボディのバッファ」 |
| 高負荷で HTTP/2 の接続エラー | `keepalive_requests`（旧 `http2_max_requests`）を上げる | [調査] nginx.md「1接続あたりのリクエスト上限」 |
| マルチテナントで複数台に分けたい | Host（サブドメイン）で振り分け | [調査] nginx.md「Hostで振り分ける」 |
| nginx の TLS で CPU が張り付く | `stream` で 443 を L4 分散 | [調査] nginx.md「TLS終端を分散する」 |
| DNS（PowerDNS）がボトルネック・水責め | TTL / キャッシュ、dnsdist、DNS 自作 | [調査] dns.md |
| 再起動試験でアプリが起動しない | `StartLimitBurst`、DB の起動待ち | [調査] linux.md「再起動試験でsystemdが止まらないようにする」 |

## よくある事故

- キャッシュを `/initialize` でクリアしていないため、前回ベンチのデータが残って FAIL する。
- 複数台構成で、プロセス内キャッシュが台ごとに食い違う。
- キャッシュにポインタを入れ、呼び出し側が書き換えて汚染される。値で返すか、変更しない前提を守る。
- `/initialize` 内で DROP / CREATE TABLE していて、手で張ったインデックスが消える。
- `innodb_flush_log_at_trx_commit` などの永続性に関わる設定は、レギュレーションを確認してから変える。

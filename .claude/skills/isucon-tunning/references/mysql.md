# MySQL (MariaDB) の解法

> isucon-memo の README.md から抜き出したもの。スニペットは雛形なので、そのまま貼らずに対象アプリの型名・テーブル名・ライブラリに合わせて書き直すこと。

## Mysql - スロークエリ

- 設定ファイルをホームディレクトリに持ってくる (`mysql --help | grep my.cnf`で探せる)
```sh
cp /etc/mysql/my.cnf my.cnf
```

- スロークエリを吐くように[my.cnf](./etc/my.cnf)設定する

```sh
sudo touch /var/log/mysql/slow-query.log
sudo chown -R mysql /var/log/mysql/slow-query.log
```

```conf
[mysqld]
log_error = /var/log/mysql/error.log
slow_query_log_file = /var/log/mysql/slow-query.log
slow_query_log = ON
long_query_time = 0.0
log_output = FILE
# general_log = OFF 計測が終わったら上記をコメントアウトしこの行を追加
skip-log-bin
```

```
sudo systemctl restart mysql
```

## DB用のインスタンスのセットアップ

```
sudo -u -i isucon
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install -y percona-toolkit gh
gh auth login
cd ./webapp
git init
git config --global user.name karamaru-alpha
git config --global user.email mrnk3078@gmail.com
git config --global pull.rebase false
git config credential.helper store
git remote add origin git@github.com:karamaru-alpha/isucon12-q.git
git fetch origin main && git reset --hard origin/main
git branch -M master main	
```

```
sudo touch /var/log/mysql/slow-query.log
sudo chown -R mysql /var/log/mysql/slow-query.log
```


## 外部からのアクセスを許容する

- goのmysql.openとsqlのinit.shもHOSTをプライベートアドレスに変更する
- '%'のユーザー作成も忘れずに！

```sh
# MariaDBの場合
/etc/mysql/mariadb.conf.d/*-server.cnf
bind-address = 0.0.0.0
```

## ユーザの作成
```sql
SELECT user, host FROM mysql.user;
DROP USER 'isucon'@'%';
CREATE USER 'isucon'@'%' IDENTIFIED BY 'isucon';
GRANT ALL PRIVILEGES ON * . * TO 'isucon'@'%';
FLUSH PRIVILEGES;
```

## MysqlからMariaDBに乗り換える

```bash
## delete
sudo apt-get remove --purge mysql-server* mysql-common
sudo rm -r /etc/mysql
sudo rm -r /var/lib/mysql
sudo apt autoremove -y
sudo apt clean
sudo aa-remove-unknown

# install
sudo apt install apt-transport-https
sudo curl -LsS -O https://downloads.mariadb.com/MariaDB/mariadb_repo_setup
sudo bash mariadb_repo_setup --mariadb-server-version=10.9
sudo rm mariadb_repo_setup
sudo apt update
sudo apt install mariadb-server mariadb-common -y
sudo mysqld --version
```

## MariaDBを最新にする

```bash
## delete
sudo rm -r /etc/ld.so.conf.d/mysql /etc/mysql/my.cnf /usr/local/mysql /var/log/mysql/error.log
sudo apt purge mariadb-server
sudo apt autoremove -y
sudo apt clean
sudo aa-remove-unknown

# install
sudo apt install apt-transport-https
sudo curl -LsS -O https://downloads.mariadb.com/MariaDB/mariadb_repo_setup
sudo bash mariadb_repo_setup --mariadb-server-version=10.9
sudo rm mariadb_repo_setup
sudo apt update
sudo apt install mariadb-server -y
sudo mysqld --version
```

## デッドロックの調査
```sql
SHOW ENGINE INNODB STATUS;
```

## TroubleShoot

- Unknown collation: 'utf8mb4_0900_ai_ci'
  - sed -i 's/utf8mb4_0900_ai_ci/utf8mb4_unicode_ci/g' sql/dump.sql

## Mysqlの起動を待つ

```go
func main() {
	// db.Open　の直後
	for {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}
}
```

## bulkInsert

```go
func bulkInsert(isuList []Isu) {
	args := make([]interface{}, 0, len(isuList)*3)
	placeHolders := &strings.Builder{}
	for i, v := range isuList {
		args = append(args, v.Col1, v.Col2, v.Col3)
		if i == 0 {
			placeHolders.WriteString(" (?, ?, ?)")
		} else {
			placeHolders.WriteString(",(?, ?, ?)")
		}
	}
	_, err = db.Exec("INSERT INTO table_name(col_1, col_2, col_3) VALUES" + placeHolders.String(), args...)
}

// sqlxの場合
// db.NamedExec("INSERT INTO `latest_isu_level` (`jia_isu_uuid`, `level`) VALUES (:jia_isu_uuid, :level)", latestIsuLevels)
```

## IN句

```go
type Isu struct {
	Col1 int `db:"col_1"`
	Col2 int `db:"col_2"`
	Col3 int `db:"col_3"`
}

// Col1の値を複数条件で検索
func in(col1s []int)　[]Isu {
	var isuList []Isu
	inPlaceHolders := "col_1 IN (?" + strings.Repeat(",?", len(levels)-1) + ")" // n=0の時がある場合は分岐が必要
	db.Select(&isuList, `SELECT * FROM isu WHERE ` + inPlaceHolders, col1s...)
}

// sqlxの場合
// query, params, err := sqlx.In(`SELECT * FROM users WHERE id IN (?)`, []int{1,2})
// db.Select(&isuList, db.Rebind(query), params...)
```

## コネクションプール

```go
func main() {
	const SQL_CONN_COUNT = 20
	// 最大接続数
	db.SetMaxOpenConns(SQL_CONN_COUNT)
	// プールできるコネクションの数
	db.SetMaxIdleConns(SQL_CONN_COUNT)
	// 接続が確立されてからコネクションを保持できる最大時間
	db.SetConnMaxLifetime(SQL_CONN_COUNT * time.Second)
	defer db.Close()
}
```

## DBの起動を待つ

```go
func main() {
	db := openDB()
	for {
        if err := db.Ping(); err == nil {
            break
        }
        time.Sleep(time.Second * 1)
    }
```

## generatedColumns

```sql
popularity INT NOT NULL,
popularity_desc INT AS (-popularity) INVISIBLE, -- index 使う時は STORED
```


## 1byte長の半角文字列をピッタリ格納する

```sql
`id` CHAR(26) CHARACTER SET latin1,
```


## UUIDをBINARY(16)で格納する

```sql
-- 2_Patch.sql
DROP FUNCTION IF EXISTS UUID_TO_BIN;

CREATE FUNCTION UUID_TO_BIN(_uuid BINARY(36)) RETURNS BINARY(16) LANGUAGE SQL DETERMINISTIC CONTAINS SQL SQL SECURITY INVOKER RETURN UNHEX(
  CONCAT(
    SUBSTR(_uuid, 15, 4),
    SUBSTR(_uuid, 10, 4),
    SUBSTR(_uuid, 1, 8),
    SUBSTR(_uuid, 20, 4),
    SUBSTR(_uuid, 25)
  )
);

DROP FUNCTION IF EXISTS BIN_TO_UUID;

CREATE FUNCTION BIN_TO_UUID(_bin BINARY(16)) RETURNS BINARY(36) LANGUAGE SQL DETERMINISTIC CONTAINS SQL SQL SECURITY INVOKER RETURN LCASE(
  CONCAT_WS(
    '-',
    HEX(SUBSTR(_bin, 5, 4)),
    HEX(SUBSTR(_bin, 3, 2)),
    HEX(SUBSTR(_bin, 1, 2)),
    HEX(SUBSTR(_bin, 9, 2)),
    HEX(SUBSTR(_bin, 11))
  )
);

ALTER TABLE user ADD COLUMN bin_uuid BINARY(16);

UPDATE user SET bin_uuid = UUID_TO_BIN(uuid);

-- uuid が PK の場合: ALTER TABLE user DROP PRIMARY KEY, ADD PRIMARY KEY (`bin_uuid`);

ALTER TABLE user DROP COLUMN uuid;

ALTER TABLE user RENAME COLUMN bin_uuid TO uuid;

-- NOT NULL などの制約があった場合は付け直す: ALTER TABLE user MODIFY COLUMN uuid BINARY(16) NOT NULL UNIQUE;
```
```go
func get(uuid string) {
	var user User
	db.Get(&user, "SELECT BIN_TO_UUID(`uuid`) FROM user WHERE uuid = UUID_TO_BIN(?)", uuid)
}
```

## 水平分割

cf. https://github.com/narusejun/isucon11-qualify/commit/fda74ca7e56b70a58a7a49c773cb892d3dae6765

## Upsert

```sql
INSERT INTO `user` (`id`, `name`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `name`=?;
-- bulk
INSERT INTO `user` (`id`, `name`) VALUES (:id, :name) ON DUPLICATE KEY UPDATE `name`=VALUES(`name`);
```

## trigger

```sql
DROP TRIGGER IF EXISTS tr1;
CREATE TRIGGER tr1 BEFORE INSERT ON playlist_favorite FOR EACH ROW INSERT INTO playlist_favorite_count (playlist_id,count) VALUES (NEW.playlist_id, 1) ON DUPLICATE KEY UPDATE playlist_favorite_count.count = playlist_favorite_count.count + 1;

DROP TRIGGER IF EXISTS tr2;
CREATE TRIGGER tr2 BEFORE DELETE ON `message` FOR EACH ROW UPDATE `channel` SET `message_cnt`=`message_cnt`-1;
```

## Group毎に最新のレコードをSELECTする

```sql
SELECT * FROM isu_condition AS a JOIN (SELECT user_id, MAX(created_at) AS created_at FROM isu_condition GROUP BY user_id) AS b
ON a.user_id = b.user_id WHERE a.created_at = b.created_at;
```

## 個テーブルの個数をサブクエリで持ってきてUpdate

```sql
UPDATE `channel`, (SELECT channel_id, COUNT(1) AS `message_cnt` FROM `message` GROUP BY channel_id) AS `summary` SET `channel`.`message_cnt`=`summary`.`message_cnt` WHERE channel.id = summary.channel_id
```


## アドバイザリロック

```go
func main() {
	db.Exec("SELECT GET_LOCK(?, ?)", id, 5)
	defer db.Exec("SELECT RELEASE_LOCK(?)", id)
}
```

---

<!-- ここから下は ISUCON10〜14 の上位解法・公式講評の調査から追記したもの -->

## interpolateParams

prepared statementの PREPARE / EXECUTE / CLOSE の往復が1回になる。pt-query-digestで `PREPARE` や `administrator command: Close stmt` が上位なら確実に効く (10〜14で毎回上位チームが使用)
```go
dsn := "isucon:isucon@tcp(127.0.0.1:3306)/isucon?parseTime=true&interpolateParams=true"
// or cfg.InterpolateParams = true
```

## 書き込み系のmy.cnf

```ini
[mysqld]
disable-log-bin                  # MySQL8はデフォルトでbinlogを出す
innodb_flush_log_at_trx_commit = 0   # 2でもよい
innodb_doublewrite = 0
innodb_flush_method = O_DIRECT
innodb_log_writer_threads = off  # CPUコアが少ないとき
innodb_buffer_pool_size = 1G     # メモリの量を見て。上げてスコアが下がった例もある
performance_schema = off
skip-name-resolve = 1
max_connections = 1024
```

- redo logを tmpfs に置く例もある(10予選, 12予選 matsuu)。再起動試験でデータが残るか必ず確認する
```ini
innodb_log_group_home_dir = /dev/shm/
```
```
# /etc/apparmor.d/usr.sbin.mysqld に追記
/dev/shm/ r,
/dev/shm/** rwk,
```
- MySQL5.7のクエリキャッシュは `query_cache_size` だけでなく `query_cache_type = ON` も必要 (8.0には無い)

cf. https://matsuu.hatenablog.com/entry/2020/09/13/131145

## 外部キー制約を外す

INSERT/UPDATE/DELETEのたびに参照先をチェックする。通常のINDEXに置き換える (11本選)
```sql
-- CONSTRAINT FK_registrations_user_id FOREIGN KEY (user_id) REFERENCES users (id)
INDEX idx_user_id (user_id)
```

## 自然キーを主キーにする

`id AUTO_INCREMENT` + 別インデックスより、一意な自然キーをPK(クラスタインデックス)にした方が読み書きとも軽い。複合PKの列順はWHEREに合わせる (11予選 isu_condition, 11本選 unread_announcements)
```sql
CREATE TABLE isu_condition (
  jia_isu_uuid CHAR(36) NOT NULL,
  `timestamp` DATETIME NOT NULL,
  ...
  PRIMARY KEY (jia_isu_uuid, `timestamp`)
);
```
- `/initialize` でテーブルをDROPする問題ではschema.sql側を書き換える

## 降順インデックス

`ORDER BY popularity DESC, id ASC` のようなASC/DESC混在はfilesortになる (10予選)
```sql
-- MySQL8
ALTER TABLE estate ADD INDEX idx_pop (popularity DESC, id ASC);
-- MySQL5.7, MariaDBはDESCが無視されるので符号反転したカラムを作る
ALTER TABLE estate ADD COLUMN popularity_desc INT AS (-popularity) STORED, ADD INDEX (popularity_desc, id);
-- ORDER BY popularity_desc, id
```
- ソート順を勝手に変えるとベンチが落ちる。`id ASC` も含めて互換を保つ

## 重い条件はORDER BY LIMITの後に評価させる

関数・LIKEなどの重い条件をWHEREに入れると、軽い条件に当たった全行で評価される。並べた結果を上から評価してLIMITで打ち切る (10予選 2087→2978点)
```sql
-- before
SELECT * FROM estate WHERE α AND ST_Contains(β) ORDER BY γ LIMIT 50
-- after
SELECT * FROM (SELECT * FROM estate WHERE α ORDER BY γ) AS t WHERE ST_Contains(β) LIMIT 50
```
cf. https://akouryy.hatenablog.jp/entry/2020/09/13/130415

## 範囲条件を区分IDにする

検索の範囲が決まった区分だけなら、区分番号のカラムを持たせて等値条件にすると複合インデックスが効く (10予選)
```sql
ALTER TABLE estate ADD COLUMN rent_id INT AS (
  CASE WHEN rent < 50000 THEN 0 WHEN rent < 100000 THEN 1 WHEN rent < 150000 THEN 2 ELSE 3 END
) STORED, ADD INDEX (rent_id);
-- WHERE rent >= ? AND rent < ? → WHERE rent_id = ?
```

## 空間インデックス

緯度経度の多角形検索はPOINTカラム + SPATIAL INDEXで1クエリにする (10予選 なぞって検索)
```sql
-- MySQL8は SRID 0 を付けないとインデックスが使われない。INSERT時にもpointを入れる
ALTER TABLE estate ADD COLUMN point POINT SRID 0 NULL;
UPDATE estate SET point = POINT(latitude, longitude);
ALTER TABLE estate MODIFY COLUMN point POINT SRID 0 NOT NULL, ADD SPATIAL INDEX (point);
SELECT * FROM estate WHERE ST_Contains(ST_PolygonFromText(?), point) ORDER BY popularity DESC, id LIMIT 50;
```
cf. https://isucon.net/archives/55025156.html

## 読んで書くをUPDATE1発にする

```sql
-- SELECT ... FOR UPDATE してから在庫を減らす → 条件付きUPDATEにしてRowsAffectedで判定 (10予選)
UPDATE chair SET stock = stock - 1 WHERE id = ? AND stock > 0;
-- 残高を読んで足して書き戻す → 相対更新 (12本選)
UPDATE users SET isu_coin = isu_coin + ? WHERE id = ?;
-- マスタを読んでINSERT → INSERT ... SELECT
INSERT INTO user_cards (user_id, card_id, amount_per_sec) SELECT ?, id, amount_per_sec FROM item_masters WHERE id = ?;
```
```go
res, _ := db.Exec("UPDATE chair SET stock = stock - 1 WHERE id = ? AND stock > 0", id)
if n, _ := res.RowsAffected(); n == 0 {
	return c.NoContent(http.StatusNotFound) // 売り切れ
}
```

## 不要なトランザクションを外す

SELECTだけのハンドラのBEGIN/COMMITは往復が増え、COMMITがスロークエリ上位に出る (11本選)。外してよいかはベンチの整合性チェックで確認する

## 論理削除をやめる

`deleted_at IS NULL` 付きの検索は行数もインデックスも膨らむ。整合性チェックで履歴を見られないなら物理削除にする (12本選 user_sessions 78910→85692点)
```sql
-- UPDATE user_sessions SET deleted_at = ? WHERE user_id = ? AND deleted_at IS NULL
DELETE FROM user_sessions WHERE user_id = ?;
-- UNIQUE (user_id, deleted_at, session_id) → UNIQUE (session_id)
```

## OFFSETをやめる(keyset pagination)

OFFSETが大きいと読み飛ばす行もスキャンする。次ページのURLをサーバーが返す(Linkヘッダなど)なら書き換えられる (11本選)
```sql
SELECT * FROM announcements WHERE id < ? ORDER BY id DESC LIMIT 20
```

## 最新値・累積値を親テーブルに持つ

履歴テーブルからLAG/ROW_NUMBERで最新状態や総距離を毎回出しているなら、親テーブルに持って差分を足し込む (14 chair_locations)
```sql
ALTER TABLE chairs ADD latitude INT NULL, ADD longitude INT NULL, ADD total_distance INT NOT NULL DEFAULT 0;
UPDATE chairs SET latitude = ?, longitude = ?, total_distance = total_distance + ? WHERE id = ?;
```
- 「空いているか」なども最新ステータスから導出せずフラグカラムにしてINDEXを張る (14 is_matchable)
- 追加したカラムの初期値を埋めた状態で初期データのダンプを作り直すと `/initialize` が速い

## 書き込み時にランキングを作る

ランキングが変わるのは書き込み時だけ。書き込み時に計算してbulk upsertし、読み込みはSELECTのみにする (12予選)
```sql
INSERT INTO ranking (tenant_id, competition_id, `rank`, score, player_id) VALUES (...), (...)
ON DUPLICATE KEY UPDATE score = VALUES(score), player_id = VALUES(player_id);
```
- best/latestのような集計は `IF(? >= IFNULL(best_score, 0), ?, best_score)` で1本のUPDATEにまとめられる (10本選)

## SQLiteからMySQLへ移す

- テナントごとのSQLiteファイルなどは MySQL へ移すと複数台・インデックス・トランザクションが使える (12予選)
- 1行ずつのINSERTに変換すると遅いので、変換済みのデータを `/initialize` で流し込めるよう事前にダンプしておく(`/initialize` は時間制限あり)
- `mysql --default-character-set=utf8mb4` を付けないと文字化けする

cf. https://isucon.net/archives/56850281.html

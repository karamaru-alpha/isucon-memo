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


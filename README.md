<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [セットアップ](#%E3%82%BB%E3%83%83%E3%83%88%E3%82%A2%E3%83%83%E3%83%97)
- [調査](#%E8%AA%BF%E6%9F%BB)
    - [Mysql - スロークエリ](#mysql---%E3%82%B9%E3%83%AD%E3%83%BC%E3%82%AF%E3%82%A8%E3%83%AA)
    - [Nginx - アクセスログ](#nginx---%E3%82%A2%E3%82%AF%E3%82%BB%E3%82%B9%E3%83%AD%E3%82%B0)
- [チューニング開始](#%E3%83%81%E3%83%A5%E3%83%BC%E3%83%8B%E3%83%B3%E3%82%B0%E9%96%8B%E5%A7%8B)
- [Go](#go)
    - [Ubuntu環境にインストール　(doc)](#ubuntu%E7%92%B0%E5%A2%83%E3%81%AB%E3%82%A4%E3%83%B3%E3%82%B9%E3%83%88%E3%83%BC%E3%83%AB%E3%80%80doc)
    - [Build](#build)
    - [logをファイルに出力](#log%E3%82%92%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E3%81%AB%E5%87%BA%E5%8A%9B)
    - [UnixDomainSocket](#unixdomainsocket)
- [Mysql (MariaDB)](#mysql-mariadb)
    - [MysqlからMariaDBに乗り換える](#mysql%E3%81%8B%E3%82%89mariadb%E3%81%AB%E4%B9%97%E3%82%8A%E6%8F%9B%E3%81%88%E3%82%8B)
    - [MariaDBを最新にする](#mariadb%E3%82%92%E6%9C%80%E6%96%B0%E3%81%AB%E3%81%99%E3%82%8B)
    - [ユーザの作成](#%E3%83%A6%E3%83%BC%E3%82%B6%E3%81%AE%E4%BD%9C%E6%88%90)
    - [TroubleShoot](#troubleshoot)
- [Nginx](#nginx)
    - [インストール](#%E3%82%A4%E3%83%B3%E3%82%B9%E3%83%88%E3%83%BC%E3%83%AB)
    - [ファイル上限を確認・拡張する](#%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E4%B8%8A%E9%99%90%E3%82%92%E7%A2%BA%E8%AA%8D%E3%83%BB%E6%8B%A1%E5%BC%B5%E3%81%99%E3%82%8B)
- [Linux](#linux)
    - [Systemdでアプリを動かす](#systemd%E3%81%A7%E3%82%A2%E3%83%97%E3%83%AA%E3%82%92%E5%8B%95%E3%81%8B%E3%81%99)
- [Nginx](#nginx-1)
    - [keepaliveを有効する](#keepalive%E3%82%92%E6%9C%89%E5%8A%B9%E3%81%99%E3%82%8B)
    - [静的ファイルの配信](#%E9%9D%99%E7%9A%84%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E3%81%AE%E9%85%8D%E4%BF%A1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## セットアップ

- Makefileから必要ツールをダウンロード
```sh
curl -OL https://raw.githubusercontent.com/karamaru-alpha/isucon-memo/main/Makefile
make setup 
```

- ghのインストール 
```
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt install -y gh
gh auth login # GitHub.com -> SSH -> /home/isucon/.ssh/id_rsa.pub -> Paste an authentication token -> https://github.com/settings/tokens
```

- githubで管理するまで
```sh
git remote add origin git@github.com:karamaru-alpha/${REPO}.git
git branch -m master main
git add . && git commit 
git push -u origin main
```

## 調査

- VMの状態確認
```sh
arch
free -h
fgrep 'cpu cores' /proc/cpuinfo | sort -u | sed 's/.*: //'
systemctl list-unit-files --type=service
```

- DBのバージョンとスキーマの確認

```sh
mysql --version
```

```sh
mysqldump -uroot -proot --host 127.0.0.1 --port 3306 ${DATABASE} --compact --no-data --compact --no-data | grep -v "^SET" | grep -v "^/\*\!" | perl -ple 's@CREATE TABLE @\nCREATE TABLE @g';
```

```sql
SELECT
  table_name, engine, table_rows,
  floor((data_length+index_length)/1024/1024) AS total_mb,
  floor((data_length)/1024/1024) AS data_mb,
  floor((index_length)/1024/1024) AS index_mb
FROM
  information_schema.tables
WHERE
  table_schema=database()
ORDER BY
  (data_length+index_length) DESC;
```

- Makefileに変数を適用する

#### Mysql - スロークエリ

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

#### Nginx - アクセスログ

- 設定ファイルをホームディレクトリに持ってくる

```
cp /etc/nginx/nginx.conf nginx.conf
cp /etc/nginx/sites-enabled/$(APP).conf $(APP)
```

- アクセスログを吐くようにnginx.confを設定する
```
sudo touch /var/log/nginx/access.log
sudo touch /var/log/nginx/error.log
sudo chmod 777 /var/log/nginx/access.log
sudo chmod 777 /var/log/nginx/error.log
```

```
http {
    log_format with_time '$remote_addr - $remote_user [$time_local] '
                 '"$request" $status $body_bytes_sent '
                 '"$http_referer" "$http_user_agent" $request_time';
    access_log /var/log/nginx/access.log with_time;
}
```

## チューニング開始

- index貼る
- LIMITつける
- app側でPrepareする

```
interpolateParams=true
```
                           

- deadlock
```sql
SHOW ENGINE INNODB STATUS;
```

## Go

#### Ubuntu環境にインストール　([doc](https://go.dev/doc/install))
```
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
apt info golang
sudo apt install golang (golang=1.18.0)
```

#### Build
```
go tool dist list
GOOS=linux GOARCH=arm64 go build -o isucon *.go
```

#### logをファイルに出力

```
sudo touch /var/log/go.log 
sudo chmod 777 /var/log/go.log
```

```go
import (
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile)
	logfile, err := os.OpenFile("/var/log/go.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
    defer logfile.Close()
    log.SetOutput(logfile)
    e.Logger.SetOutput(logfile)
    log.Print("initialize!!!!")
}
```

#### UnixDomainSocket

cf. [karamaru-alpha/kayac-isucon-2022](https://github.com/karamaru-alpha/kayac-isucon-2022/compare/unix-domain?expand=1)

## Mysql (MariaDB)

#### MysqlからMariaDBに乗り換える

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

#### MariaDBを最新にする

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

#### ユーザの作成
```sql
DROP USER 'isucon'@'localhost';
CREATE USER 'isucon'@'localhost' IDENTIFIED BY 'isucon';
GRANT ALL PRIVILEGES ON * . * TO 'isucon'@'localhost';
FLUSH PRIVILEGES;
```

#### TroubleShoot

- Unknown collation: 'utf8mb4_0900_ai_ci'
  - sed -i 's/utf8mb4_0900_ai_ci/utf8mb4_unicode_ci/g' sql/dump.sql

## Nginx

#### インストール

```bash
sudo apt update
sudo apt install nginx
sudo ufw allow 'Nginx Full'
sudo systemctl enable nginx
systemctl list-unit-files --type=service
```

#### ファイル上限を確認・拡張する

```bash
ps ax | grep nginx | grep worker
cat /proc/${PID}/limits

sudo mkdir /etc/systemd/system/nginx.service.d

vi /etc/systemd/system/nginx.service.d/limit.conf
[Service]
LimitNOFILE=32768


systemctl daemon-reload
systemctl restart nginx
```

## Linux

#### Systemdでアプリを動かす

```
cd /etc/systemd/system
sudo vim golang.service

---
[Unit]
Description = isucon go servce

[Service]
ExecStart=/home/isucon/webapp/golang/isucon
WorkingDirectory=/home/isucon/webapp/golang
Restart=always
Type=simple
User=isucon
Group=isucon
# Other directives omitted
# (file size)
LimitFSIZE=infinity
# (cpu time)
LimitCPU=infinity
# (virtual memory size)
LimitAS=infinity
# (open files)
LimitNOFILE=64000
# (processes/threads)
LimitNPROC=64000

[Install]
WantedBy = multi-user.target
---

sudo systemctl daemon-reload
```

## Nginx

#### keepaliveを有効する

HTTP/1.1を使用する&Connectionヘッダを空にする必要がある
```conf
upstream app {
  server 127.0.0.1:3000;
  keepalive 32;
  keepalive_requests 10000;
}
server {
  listen 80;
  root /public/;
  location / {
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_pass http://app;
  }
}
```

#### 静的ファイルの配信

```conf
server {
  listen 80;

  root /public/;

  location / {
    proxy_pass http://127.0.0.1:3000;
  }

  location /assets/ {
    try_files $uri /;
    expires 1d;
  }
}
```

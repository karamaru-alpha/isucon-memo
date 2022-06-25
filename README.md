## 初手

- ベンチを回す

## セットアップ

- Makefileから必要ツールをダウンロード
```
curl -OL https://raw.githubusercontent.com/karamaru-alpha/isucon-memo/main/Makefile
make setup 
```

- githubで管理するまで
```
git remote add origin git@github.com:karamaru-alpha/${REPO}.git
git branch -m master main
git add . &&& git commit 
git push -u origin main
```

## 調査

- Makefileに変数を適用する

- データ量とスキーマをissueに起票する

```
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

#### Mysql - スロークエリ

- 設定ファイルをホームディレクトリに持ってくる (`mysql --help | grep my.cnf`で探せる)
```
cp /etc/mysql/my.cnf my.cnf
```

- スロークエリを吐くように[my.cnf](./etc/my.cnf)設定する

```
sudo touch /var/log/mysql/slow-query.log
sudo chown -R mysql /var/log/mysql/slow-query.log
```

```
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
sudo chown -R nginx /var/log/nginx/access.log
```

```
http {
    log_format with_time '$remote_addr - $remote_user [$time_local] '
                 '"$request" $status $body_bytes_sent '
                 '"$http_referer" "$http_user_agent" $request_time';
    access_log /var/log/nginx/access.log with_time;
}
```

#### GithubCliで自動解析する

```
./analyze.sh 初期実装 100
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

#### インストール




#### アンインストール
```
```

```bash
sudo apt-get remove --purge mysql-server* mysql-common
sudo rm -r /etc/mysql
sudo rm -r /var/lib/mysql
sudo apt autoremove -y
sudo apt clean
sudo aa-remove-unknown
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


#### アーキテクチャ確認
```
arch
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

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

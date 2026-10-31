# nginx の解法

> isucon-memo の README.md から抜き出したもの。スニペットは雛形なので、そのまま貼らずに対象アプリの型名・テーブル名・ライブラリに合わせて書き直すこと。

## Nginx - アクセスログ

- 設定ファイルをホームディレクトリに持ってくる

```sh
cp /etc/nginx/nginx.conf nginx.conf
cp /etc/nginx/sites-enabled/$(APP).conf $(APP)
```

- アクセスログを吐くようにnginx.confを設定する
```sh
sudo touch /var/log/nginx/access.log
sudo touch /var/log/nginx/error.log
sudo chmod 777 /var/log/nginx/access.log
sudo chmod 777 /var/log/nginx/error.log
```

```nginx
http {
    log_format with_time '$remote_addr - $remote_user [$time_local] '
      '"$request" $status $body_bytes_sent '
      '"$http_referer" "$http_user_agent" $request_time';
    access_log /var/log/nginx/access.log with_time;
}
``` 

## keepaliveを有効する

HTTP/1.1を使用する&Connectionヘッダを空にする必要がある
```nginx configuration
upstream s1 {
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
    proxy_pass http://s1;
  }
}
```

## ファイル上限を確認・拡張する

```bash
ps ax | grep nginx | grep worker
cat /proc/${PID}/limits

sudo mkdir /etc/systemd/system/nginx.service.d

sudo vi /etc/systemd/system/nginx.service.d/limit.conf
[Service]
LimitNOFILE=32768


sudo systemctl daemon-reload
sudo systemctl restart nginx
```

## 静的ファイルのクライアントキャッシュ

```nginx configuration
// root /home/isucon/webapp/public;
// index index.html;
location /assets/ {
  expires 1d;
  try_files $uri /index.html;
}
location ~ ^/isu/(.*?) {
  rewrite .* /index.html;
  alias /home/isucon/webapp/public/index.html;
}
location /register {
  alias /home/isucon/webapp/public/index.html;
} 
```

## レスポンスキャッシュ(ProxyCache)

```nginx configuration
http {
    # キャッシュ先のファイル指定・2階層で保存・zone1キー名で1M確保・1ギガまで使う・2分で削除
    proxy_cache_path /var/cache/nginx/cache levels=1:2 keys_zone=zone1:1m max_size=1g inactive=2m;
    proxy_temp_path  /var/cache/nginx/tmp;

    location /path/to {
	    proxy_http_version 1.1;
            proxy_set_header Connection "";
            proxy_cache zone1;
            proxy_cache_valid 200 302 2m;
            # proxy_cache_key $scheme$proxy_host$uri$is_args$args;
            proxy_pass http://s1;
    }
}
```

```sh
sudo mkdir -p /var/cache/nginx/cache
sudo mkdir -p /var/cache/nginx/tmp
sudo chmod 777 /var/cache/nginx/cache
sudo chmod 777 /var/cache/nginx/tmp
```


## 認証付きの静的配信(X-Accel-Redirect)

```go
func getIsuIcon(c echo.Context) error {
	// 認証してから
	// ↓必要であれば存在チェック
	// if _, err := os.Stat(fmt.Sprintf("%s/%s_%s", iconFilePath, jiaUserID, jiaIsuUUID)); err != nil {
	// 	return c.String(http.StatusNotFound, "not found: isu")
	// }
	c.Response().Header().Set("X-Accel-Redirect", fmt.Sprintf("/icon/%s_%s", jiaUserID, jiaIsuUUID))
	return c.NoContent(http.StatusOK)
}
```

```nginx configuration
# ここにリクエストが来て -> app
location ^~ /api/isu/(.*)/icon {
    expires 1d;
    add_header cache-control public;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_pass http://app;
}

# ここでaccel-redirect <- app
location /icon/ {
    internal;
    alias /home/isucon/webapp/icons/;
    expires 1d;
    add_header cache-control public;
}
```

## リクエストメソッドでリダイレクト

```nginx configuration
location /path {
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    if ($request_method = GET) {
        proxy_pass http://app1;
        break;
    }
    proxy_pass http://app2;
}

# limit_except GET {
#     proxy_pass http://app1;
# }
# proxy_pass http://app2;
```

## パスパラメータでリダイレクト

```nginx configuration
location /path {
    location ~ ^/path/karamaru|karaki {
        proxy_pass http://s1;
    }
    location /path {
        proxy_pass http://s2;
    }
}
```

## クエリパラメータでリダイレクト

```nginx configuration
# $arg_{キー}でクエリストリングが取得できる
location /query {
    if ( $arg_name ~ karamaru|karaki ) {
        proxy_pass http://s1;
        break;
    }
    proxy_pass http://s2;
}
```

## URLの部分一致でリダイレクト

```nginx configuration
location / {
   location ~ ^.*?karamaru|karaki.*$ {
        proxy_pass http://s1;
   }
   location / {
          proxy_pass http://s2;
   }
}
```

## Botからのリクエストを拒否

```nginx configuration
map $http_user_agent $bot {
    default 0;
    "~ISUCONbot" 1;
    "~Mediapartners-ISUCON" 1;
    "~ISUCONCoffee" 1;
    "~ISUCONFeedSeeker" 1;
    "~crawler \(https://isucon\.invalid/(support/faq/|help/jp/)" 1;
    "~isubot" 1;
    "~Isupider" 1;
    "~*(bot|crawler|spider)(?:[-_ .\/;@()]|$)" 1;
}

server {
    root /home/isucon/isucon10-qualify/webapp/public;
    listen 80 default_server;
    listen [::]:80 default_server;

    if ($bot = 1) { return 503; }
    
    # if ( $http_user_agent ~* ISUCONbot(-Mobile)? ) { return 503; }
}
```

## 圧縮 / 事前圧縮

```nginx configuration
gzip on;
gzip_types text/css application/javascript application/json application/font-woff application/font-tff image/gif image/png image/jpeg image/svg+xml image/x-icon application/octet-stream;
gzip_min_length 1k;
```

```nginx configuration
location /asset/ { 
    gzip_static always;
    gunzip on;
    expires 1d;
    try_files $uri /index.html;
}
```

---

<!-- ここから下は ISUCON10〜14 の上位解法・公式講評の調査から追記したもの -->

## SPAのindex.htmlを返す

フロントのルーティング用パスはnginxで `index.html` にフォールバックし、変わらないものには `Cache-Control` を付ける (11予選)
```nginx configuration
root /home/isucon/webapp/public;
location / {
    try_files $uri /index.html;
    expires 1d;
}
location /api/ {
    proxy_pass http://app;
}
```

## リクエストボディのバッファ

エラーログに `a client request body is buffered to a temporary file` が出ていたら広げる (10予選 CSV入稿)
```nginx configuration
client_max_body_size 10m;
client_body_buffer_size 512k;
```
大きなアップロード・ダウンロードはバッファリング自体を切ってアプリにストリームで渡す (11本選)
```nginx configuration
location ~ ^/api/courses/.*/assignments$ {
    proxy_request_buffering off;
    proxy_pass http://app;
}
location ~ ^/api/courses/.*/export$ {
    proxy_buffering off;
    proxy_pass http://app;
}
```

## 1接続あたりのリクエスト上限

HTTP/2はデフォルト1000リクエストで接続を切り、高負荷時にエラーになる (11予選 FAIL→70,000点)
```nginx configuration
http {
    keepalive_requests 1000000;   # nginx 1.19.7未満は http2_max_requests
}
```

## Hostで振り分ける

マルチテナントでテナントごとにデータを持つサーバーが決まっているなら、サブドメインで振り分ければデータを共有せず複数台にできる (12予選)
```nginx configuration
map $host $backend {
    default                  127.0.0.1:3000;
    "~^[a-m][a-z0-9-]*\.t\.isucon\.dev$" 192.168.0.12:3000;
}
location /api/ {
    proxy_set_header Host $host;
    proxy_pass http://$backend;
}
```
cf. https://isucon.net/archives/56842718.html

## TLS終端を分散する

nginxのTLS処理でCPUが張り付くなら、`stream` で443をTCPのまま各台に振り分けて各台で終端する (14感想戦)
```nginx configuration
stream {
    upstream backend {
        server 127.0.0.1:8443 weight=2;
        server 192.168.0.12:8443 weight=5;
    }
    server {
        listen 443;
        proxy_pass backend;
    }
}
# 各台の http { server { listen 8443 ssl http2; ... } }
```
cf. https://zenn.dev/ponyo877/articles/1e29255b9d19bf

## worker設定

```nginx configuration
worker_processes auto;
worker_cpu_affinity auto;
```

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [TODO](#todo)
- [セットアップ](#%E3%82%BB%E3%83%83%E3%83%88%E3%82%A2%E3%83%83%E3%83%97)
    - [Makefileから必要ツールをダウンロード](#makefile%E3%81%8B%E3%82%89%E5%BF%85%E8%A6%81%E3%83%84%E3%83%BC%E3%83%AB%E3%82%92%E3%83%80%E3%82%A6%E3%83%B3%E3%83%AD%E3%83%BC%E3%83%89)
    - [ghのインストール](#gh%E3%81%AE%E3%82%A4%E3%83%B3%E3%82%B9%E3%83%88%E3%83%BC%E3%83%AB)
    - [aliasの導入](#alias%E3%81%AE%E5%B0%8E%E5%85%A5)
    - [githubで管理するまで](#github%E3%81%A7%E7%AE%A1%E7%90%86%E3%81%99%E3%82%8B%E3%81%BE%E3%81%A7)
- [調査](#%E8%AA%BF%E6%9F%BB)
    - [VMの状態確認](#vm%E3%81%AE%E7%8A%B6%E6%85%8B%E7%A2%BA%E8%AA%8D)
    - [DBのバージョンとスキーマの確認](#db%E3%81%AE%E3%83%90%E3%83%BC%E3%82%B8%E3%83%A7%E3%83%B3%E3%81%A8%E3%82%B9%E3%82%AD%E3%83%BC%E3%83%9E%E3%81%AE%E7%A2%BA%E8%AA%8D)
    - [Mysql - スロークエリ](#mysql---%E3%82%B9%E3%83%AD%E3%83%BC%E3%82%AF%E3%82%A8%E3%83%AA)
    - [Nginx - アクセスログ](#nginx---%E3%82%A2%E3%82%AF%E3%82%BB%E3%82%B9%E3%83%AD%E3%82%B0)
    - [Go - ログ](#go---%E3%83%AD%E3%82%B0)
- [Go](#go)
    - [Ubuntu環境にインストール](#ubuntu%E7%92%B0%E5%A2%83%E3%81%AB%E3%82%A4%E3%83%B3%E3%82%B9%E3%83%88%E3%83%BC%E3%83%AB)
    - [Build](#build)
    - [SingleFlight](#singleflight)
    - [httpを広げる](#http%E3%82%92%E5%BA%83%E3%81%92%E3%82%8B)
    - [シリアライザを切り替える](#%E3%82%B7%E3%83%AA%E3%82%A2%E3%83%A9%E3%82%A4%E3%82%B6%E3%82%92%E5%88%87%E3%82%8A%E6%9B%BF%E3%81%88%E3%82%8B)
    - [オンメモリ戦略](#%E3%82%AA%E3%83%B3%E3%83%A1%E3%83%A2%E3%83%AA%E6%88%A6%E7%95%A5)
    - [画像をfileで読み書き](#%E7%94%BB%E5%83%8F%E3%82%92file%E3%81%A7%E8%AA%AD%E3%81%BF%E6%9B%B8%E3%81%8D)
    - [一定時間毎に処理をする](#%E4%B8%80%E5%AE%9A%E6%99%82%E9%96%93%E6%AF%8E%E3%81%AB%E5%87%A6%E7%90%86%E3%82%92%E3%81%99%E3%82%8B)
    - [UnixDomainSocket](#unixdomainsocket)
- [Mysql (MariaDB)](#mysql-mariadb)
    - [DB用のインスタンスのセットアップ](#db%E7%94%A8%E3%81%AE%E3%82%A4%E3%83%B3%E3%82%B9%E3%82%BF%E3%83%B3%E3%82%B9%E3%81%AE%E3%82%BB%E3%83%83%E3%83%88%E3%82%A2%E3%83%83%E3%83%97)
    - [外部からのアクセスを許容する](#%E5%A4%96%E9%83%A8%E3%81%8B%E3%82%89%E3%81%AE%E3%82%A2%E3%82%AF%E3%82%BB%E3%82%B9%E3%82%92%E8%A8%B1%E5%AE%B9%E3%81%99%E3%82%8B)
    - [ユーザの作成](#%E3%83%A6%E3%83%BC%E3%82%B6%E3%81%AE%E4%BD%9C%E6%88%90)
    - [MysqlからMariaDBに乗り換える](#mysql%E3%81%8B%E3%82%89mariadb%E3%81%AB%E4%B9%97%E3%82%8A%E6%8F%9B%E3%81%88%E3%82%8B)
    - [MariaDBを最新にする](#mariadb%E3%82%92%E6%9C%80%E6%96%B0%E3%81%AB%E3%81%99%E3%82%8B)
    - [デッドロックの調査](#%E3%83%87%E3%83%83%E3%83%89%E3%83%AD%E3%83%83%E3%82%AF%E3%81%AE%E8%AA%BF%E6%9F%BB)
    - [TroubleShoot](#troubleshoot)
    - [Mysqlの起動を待つ](#mysql%E3%81%AE%E8%B5%B7%E5%8B%95%E3%82%92%E5%BE%85%E3%81%A4)
    - [bulkInsert](#bulkinsert)
    - [IN句](#in%E5%8F%A5)
    - [コネクションプール](#%E3%82%B3%E3%83%8D%E3%82%AF%E3%82%B7%E3%83%A7%E3%83%B3%E3%83%97%E3%83%BC%E3%83%AB)
    - [DBの起動を待つ](#db%E3%81%AE%E8%B5%B7%E5%8B%95%E3%82%92%E5%BE%85%E3%81%A4)
    - [generatedColumns](#generatedcolumns)
    - [1byte長の半角文字列をピッタリ格納する](#1byte%E9%95%B7%E3%81%AE%E5%8D%8A%E8%A7%92%E6%96%87%E5%AD%97%E5%88%97%E3%82%92%E3%83%94%E3%83%83%E3%82%BF%E3%83%AA%E6%A0%BC%E7%B4%8D%E3%81%99%E3%82%8B)
    - [UUIDをBINARY(16)で格納する](#uuid%E3%82%92binary16%E3%81%A7%E6%A0%BC%E7%B4%8D%E3%81%99%E3%82%8B)
    - [水平分割](#%E6%B0%B4%E5%B9%B3%E5%88%86%E5%89%B2)
    - [Upsert](#upsert)
    - [trigger](#trigger)
    - [Group毎に最新のレコードをSELECTする](#group%E6%AF%8E%E3%81%AB%E6%9C%80%E6%96%B0%E3%81%AE%E3%83%AC%E3%82%B3%E3%83%BC%E3%83%89%E3%82%92select%E3%81%99%E3%82%8B)
    - [アドバイザリロック](#%E3%82%A2%E3%83%89%E3%83%90%E3%82%A4%E3%82%B6%E3%83%AA%E3%83%AD%E3%83%83%E3%82%AF)
- [Nginx](#nginx)
    - [インストール](#%E3%82%A4%E3%83%B3%E3%82%B9%E3%83%88%E3%83%BC%E3%83%AB)
    - [keepaliveを有効する](#keepalive%E3%82%92%E6%9C%89%E5%8A%B9%E3%81%99%E3%82%8B)
    - [ファイル上限を確認・拡張する](#%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E4%B8%8A%E9%99%90%E3%82%92%E7%A2%BA%E8%AA%8D%E3%83%BB%E6%8B%A1%E5%BC%B5%E3%81%99%E3%82%8B)
    - [静的ファイルのクライアントキャッシュ](#%E9%9D%99%E7%9A%84%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E3%81%AE%E3%82%AF%E3%83%A9%E3%82%A4%E3%82%A2%E3%83%B3%E3%83%88%E3%82%AD%E3%83%A3%E3%83%83%E3%82%B7%E3%83%A5)
    - [レスポンスキャッシュ(ProxyCache)](#%E3%83%AC%E3%82%B9%E3%83%9D%E3%83%B3%E3%82%B9%E3%82%AD%E3%83%A3%E3%83%83%E3%82%B7%E3%83%A5proxycache)
    - [認証付きの静的配信(X-Accel-Redirect)](#%E8%AA%8D%E8%A8%BC%E4%BB%98%E3%81%8D%E3%81%AE%E9%9D%99%E7%9A%84%E9%85%8D%E4%BF%A1x-accel-redirect)
    - [リクエストメソッドでリダイレクト](#%E3%83%AA%E3%82%AF%E3%82%A8%E3%82%B9%E3%83%88%E3%83%A1%E3%82%BD%E3%83%83%E3%83%89%E3%81%A7%E3%83%AA%E3%83%80%E3%82%A4%E3%83%AC%E3%82%AF%E3%83%88)
    - [パスパラメータでリダイレクト](#%E3%83%91%E3%82%B9%E3%83%91%E3%83%A9%E3%83%A1%E3%83%BC%E3%82%BF%E3%81%A7%E3%83%AA%E3%83%80%E3%82%A4%E3%83%AC%E3%82%AF%E3%83%88)
    - [クエリパラメータでリダイレクト](#%E3%82%AF%E3%82%A8%E3%83%AA%E3%83%91%E3%83%A9%E3%83%A1%E3%83%BC%E3%82%BF%E3%81%A7%E3%83%AA%E3%83%80%E3%82%A4%E3%83%AC%E3%82%AF%E3%83%88)
    - [URLの部分一致でリダイレクト](#url%E3%81%AE%E9%83%A8%E5%88%86%E4%B8%80%E8%87%B4%E3%81%A7%E3%83%AA%E3%83%80%E3%82%A4%E3%83%AC%E3%82%AF%E3%83%88)
    - [Botからのリクエストを拒否](#bot%E3%81%8B%E3%82%89%E3%81%AE%E3%83%AA%E3%82%AF%E3%82%A8%E3%82%B9%E3%83%88%E3%82%92%E6%8B%92%E5%90%A6)
    - [圧縮 / 事前圧縮](#%E5%9C%A7%E7%B8%AE--%E4%BA%8B%E5%89%8D%E5%9C%A7%E7%B8%AE)
- [Linux](#linux)
    - [ファイルディスクリプタ上限up](#%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB%E3%83%87%E3%82%A3%E3%82%B9%E3%82%AF%E3%83%AA%E3%83%97%E3%82%BF%E4%B8%8A%E9%99%90up)
    - [Systemdでアプリを動かす](#systemd%E3%81%A7%E3%82%A2%E3%83%97%E3%83%AA%E3%82%92%E5%8B%95%E3%81%8B%E3%81%99)
    - [shellの変更](#shell%E3%81%AE%E5%A4%89%E6%9B%B4)
    - [githubの鍵でssh](#github%E3%81%AE%E9%8D%B5%E3%81%A7ssh)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## TODO
- pprofをいい感じに取る方法の模索
- sync.Map活用

## セットアップ

#### Makefileから必要ツールをダウンロード
```sh
curl -OL https://raw.githubusercontent.com/karamaru-alpha/isucon-memo/main/Makefile
make setup 
```

#### ghのインストール 
```sh
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install -y gh
gh auth login # GitHub.com -> SSH -> /home/isucon/.ssh/id_rsa.pub -> Paste an authentication token -> https://github.com/settings/tokens
```

#### aliasの導入
```sh
sudo cat <<EOL >> ~/.bashrc
alias cm="git commit -m"
alias ad="git add ."
alias co="git checkout"
alias save="git add . && git commit -m"
alias cob="git checkout -b"
alias mg="git merge"
alias rename="git branch -m"
alias del="git branch -D"
alias refresh="git checkout . && git clean -df"
EOL
source ~/.bashrc
```

#### githubで管理するまで
```sh
git remote add origin git@github.com:karamaru-alpha/isucon12-q.git
git add . && git commit -m "init"
git branch -M master main
git push -u origin main
```

## 調査

#### VMの状態確認
```sh
arch
free -h
fgrep 'cpu cores' /proc/cpuinfo | sort -u | sed 's/.*: //'
systemctl list-unit-files --type=service
```

#### DBのバージョンとスキーマの確認
```sh
mysql --version
```

```sh
mysqldump -uisucon -pisucon --host 127.0.0.1 --port 3306 ${DATABASE} --compact --no-data --compact --no-data | grep -v "^SET" | grep -v "^/\*\!" | perl -ple 's@CREATE TABLE @\nCREATE TABLE @g';
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

#### Nginx - アクセスログ

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

#### Go - ログ

```
sudo touch /var/log/go.log 
sudo chmod 777 /var/log/go.log
```

```go
import (
	"log"

    log2 "github.com/labstack/gommon/log"
)

func main() {
	log.SetFlags(log.Lshortfile)
	logfile, err := os.OpenFile("/var/log/go.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
	defer logfile.Close()
	log.SetOutput(logfile)
	log.Print("main!!!!")
	e.Logger.SetOutput(logfile)
	e.Logger.SetLevel(log2.ERROR)
}
```

## Go

#### Ubuntu環境にインストール
```
wget https://go.dev/dl/go1.19.linux-amd64.tar.gz -O go.tar.gz
sudo tar -C /usr/local -xzf go.tar.gz
sudo rm -rf go.tar.gz

sudo cat <<EOL >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin
EOL
source ~/.bashrc
```

cf. https://go.dev/dl/

#### Build
```
go tool dist list
GOOS=linux GOARCH=arm64 go build -o isucon *.go
```

#### SingleFlight

```go
package main

import (
	"log"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

var flight singleflight.Group

func main() {
	//　同一 name が処理中なら一緒に結果を待つ
	v, err, _ := flight.Do("group1", func() (interface{}, error) {
		// 時間がかかる処理
		return time.Now(), nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

#### httpを広げる

```go
func main() {
  http.DefaultTransport.(*http.Transport).MaxIdleConns = 0
  http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 1024
  http.DefaultTransport.(*http.Transport).ForceAttemptHTTP2 = true
}
```

#### シリアライザを切り替える

- goccy/go-json
```go
import (
    "github.com/goccy/go-json"
    "github.com/labstack/echo/v4"
)

type JSONSerializer struct{}

func (j *JSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
    return json.NewEncoder(c.Response()).Encode(i)
}

func (j *JSONSerializer) Deserialize(c echo.Context, i interface{}) error {
    return json.NewDecoder(c.Request().Body).Decode(i)
}

func main() {
    e := echo.New()
    e.JSONSerializer = &JSONSerializer{}
}
```

- bytedance/sonic
```go
import (
    "github.com/bytedance/sonic/decoder"
    "github.com/bytedance/sonic/encoder"
    "github.com/labstack/echo/v4"
)

type JSONSerializer struct{}

func (j *JSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
    buf, err := encoder.Encode(i, 0)
    if err != nil {
        return err
    }
    _, err = c.Response().Write(buf)
    return err
}

func (j *JSONSerializer) Deserialize(c echo.Context, i interface{}) error {
    var buf bytes.Buffer
    buf.ReadFrom(c.Request().Body)
    return decoder.NewDecoder(buf.String()).Decode(i)
}

func main() {
    e := echo.New()
    e.JSONSerializer = &JSONSerializer{}
}
```

#### オンメモリ戦略

- map[T]

```go
package main

import (
  "time"

  "github.com/patrickmn/go-cache"
)

type Cache[V any] struct {
  cache *cache.Cache
}

func (c *Cache[V]) Get(key string) (V, bool) {
  v, ok := c.cache.Get(key)
  if ok {
    return v.(V), true
  }
  var defaultValue V
  return defaultValue, false
}

func (c *Cache[V]) Set(key string, value V, ttl time.Duration) {
  c.cache.Set(key, value, ttl)
}

func (c *Cache[V]) Flush() {
  c.cache.Flush()
}

type User struct {
  Name string
}

var userCache = Cache[User]{
  cache: cache.New(cache.NoExpiration, cache.NoExpiration),
}

func main() {
  user := User{Name: "karamaru"}
  userCache.Set(user.Name, user, time.Second*1)
  userCache.Get(user.Name)
}
```

- slice[T]

```go
package main

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache[V any] struct {
	cache *cache.Cache
}

func (c *Cache[V]) Get(key string) (V, bool) {
	v, ok := c.cache.Get(key)
	if ok {
		return v.(V), true
	}
	var defaultValue V
	return defaultValue, false
}

func (c *Cache[V]) Set(key string, value V, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

func (c *Cache[V]) Flush() {
	c.cache.Flush()
}

type User struct {
	Name string
}

var userCache = Cache[[]User]{
	cache: cache.New(cache.NoExpiration, cache.NoExpiration),
}

func main() {
	users := []User{{Name: "karamaru"}}
	userCache.Set("users", users, time.Second*1)
}
```

- map1:1
```go
type omIsuT struct {
  M sync.RWMutex
  V map[string]*Isu
}

var omIsu omIsuT

func (o *omIsuT) Get(k string) (*Isu, bool) {
  o.M.RLock()
  v, ok := o.V[k]
  o.M.RUnlock()
  return v, ok
}

func (o *omIsuT) Set(v Isu) {
  o.M.Lock()
  o.V[v.ID] = v
  o.M.Unlock()
}

func main() {
  omIsu = omIsuT{
    V: map[string]&Isu{}
  }
} 
```

- map1:N

```go
type omIsuListT struct {
	M sync.RWMutex
	V map[string][]*Isu
}

var omIsuList omIsuListT

func (o *omIsuListT) Get(k string) ([]*Isu, bool) {
	o.M.RLock()
	v, ok := o.V[k]
	o.M.RUnlock()
	return v, ok
}

func (o *omIsuListT) Set(k string, v []*Isu) {
	o.M.Lock()
	o.V[k] = append(o.V[k], v...)
	o.M.Unlock()
}

func main() {
	omIsuList = omIsuListT{
		V: map[string][]&Isu{} // make(map[string][]Isu, len(hogehoge))
	}
}
```

- slice

```go
type omIsuListT struct {
	M sync.RWMutex
	V []*Isu
}

var omIsuList omIsuListT

func (o *omIsuListT) Get() ([]*Isu) {
	o.M.RLock()
	defer o.M.RUnlock()
	return o.V
}

func (o *omIsuListT) Set(v []*Isu) {
	o.M.Lock()
	o.V = append(o.V, v...)
	o.M.Unlock()
}

func main() {
	omIsuList = omIsuListT{}
}
```

- slice期限付き

```go
type omIsuListT struct {
	M sync.RWMutex
	T time.Time
	V []*Isu
}

var omIsuList omIsuListT

func (o *omIsuListT) Get() ([]*Isu, bool) {
	o.M.RLock()
	defer o.M.RUnlock()
	if o.T.After(time.Now()) {
		return o.V, true
	}
	return nil, false
}

// 完全置換+期限伸ばす
func (o *omIsuListT) Set(v []*Isu) {
	o.M.Lock()
	o.T = time.Now().Add(time.Second * 1) // キャッシュ時間
	o.V = v
	o.M.Unlock()
}

func main() {
	omIsuList = omIsuListT{
		T: time.Now(),
	}
}
```

#### 画像をfileで読み書き

```go
// init 時の掃除&ディレクトリ設置
func initialize() {
	if err = os.RemoveAll(iconFilePath); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
	if err := os.MkdirAll(iconFilePath, os.ModePerm); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
}

// 書き込み
func write() {
	for _, v := range isuImages {
		if err := os.WriteFile(fmt.Sprintf("%s/%s_%s", iconFilePath, v.JIAUserID, v.JIAIsuUUID), v.Image, os.ModePerm); err != nil {
			c.Logger().Error(err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}
}

// 読み込み
func read() {
	image, err := os.ReadFile(fmt.Sprintf("%s/%s_%s", iconFilePath, jiaUserID, jiaIsuUUID))
}

// コピー(読み込みより早い)
func copy() {
	file, err := os.Open(defaultIconFilePath)
	defer file.Close()

	f, err := os.Create(fmt.Sprintf("%s/%s_%s", iconFilePath, jiaUserID, jiaIsuUUID))
	defer f.Close()

	_, err = io.Copy(f, file)
}
```

cf. https://github.com/narusejun/isucon11-qualify/commit/e3cc31346fb89455a0f0123e9ea08156914e28c4

cf. https://github.com/tatsumack/isu11q/commit/ae2da92c4a11b184c9fb479f7cde39e935140baf

#### 一定時間毎に処理をする

```go
func loop() {
	for range time.Tick(time.Second) {
		// something to do
	}
}
```

#### UnixDomainSocket

```go
func main() {
  socket_file := "/home/isucon/webapp/tmp/app.sock"
  os.Remove(socket_file)
  
  l, err := net.Listen("unix", socket_file)
  if err != nil {
  e.Logger.Fatal(err)
  }
  
  // go runユーザとnginxのユーザ（グループ）を同じにすれば777じゃなくてok
  err = os.Chmod(socket_file, 0777)
  if err != nil {
  e.Logger.Fatal(err)
  }
  
  e.Listener = l
  e.Logger.Fatal(e.Start(""))
  // http.Serve(l, mux)
}
```

```conf
upstream s1 {
  server unix:/home/isucon/webapp/tmp/app.sock;
  keepalive 32;
  keepalive_requests 10000;
}

location /api {
  proxy_http_version 1.1;
  proxy_set_header Connection "";
  proxy_pass   http://s1;
}
```

```
mkdir tmp
-> gitignore
```

cf. https://github.com/narusejun/isucon11-qualify/commit/3d9f96bfe12f263a0ea8f3aa759b8e73c2659f0a

## Mysql (MariaDB)

#### DB用のインスタンスのセットアップ

```
sudo -u -i isucon
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update
sudo apt install -y gh
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

```
sudo cat <<EOL >> ~/.bashrc
alias cm="git commit -m"
alias ad="git add ."
alias co="git checkout"
alias save="git add . && git commit -m"
alias cob="git checkout -b"
alias mg="git merge"
alias rename="git branch -m"
alias del="git branch -D"
alias refresh="git checkout . && git clean -df"
EOL
source ~/.bashrc
```

#### 外部からのアクセスを許容する

- goのmysql.openとsqlのinit.shもHOSTをプライベートアドレスに変更する
- '%'のユーザー作成も忘れずに！

```sh
# MariaDBの場合
/etc/mysql/mariadb.conf.d/*-server.cnf
bind-address = 0.0.0.0
```

#### ユーザの作成
```sql
SELECT user, host FROM mysql.user;
DROP USER 'isucon'@'%';
CREATE USER 'isucon'@'%' IDENTIFIED BY 'isucon';
GRANT ALL PRIVILEGES ON * . * TO 'isucon'@'%';
FLUSH PRIVILEGES;
```

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

#### デッドロックの調査
```sql
SHOW ENGINE INNODB STATUS;
```

#### TroubleShoot

- Unknown collation: 'utf8mb4_0900_ai_ci'
  - sed -i 's/utf8mb4_0900_ai_ci/utf8mb4_unicode_ci/g' sql/dump.sql

#### Mysqlの起動を待つ

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

#### bulkInsert

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

#### IN句

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

#### コネクションプール

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

#### DBの起動を待つ

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

#### generatedColumns

```sql
popularity INT NOT NULL,
popularity_desc INT AS (-popularity) INVISIBLE, -- index 使う時は STORED
```


#### 1byte長の半角文字列をピッタリ格納する

```sql
`id` CHAR(26) CHARACTER SET latin1,
```


#### UUIDをBINARY(16)で格納する

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

#### 水平分割

cf. https://github.com/narusejun/isucon11-qualify/commit/fda74ca7e56b70a58a7a49c773cb892d3dae6765

#### Upsert

```sql
INSERT INTO `user` (`id`, `name`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `name`=?;
-- bulk
INSERT INTO `user` (`id`, `name`) VALUES (:id, :name) ON DUPLICATE KEY UPDATE `name`=VALUES(`name`);
```

#### trigger

```sql
DROP TRIGGER IF EXISTS tr1;
CREATE TRIGGER tr1 BEFORE INSERT ON playlist_favorite FOR EACH ROW INSERT INTO playlist_favorite_count (playlist_id,count) VALUES (NEW.playlist_id, 1) ON DUPLICATE KEY UPDATE playlist_favorite_count.count = playlist_favorite_count.count + 1;
```

#### Group毎に最新のレコードをSELECTする

```sql
SELECT * FROM isu_condition AS a JOIN (SELECT user_id, MAX(created_at) AS created_at FROM isu_condition GROUP BY user_id) AS b
ON a.user_id = b.user_id WHERE a.created_at = b.created_at;
```


#### アドバイザリロック

```go
func main() {
	db.Exec("SELECT GET_LOCK(?, ?)", id, 5)
	defer db.Exec("SELECT RELEASE_LOCK(?)", id)
}
```

## Nginx

#### インストール

```bash
sudo apt update
sudo apt install nginx
sudo ufw allow 'Nginx Full'
sudo systemctl enable nginx
systemctl list-unit-files --type=service
```

#### keepaliveを有効する

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

#### ファイル上限を確認・拡張する

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

#### 静的ファイルのクライアントキャッシュ

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

#### レスポンスキャッシュ(ProxyCache)

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


#### 認証付きの静的配信(X-Accel-Redirect)

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

#### リクエストメソッドでリダイレクト

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

#### パスパラメータでリダイレクト

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

#### クエリパラメータでリダイレクト

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

#### URLの部分一致でリダイレクト

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

#### Botからのリクエストを拒否

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

#### 圧縮 / 事前圧縮

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

## Linux

#### ファイルディスクリプタ上限up

```
sudo vi /etc/security/limits.conf
*                hard   nofile           10240
*                soft   nofile           10240

ulimit -n
```

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

# EnvironmentFile=/home/isucon/env.shcm

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

#### shellの変更

sudo権限があるuserで
```shell
sudo usermod -s /bin/bash isucon
```

#### githubの鍵でssh

```shell
cat ~/.ssh/id_rsa.pub | pbcopy

sudo chmod 0700 ~/.ssh
sudo vim ~/.ssh/authorized_keys
sudo chmod 0600 ~/.ssh/authorized_keys
sudo systemctl restart sshd.service
```

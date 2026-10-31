# Go の解法

> isucon-memo の README.md から抜き出したもの。スニペットは雛形なので、そのまま貼らずに対象アプリの型名・テーブル名・ライブラリに合わせて書き直すこと。

## 汎用キャッシュ Cacher[T]（isucon-memo/main.go）

TTL 付きの RWMutex map。ttl に 0 以下を渡すと無期限。

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

const CACHE_KEY = "CACHE_KEY"

type Cacher[T any] struct {
	Mutex sync.RWMutex
	Cache map[string]struct {
		Value   T
		Expired time.Time
	}
}

func (c *Cacher[T]) Get(key string) (T, bool) {
	c.Mutex.RLock()
	cache, ok := c.Cache[key]
	c.Mutex.RUnlock()
	if ok && (cache.Expired.IsZero() || time.Now().Before(cache.Expired)) {
		return cache.Value, true
	}
	var defaultValue T
	return defaultValue, false
}

func (c *Cacher[T]) GetAll() []T {
	c.Mutex.RLock()
	slice := make([]T, 0, len(c.Cache))
	for _, v := range c.Cache {
		slice = append(slice, v.Value)
	}
	c.Mutex.RUnlock()
	return slice
}

func (c *Cacher[T]) Set(key string, value T, ttl time.Duration) {
	c.Mutex.Lock()
	var expired time.Time
	if ttl > 0 {
		expired = time.Now().Add(ttl)
	}
	c.Cache[key] = struct {
		Value   T
		Expired time.Time
	}{
		Value:   value,
		Expired: expired,
	}
	c.Mutex.Unlock()
}

func (c *Cacher[T]) Delete(key string) {
	c.Mutex.Lock()
	delete(c.Cache, key)
	c.Mutex.Unlock()
}

func (c *Cacher[T]) Flush() {
	c.Mutex.Lock()
	c.Cache = make(map[string]struct {
		Value   T
		Expired time.Time
	})
	c.Mutex.Unlock()
}

func InitCacher[T any]() Cacher[*T] {
	return Cacher[*T]{
		Cache: make(map[string]struct {
			Value   *T
			Expired time.Time
		}, 0),
	}
}

type Thing struct {
	Name string
}

var ThingCacher = InitCacher[Thing]()

func main() {
	isu := &Thing{Name: "isu"}
	ThingCacher.Set("isu", isu, -1)
	fmt.Println(ThingCacher.Get("isu"))
}
```

## Go - pprof

```go
package main

import (
  "log"
  "net/http"
  _ "net/http/pprof"
)

func main() {
  go func() {
    log.Fatal(http.ListenAndServe(":6060", nil))
  }()
}
```

```shell
sudo apt install -y graphviz
go tool pprof -http=0.0.0.0:1080 http://localhost:6060/debug/pprof/profile\?seconds=60
```

-> http://localhost:6060/debug/pprof

## Build
```
go tool dist list
GOOS=linux GOARCH=arm64 go build -o isucon *.go
```

### CPU の命令セットレベルを上げる（GOAMD64）

Go は amd64 向けにデフォルトで `GOAMD64=v1`（全 x86-64 CPU で動く最低限の命令）でビルドする。`v3` にすると AVX2・BMI2・LZCNT・FMA などを使ったコードを生成する。

| レベル | v1 から増える主な命令 |
| --- | --- |
| v1（デフォルト） | — |
| v2 | POPCNT、SSE3、SSE4.1/4.2、SSSE3、CMPXCHG16B |
| v3 | v2 + AVX、AVX2、BMI1/2、F16C、FMA、LZCNT、MOVBE |
| v4 | v3 + AVX512 系（ただし現在の Go は AVX512 命令を生成しないので、v3 との差はほぼない） |

```sh
# サーバーの CPU が v3 に対応しているか確認する（glibc 2.33 以降）
/lib64/ld-linux-x86-64.so.2 --help | grep 'x86-64-v'
# => x86-64-v3 (supported, searched) なら OK

# ビルド（Makefile の go build にも付ける）
GOAMD64=v3 go build -o isucon
```

効果の目安:

- 「2 割速くなる」は、protobuf-go の `SizeVarint` という 1 関数のマイクロベンチの数字。BSR → LZCNT への置き換えで Zen 2 では約 -50%、コードの書き換えも合わせて Skylake で約 -19% になった。コミットの作者自身が「実際の Protobuf 利用ではこの変更単体で計測できるほどの改善はない」と書いている（protocolbuffers/protobuf-go@8e8926e）。
- ISUCON のアプリは DB 待ちや I/O 待ちが支配的なので、スコアへの効果は数 % 以内か誤差になると見ておく。ビット演算・ハッシュ・JSON・画像処理などで Go の CPU が張り付いているときほど効く可能性がある。
- 公式も「上位レベルで速くなることはあるが保証はない。ベンチで確かめること」としている（go.dev/wiki/MinimumRequirements#amd64）。

注意:

- 対応していない CPU で動かすと、起動直後に `This program can only be run on AMD64 processors with v3 microarchitecture support.` で落ちる。ビルドしたマシンではなく、**実行するサーバー**の CPU で確認する。
- `GOARCH=arm64`（Graviton など）のサーバーでは `GOAMD64` は効かない。arm64 には `GOARM64`（Go 1.23 以降、デフォルト `v8.0`）がある。
- 他の施策と同じく 1 ベンチ 1 施策で入れ、改善しなければ戻す。変更が 1 行で済むので、上位のボトルネックを潰し終えて Go の CPU 使用率が高い段階で試すとよい。

## SingleFlight

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

## httpを広げる

```go
func main() {
  http.DefaultTransport.(*http.Transport).MaxIdleConns = 0
  http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 1024
  http.DefaultTransport.(*http.Transport).ForceAttemptHTTP2 = true
}
```

## シリアライザを切り替える

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

## オンメモリ戦略

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

## 画像をfileで読み書き

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

## 一定時間毎に処理をする

```go
func loop() {
	for range time.Tick(time.Second) {
		// something to do
	}
}
```

## UnixDomainSocket

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

## httpリクエスト

```go
func main() {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://172.31.5.58:5000/initialize/isu3", nil)
    if err != nil {
    panic(err)
    }
    _, err = http.DefaultClient.Do(req)
    if err != nil {
    return err
}
```


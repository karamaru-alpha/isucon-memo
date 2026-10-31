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

---

<!-- ここから下は ISUCON10〜14 の上位解法・公式講評の調査から追記したもの -->

## デバッグモード・ログ出力を止める

stdoutへの大量ログはjournaldのCPUを食う。初手で確認し、終盤はslow_query_log・nginxのaccess_logも切る (10予選, 11予選)
```go
// e.Use(middleware.Logger()) を削除
e.Debug = false
e.Logger.SetLevel(log.ERROR)
```

## 非同期でバルクINSERTする

即時に読まれない書き込み(ログ・計測値)はchannelに溜めて202を返し、定期的にまとめてINSERTする (11予選 isu_condition, 14 chair_locations)
```go
var insertCh = make(chan []any, 10000)

func insertWorker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	var ph []string
	var args []any
	for {
		select {
		case v := <-insertCh:
			ph = append(ph, "(?, ?, ?)")
			args = append(args, v...)
		case <-ticker.C:
			if len(ph) == 0 {
				continue
			}
			if _, err := db.Exec("INSERT INTO isu_condition (jia_isu_uuid, timestamp, message) VALUES "+strings.Join(ph, ","), args...); err != nil {
				log.Print(err)
			}
			ph, args = ph[:0], args[:0]
		}
	}
}

// handler: insertCh <- []any{uuid, ts, msg}; return c.NoContent(http.StatusAccepted)
```
cf. https://github.com/narusejun/isucon11-qualify , https://blog.agatan.dev/posts/isucon11-qualify

## マイクロキャッシュ(短TTL + SingleFlight)

引数が定数の重い集計は、singleflightで同時実行を1本にまとめた上で数百ms保持する (11本選 GPA)
```go
var (
	sf       singleflight.Group
	mu       sync.RWMutex
	cached   []float64
	cachedAt time.Time
)

func getGPAs() ([]float64, error) {
	mu.RLock()
	if time.Since(cachedAt) < 900*time.Millisecond {
		defer mu.RUnlock()
		return cached, nil
	}
	mu.RUnlock()
	v, err, _ := sf.Do("gpa", func() (any, error) {
		start := time.Now() // 計算開始時刻を保存して保守的に
		var r []float64
		if err := db.Select(&r, query); err != nil {
			return nil, err
		}
		mu.Lock()
		cached, cachedAt = r, start
		mu.Unlock()
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]float64), nil
}
```
cf. https://github.com/fujiwara/isucon11-f

## 圧縮済みレスポンスを事前生成する

「N秒古くてよい」レスポンスはtickerで作り、gzip済みのバイト列で持ってそのまま返す (10本選 ダッシュボード)
```go
var dashboard atomic.Pointer[[]byte]

func dashboardWorker() {
	for range time.Tick(500 * time.Millisecond) {
		res, _ := json.Marshal(makeDashboard())
		var buf bytes.Buffer
		w, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
		w.Write(res)
		w.Close()
		b := buf.Bytes()
		dashboard.Store(&b)
	}
}

func getDashboard(c echo.Context) error {
	c.Response().Header().Set("Content-Encoding", "gzip")
	return c.Blob(http.StatusOK, "application/json", *dashboard.Load())
}
```
cf. https://kazeburo.hatenablog.com/entry/2020/10/14/111149

## 状態変化をchannelで通知する

ロングポーリング/SSEで状態変化を待つとき、DBを定期的に見に行かず更新側からシグナルを送る (14 chair通知)
```go
var sigs sync.Map // id -> chan struct{}

func register(id string) chan struct{} {
	ch := make(chan struct{}, 1) // バッファ1で取りこぼさない
	sigs.Store(id, ch)
	return ch
}
func unregister(id string, ch chan struct{}) { sigs.CompareAndDelete(id, ch) }

// 状態を更新した箇所で呼ぶ
func signal(id string) {
	if v, ok := sigs.Load(id); ok {
		select {
		case v.(chan struct{}) <- struct{}{}:
		default: // 既に通知待ちなら捨てる
		}
	}
}

// 待つ側: select { case <-ctx.Done(): case <-ch: case <-time.After(time.Second): /* 保険 */ }
```
cf. https://github.com/takonomura/isucon14

## Server-Sent Events

ポーリングをSSEにする。nginxのバッファリングを `X-Accel-Buffering: no` で切るのを忘れない (14)
```go
func sse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	rc := http.NewResponseController(w)
	ch := register(id)
	defer unregister(id, ch)
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			b, _ := json.Marshal(getNotification(id))
			fmt.Fprintf(w, "data: %s\n\n", b)
			if err := rc.Flush(); err != nil {
				return
			}
		}
	}
}
```
- 公式講評では「上位に必須ではなかった」。まずは `retry_after_ms` を状態に応じて伸ばす方が安い
cf. https://isucon.net/archives/58869617.html

## キー単位のロック

flock(ファイルロック)やアプリ全体のMutexはキー単位のMutexにする。複数台ならDBのトランザクション/条件付きUPDATEにする (12予選)
```go
var locks sync.Map // key -> *sync.Mutex

func lock(key string) func() {
	m, _ := locks.LoadOrStore(key, &sync.Mutex{})
	m.(*sync.Mutex).Lock()
	return m.(*sync.Mutex).Unlock
}

// defer lock(tenantID)()
```

## ID採番をアプリでする

`REPLACE INTO id_generator` のような採番テーブルは書き込みの直列化点になる (12予選・本選)
```go
// 文字列でよいなら
id := uuid.NewString()

// 数値で1台なら AUTO_INCREMENT + res.LastInsertId()

// 数値で複数台ならSnowflake (github.com/bwmarrin/snowflake)
node, _ := snowflake.NewNode(nodeID) // 台ごとに変える
id := node.Generate().Int64()
```

## ユーザーID単位のシャーディング

- 下位ビットはシーケンス番号でほぼ0なので、Snowflake IDをそのまま `% N` すると偏る
- セッションIDにユーザーIDを埋め込むと、セッション検証時に引くべきシャードが分かる
- `/initialize` は全DBホストに並列で配る
```go
func selectDB(userID int64) *sqlx.DB { return dbs[int(userID>>23)%len(dbs)] }

// 発行時
sessID := fmt.Sprintf("%s::%d", uuid.NewString(), userID)
// 検証時
parts := strings.Split(c.Request().Header.Get("x-session"), "::")
uid, _ := strconv.ParseInt(parts[1], 10, 64)
selectDB(uid).Get(&sess, "SELECT * FROM user_sessions WHERE session_id = ?", c.Request().Header.Get("x-session"))
```
cf. https://github.com/narusejun/isucon12-final/blob/master/app/webapp/go/db.go

## ETag / If-None-Matchで304を返す

画像はハッシュを保存しておき、一致したらボディを返さない。ヘッダ値はダブルクォート付き (13 アイコン)
```go
etag := `"` + iconHash + `"`
if c.Request().Header.Get("If-None-Match") == etag {
	return c.NoContent(http.StatusNotModified)
}
c.Response().Header().Set("ETag", etag)
return c.Blob(http.StatusOK, "image/jpeg", image)
```
- ハッシュはINSERT時に計算する(triggerなら `SET NEW.icon_hash = SHA2(NEW.image, 256)`)
cf. https://isucon.net/archives/58001272.html

## 外部コマンドを呼ばない・ファイルはストリームで扱う

- `exec.Command("zip", ...)` は `archive/zip` の `zip.Store`(無圧縮)にする。`cp` はハードリンクにする (11本選)
- アップロードは `io.ReadAll` せず `io.Copy` でファイルに書く
```go
zw := zip.NewWriter(f)
fw, _ := zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
src, _ := os.Open(path)
io.Copy(fw, src)
src.Close()
zw.Close()
c.Response().Header().Set("X-Accel-Redirect", "/assignments/"+classID+".zip")
```
cf. https://isucon.net/archives/56163308.html

## アロケーション・GCを減らす

pprofで `mallocgc` / GCが上位に来たとき (12本選)
```go
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

buf := bufPool.Get().(*bytes.Buffer)
buf.Reset()
defer bufPool.Put(buf)
```
```
# systemdのEnvironmentに書く(ビルド時ではなく実行時の環境変数)
Environment=GOGC=2000
```

## PGO

pprofで取ったCPUプロファイルを使ってビルドする。終盤にノーリスクで入れられる (13, 14)
```sh
curl -o default.pgo "http://localhost:6060/debug/pprof/profile?seconds=60" # ベンチ中に取得
go build -pgo=default.pgo -o isucon
```

## SQLに呼び出し元をコメントで付ける

スロークエリを見てもどのハンドラか分からない問題を解消する (14 LayerX)
```go
func withCaller(query string) string {
	_, file, line, _ := runtime.Caller(1)
	// pt-query-digestは数字を?に正規化するので行番号は全角にする
	return fmt.Sprintf("/* %s:%s */ %s", filepath.Base(file), toFullWidth(line), query)
}

func toFullWidth(n int) string {
	return strings.Map(func(r rune) rune { return r - '0' + '０' }, strconv.Itoa(n))
}

// db.Select(&users, withCaller("SELECT * FROM users WHERE id = ?"), id)
```
cf. https://zenn.dev/layerx/articles/3bd55b77e047c4

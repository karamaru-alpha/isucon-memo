# DNS の解法

> ISUCON10〜14 の上位解法・公式講評の調査から追記したもの。スニペットは雛形なので、対象アプリに合わせて書き直すこと。

ISUCON13でPowerDNS(MySQL backend)が出題範囲に入り、ベンチが存在しないサブドメインを大量に問い合わせてきた(水責め)。

## TTLとキャッシュ

初期状態はTTL=0・pdnsのキャッシュ無効で、名前解決のたびにSQLが走る
```
# ゾーンファイル: 固定レコードのTTLを伸ばす
pipe     60 IN A  <ISUCON_SUBDOMAIN_ADDRESS>
# pdns.conf
query-cache-ttl=60
negquery-cache-ttl=60
```
```sql
-- PowerDNSのrecordsにインデックスが無い
ALTER TABLE records ADD INDEX idx_name (name);
```
- ワイルドカードレコード(`*`)にすると水責めのクエリも解決に成功し、後段のHTTPSにアクセスが来て逆効果

cf. https://kazeburo.hatenablog.com/entry/2023/12/02/235258

## dnsdistで水責めを弾く

pdnsを1053番などに移し、dnsdistを53番で前段に置く
```lua
addLocal("0.0.0.0:53", {reusePort=true})
newServer({address="127.0.0.1:1053", name="pdns"})
newServer({address="127.0.0.1:1053", name="abuse", pool="abuse"})
setACL({"0.0.0.0/0", "::/0"})
-- ラベル数が想定外ならDrop
addAction(QNameLabelsCountRule(0, 4), DropAction())
-- 長すぎる名前は2qpsのプールへ
addAction(QNameWireLengthRule(0, 32), QPSPoolAction(2, "abuse"))
-- NXDOMAINの応答を遅らせると攻撃側のQPSが落ちる
addResponseAction(RCodeRule(DNSRCode.NXDOMAIN), DelayResponseAction(1000))
```
cf. https://diary.sorah.jp/2023/11/28/isucon13 , https://github.com/shirokanezoo/isucon13/blob/main/dnsdist.conf

## DNSサーバーを自作する

登録済みサブドメインをオンメモリで持ち、アプリ内でDNSを返す。未知の名前には応答しない (13優勝)
```go
import "github.com/miekg/dns"

var subdomains sync.Map // "name.u.isucon.dev." -> struct{}

func startDNS() {
	dns.HandleFunc("u.isucon.dev.", func(w dns.ResponseWriter, r *dns.Msg) {
		name := r.Question[0].Name
		if _, ok := subdomains.Load(name); !ok {
			return // 無応答
		}
		m := new(dns.Msg)
		m.SetReply(r)
		rr, _ := dns.NewRR(name + " 120 IN A " + os.Getenv("ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS"))
		m.Answer = []dns.RR{rr}
		w.WriteMsg(m)
	})
	go (&dns.Server{Addr: ":53", Net: "udp"}).ListenAndServe()
}
// ユーザー登録時に subdomains.Store、/initialize で全件読み直す
```
cf. https://zenn.dev/tohutohu/articles/923bdf5dcd73af , https://github.com/narusejun/isucon13-final/blob/main/app/go/dns.go

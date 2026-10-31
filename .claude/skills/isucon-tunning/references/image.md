# 画像・静的ファイル

> ISUCON3本選・7予選・11予選・13・private-isu の上位解法・参加記・公式講評の調査から追記したもの。画像（アイコン・投稿画像）が kataribe の上位にいる、または帯域が詰まっているときに読む。
> 基本のスニペットは go.md「画像をfileで読み書き」「ETag / If-None-Matchで304を返す」、nginx.md「認証付きの静的配信(X-Accel-Redirect)」にある。ここでは使い分けと落とし穴をまとめる。

## まず確認すること

- マニュアル: 画像の内容をベンチが検証するか（ハッシュ比較なら再圧縮・リサイズ不可）、条件付き GET（`If-None-Match` / `If-Modified-Since`）を送ってくるか、304 に加点があるか
- kataribe: 画像パスの Sum / Count と、ステータス 304 の割合
- 帯域: ベンチ中に `sar -n DEV 1` で送信量を見る。画像回は外向き NIC が先に詰まることがある (7予選は 100Mbps を使い切っていた)
- DB: 画像が BLOB で入っているなら、MySQL からの転送量だけで数百 Mbps 出ていることがある (7予選 600Mbps)

施策の順番の目安は「BLOB をファイルに出す → nginx から配信 + キャッシュヘッダ → 304 を返す → 複数台の配置を決める」。

## BLOB をファイルに出す

(private-isu, 7予選, 11予選, 13)

- 初期データの画像は `/initialize`（または起動時）にファイルへ書き出す。書き出し先は `/initialize` で一度消してから作り直す
- POST（アップロード・アイコン更新）はファイルに書き、DB には入れない。DB にはハッシュやパスだけ残す
- 画像カラムを持つテーブルを `SELECT *` していないか確認する。一覧系のクエリが BLOB を読み捨てていることが多い (private-isu の `imgdata`)。カラムを明示するか、画像を別テーブル・ファイルに追い出す
- 書き出しは go.md「画像をfileで読み書き」のスニペットを使う

### 書き出していない画像はアプリに回す

全件を事前に書き出す時間がないときや、書き出し漏れの保険として、nginx で見つからなければアプリに回し、アプリが DB から読んでファイルに保存してから返す (private-isu)。

```nginx configuration
location /image/ {
    root /home/isucon/private_isu/webapp/public/;
    expires 1d;
    try_files $uri @app;
}
location @app {
    internal;
    proxy_pass http://app;
}
```

- `try_files` は 403 ではフォールバックしない。アプリが作ったファイルやディレクトリを nginx ユーザーが読めないと、画像が 403 になり続ける。パーミッションを確認する
- ファイル名に拡張子を付けておくと、nginx が `mime.types` から `Content-Type` を決められる

## nginx から配信してキャッシュヘッダを付ける

(7予選, 11予選)

```nginx configuration
location /icons/ {
    root /home/isucon/webapp/public;
    add_header Cache-Control "public, max-age=86400";
    # expires 1d; でもよいが、7予選では public を付けたことで 304 が大きく増えた
}
```

- 7予選では `expires` だけだと 304 が増えず、`Cache-Control: public` に変えたら 304 が大幅に増えてスコアが約 8 万点に伸びた。`immutable` を足しても効果はなかった (kazeburo)
- 7予選では、最初のレスポンスに `Last-Modified` が付いていなかったため 304 が返らなかった例もある。tcpdump や `curl -I` で、最初の 200 に `Last-Modified` / `ETag` が付いているか確認する
- JS / CSS / フォントにも付けるとスコアが落ちたという報告がある (7予選 sorah、後に付け直して上昇)。付ける範囲を広げたら必ずベンチで確かめる
- 認証が必要な画像は nginx.md「認証付きの静的配信(X-Accel-Redirect)」で、認証だけアプリ、送信は nginx にする

## アプリ独自のハッシュで 304 を返す

(13)

ベンチが「API レスポンスに含まれるハッシュ」を `If-None-Match` に入れて送ってくる場合、nginx が自動で付ける ETag（`"mtime-size"` 形式）とは一致しない。304 の判定はアプリで行う。

- ハッシュは毎回 SHA256 を計算せず、更新時に計算して保存する（DB の `icon_hash` カラム、トリガー、またはプロセス内の map）。go.md「ETag / If-None-Matchで304を返す」参照
- アイコン未設定ユーザーに返す NoImage 画像のハッシュは起動時に 1 回だけ計算しておく
- `If-None-Match` の値はダブルクォート付き。比較する前に外すか、`"` を付けたハッシュと比較する
- 200 を返すときも ETag ヘッダを付ける
- 13 では「アイコン更新から 2 秒以内にハッシュを反映すること」という条件があった。キャッシュを入れるなら反映猶予をマニュアルで確認する
- ユーザー一覧などで返すハッシュのために、毎回 BLOB を読んでいないかも確認する (13 の `fillUserResponse`)

`If-None-Match` 付きだけアプリに回し、それ以外は nginx から返す構成にすると、アプリは比較だけで済む。200 の本体は X-Accel-Redirect で nginx に送らせる (13 白金動物園)。

```nginx configuration
location ~ ^/api/user/(?<username>[^/]+)/icon$ {
    # If-None-Match 付きはアプリへ。アプリが 304 か X-Accel-Redirect 付き 200 を返す
    error_page 418 = @app;
    if ($http_if_none_match) { return 418; }
    try_files /icons/$username.jpg @app;   # root 配下に書き出したファイル
}
location @app {
    proxy_pass http://app;
}
```

- location 内の `if` で安全に使えるのは `return` と `rewrite ... last` だけ。`if` の中に `proxy_pass` を書いて `try_files` と並べると期待どおりに動かないので、`return 418` + `error_page` で named location に回す
- `nginx -t` で確認してからデプロイする
- 白金動物園は、アプリ側の比較処理の CPU が残り、効果は限定的だったと振り返っている。ハッシュはメモリから引けるようにしておく

## 複数台で画像をどこに置くか

(3本選, 7予選)

| 方式 | 内容 | 向いているとき |
| --- | --- | --- |
| 1 台に集約 | 画像の書き込み（POST）と配信を 1 台に寄せ、他の台の nginx はその台へ proxy する。内向き帯域は 1Gbps 程度あるので詰まりにくい | 画像の更新が少ない、構成を単純にしたい (3本選優勝チーム、7予選 sorah) |
| 各台で保存 + フォールバック | アップロードを受けた台に保存し、`try_files $uri @other;` で無ければ他の台へ proxy する | 画像の POST が多く 1 台に寄せると重い (7予選) |
| 共有ストレージ | WebDAV や NFS で 1 台に集める | 既存の書き込み経路を変えたくない (7予選 kazeburo は WebDAV) |

- どの方式でも、`/initialize` で全台の画像ディレクトリを初期状態に戻す
- POST を 1 台に寄せる場合は nginx.md「リクエストメソッドでリダイレクト」でメソッドごとに振り分ける
- アップロードが大きい場合は nginx.md「リクエストボディのバッファ」で `client_body_buffer_size` を広げる (7予選は `/profile` だけ 5000k)

## リサイズ・変換が必要な問題

(3本選)

- 3本選の優勝チームは、既存画像を競技中に事前リサイズし、新しい画像はアップロード時にリサイズしてから保存した。リクエストのたびに `convert` を呼ぶ方式をやめている
- 非同期でリサイズする案は、反映までの不整合が難しく途中で捨てている
- 画像の品質を下げてファイルサイズを減らす手は、ベンチが画像の内容を検証しない場合に限る。最近の回は内容を検証するので、まずマニュアルを確認する
- 外部コマンドを呼んでいる場合は、Go の `image` パッケージに置き換える。go.md「外部コマンドを呼ばない・ファイルはストリームで扱う」参照

## 画像を gzip しない

JPEG / PNG / GIF はすでに圧縮されているので、gzip しても小さくならず CPU だけ使う。`gzip_types` に `image/jpeg` `image/png` `image/gif` を入れない（SVG は効く）。

cf. https://isucon.net/archives/58001272.html (13講評)
cf. https://diary.sorah.jp/2023/11/28/isucon13 (13 白金動物園)
cf. https://kazeburo.hatenablog.com/entry/2017/10/23/181843 (7予選)
cf. https://diary.sorah.jp/2017/10/23/isucon7q (7予選)
cf. https://blog.nomadscafe.jp/2013/11/isucon3-isucon.html (3本選優勝)
cf. https://github.com/kuri0616/isucon-private/pull/4 (private-isu)

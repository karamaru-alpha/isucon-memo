---
name: isucon-deploy
description: ISUCONの競技サーバー(isu1〜isu3)にsshして `make isu${N}` を実行し、GitHub(karamaru-alpha/isucon2026)のGoアプリをデプロイしてログをローテートする。「デプロイして」「isu1に反映」「isucon-deploy」「サーバーに上げて」などの言葉が出たら使う。
---

# isucon-deploy

ローカルの変更を GitHub 経由で競技サーバーに反映し、計測用ログを空にして次のベンチに備える。

## 前提

| 項目 | 値 |
| --- | --- |
| リポジトリ | `karamaru-alpha/isucon2026`（ローカル: `~/github.com/karamaru-alpha/isucon2026`） |
| ssh ホスト | `isu1` `isu2` `isu3`（`~/.ssh/config` のエイリアス） |
| サーバー上の作業ディレクトリ | `/home/isucon/webapp`（Makefile・analyze.sh がある場所。違う場合は Makefile の `APP_PATH` に合わせる） |

`make isu${N}` は「git pull → 設定ファイルのコピー → `go build` → サービス再起動」を行うターゲット（isucon-memo の Makefile テンプレート）。

## 引数

- `N`: デプロイ先。`1` / `2` / `3` / `all`（カンマ区切りも可）。省略したら `1`。

## 手順

### 1. ローカルの状態確認

```sh
cd ~/github.com/karamaru-alpha/isucon2026
git status --short
git rev-parse --abbrev-ref HEAD
git log origin/$(git rev-parse --abbrev-ref HEAD)..HEAD --oneline
```

- 未コミットの差分があるなら、何をコミットするかユーザーに確認してからコミットする。
- 未 push のコミットがあれば push する（サーバーは GitHub から pull するので、push しないと反映されない）。
- デプロイ対象のコミットハッシュ `LOCAL_SHA=$(git rev-parse HEAD)` を控えておく。

### 2. サーバーのブランチ確認

```sh
ssh isu${N} "cd /home/isucon/webapp && git rev-parse --abbrev-ref HEAD"
```

`make isu${N}` はサーバーで今チェックアウトしているブランチを pull する。ローカルのブランチと違う場合は、切り替えてよいかユーザーに確認してから実行する。

```sh
ssh isu${N} "cd /home/isucon/webapp && git fetch origin && git checkout <branch>"
```

### 3. デプロイ

```sh
ssh isu${N} "cd /home/isucon/webapp && make isu${N}"
```

- `all` や複数台を指定された場合は、各台を並列に（1 メッセージで複数の Bash 呼び出しを）実行する。
- `go build` が失敗したら、ここで止めてエラーを報告する。ローテートや再起動には進まない。

### 4. ログのローテート

前回ベンチのログが混ざらないよう、nginx / MySQL / Go のログを空にする。truncate なので、プロセスがファイルを開いたままでも安全。

```sh
ssh isu${N} "cd /home/isucon/webapp && make rotate"
```

`rotate` ターゲットがない場合（`make: *** No rule to make target 'rotate'`）は直接実行する:

```sh
ssh isu${N} "sudo truncate -s 0 /var/log/nginx/access.log /var/log/nginx/error.log /var/log/mysql/slow-query.log /var/log/mysql/error.log /var/log/go.log 2>/dev/null; true"
```

DB を別サーバーに分けている場合は、DB サーバーでも slow-query.log を truncate する。

### 5. 反映確認

```sh
ssh isu${N} "cd /home/isucon/webapp && git rev-parse HEAD && systemctl list-units --type=service --state=running | grep -e isu -e nginx -e mysql -e mariadb"
```

- サーバーの HEAD が `LOCAL_SHA` と一致すること。
- アプリ・nginx・MySQL が `running` であること。落ちていたら `ssh isu${N} "sudo journalctl -u <service> -n 50 --no-pager"` で原因を確認する。
- nginx の設定を変えた場合は `ssh isu${N} "sudo nginx -t"` も確認する。

### 6. 報告

次の内容を短くまとめて報告する:

- デプロイ先ホストとコミット（`LOCAL_SHA` の短縮形と件名）
- ローテートの実施結果
- サービスの状態

最後に「ポータルでベンチを回したら `/isucon-survey` で計測するのだ」と次の行動を案内する。

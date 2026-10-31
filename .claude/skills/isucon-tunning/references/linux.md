# Linux / systemd の解法

> isucon-memo の README.md から抜き出したもの。スニペットは雛形なので、そのまま貼らずに対象アプリの型名・テーブル名・ライブラリに合わせて書き直すこと。

## ファイルディスクリプタ上限up

```
sudo vi /etc/security/limits.conf
*                hard   nofile           10240
*                soft   nofile           10240

ulimit -n
```

## Systemdでアプリを動かす

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

---

<!-- ここから下は ISUCON10〜14 の上位解法・公式講評の調査から追記したもの -->

## 再起動試験でsystemdが止まらないようにする

起動順がずれて再起動を繰り返すとsystemdがunitを諦めて止める (10予選, 11本選, 12本選)
```
sudo systemctl edit isucon.go.service
---
[Unit]
After=mysql.service
StartLimitBurst=999
---
```
- アプリ側でもDBの起動を待つ(mysql.md「DBの起動を待つ」)

# ssh-tunnel-socks5

## Build

```
git clone https://github.com/imaben/ssh-tunnel-socks5.git
cd ssh-tunnel-socks5
go build
```

## Config

```
[remote]
addr="yourhost:22"
user="root"
passwd="root"

[local]
listen="127.0.0.1:1080"
```

## Run

```
./ssh-tunnel-socks5 -c config.toml
```

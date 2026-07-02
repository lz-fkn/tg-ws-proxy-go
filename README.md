# TG WS Proxy Go для НЕвстроенных устройств

### Предисловие

изначально я пользовался оригинальным прокси на питоне, но т.к. прокси крутился на бедном одноплатнике с 1гб памяти, тратить 200+ мб на один прокси для тг было слишком расточительно. мне приглянулся его форк, переписанный на гоу, но сделанный для всяких встроенных устройств по типу роутеров. мне лично такое нахуй не упало, поэтому немного подобрав этот форк под себя (заменил логгинг на свой простой, пусть systemd с ними ебётся), получилась вот такая поебота. нихуя в целом не поменялось, но из-за того что моды делались для околодесктопа логи могут быть слишком жирными, да и если вид репы вас не устраивает то мне похуй, делал для себя

### Сборка

надеюсь го и гит у вас есть, если нет то сами найдёте в инете, не маленькие
```bash
go version # нужна хотя бы >1.22

git clone -b main-go https://github.com/lz-fkn/tg-ws-proxy-go
go mod tidy
go build -ldflags="-s -w" 
```

### Использование
вместо всяких конфиг файлов и прочего говна теперь просто аргументы (они и так были, но...)

```
Usage of tg-ws-proxy:
  -buf-kb int
        Socket buffer size in KB (default 256)
  -cfproxy-domain string
        Cloudflare-proxied domain for WS fallback (default "pclead.co.uk")
  -cfproxy-domains string
        Comma-separated Cloudflare proxy domain pool for WS fallback
  -cfproxy-domains-url string
        URL to fetch CF proxy domain list from (default "https://raw.githubusercontent.com/Flowseal/tg-ws-proxy/main/.github/cfproxy-domains.txt")
  -cfproxy-priority
        Try cfproxy before TCP fallback (default true)
  -cfproxy-worker-domain string
        Comma-separated Cloudflare Worker domain(s) for WS fallback (e.g. name-1234.user.workers.dev); tried first when set
  -credits
        Show credits and exit
  -dc-ip value
        Target DC IP as DC:IP; repeatable
  -dc-ip-default string
        Default WS target IP for all implicit DCs when --dc-ip is not provided (default "149.154.167.220")
  -dc-ip-default-pool string
        Default WS target IP pool for implicit DCs, comma-separated
  -dc-ip-pool value
        Target pool as DC:IP1,IP2,...; repeatable
  -fake-tls-domain string
        Enable Fake TLS (ee-secret) with masking domain (https://github.com/Flowseal/tg-ws-proxy/blob/main/docs/FakeTlsNginx.md)
  -gen-secret
        Generate random secret and print it
  -host string
        Listen host (default "127.0.0.1")
  -loglevel int
        Log Level (0=FATAL, 1=ERROR, 2=WARN, 3=INFO, 4=VERBOSE) (default 3)
  -max-conns int
        Max concurrent client sessions (default 1024)
  -no-cfproxy
        Disable Cloudflare proxy fallback
  -no-cfproxy-domain-refresh
        Disable periodic CF proxy domain refresh from URL
  -pool-size int
        WS pool size per DC (default 4)
  -port int
        Listen port (default 1443)
  -print-link
        Print the tg:// connect link and exit
  -secret string
        MTProto secret (32 hex chars)
  -timestamps
        Show timestamps in logs
```
# chasquid-rspamd
[Chasquid](https://blitiri.com.ar/p/chasquid/) hook for [rspamd](https://rspamd.com/)

## Installation
Build the binary yourself using `go build .` or download one of the archives/packages from [GitHub releases](https://github.com/Thor77/chasquid-rspamd/releases).

## Usage
```
Usage of chasquid-rspamd:
  -url string
    	rspamd control url (default "http://127.0.0.1:11333")
```

Add the hook to `/etc/chasquid/hooks/post-data` and adjust the path if the binary was not installed into `$PATH`.
```bash
# rspamd
if command -v chasquid-rspamd >/dev/null; then
	chasquid-rspamd < "$TF" 2>/dev/null
fi
```

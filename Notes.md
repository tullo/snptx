# Notes

go run

* `go run ./cmd/web`
* `go run ./cmd/web -addr :80`
* `go run ./cmd/web -addr=:80`
* `export SNPTX_ADDR=":9999"`
* `go run ./cmd/web -addr=$SNPTX_ADDR`
* `go run github.com/tullo/snptx`

go build

* `go build ./cmd/web`
* `./web`
* `go build -o app ./cmd/web`
* `./app`

Decoupled Logging

```bash
# redirect the stdout and stderr streams to on-disk files
go run ./cmd/web >>/tmp/info.log 2>>/tmp/error.log
```

http.FileServer

* range requests are fully supported
* application is serving large files
* and you want to support resumable downloads
* example: to request bytes 100-199 of the logo.png
* `curl -i -H "Range: bytes=100-199" --output - http://localhost:4200/static/img/logo.png`

```console
HTTP/1.1 206 Partial Content
Accept-Ranges: bytes
Content-Length: 100
Content-Range: bytes 100-199/1075
Content-Type: image/png
Last-Modified: Thu, 04 May 2017 13:07:52 GMT
Date: Tue, 10 Mar 2020 11:28:59 GMT
```

## Disabling Directory Listings

Add a blank index.html file to the specific directories:

`find ./ui/static -type d -exec touch {}/index.html \;`

## Installing mysql database driver

``` bash
# download the latest version of the package
go get -u github.com/go-sql-driver/mysql
go: downloading github.com/go-sql-driver/mysql v1.5.0
go: github.com/go-sql-driver/mysql upgrade => v1.5.0

```

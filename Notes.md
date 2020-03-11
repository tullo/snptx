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

## Installing dependencies

``` bash
# download the latest available version under major release number 1
# go get -u github.com/go-sql-driver/mysql@v1

# download the latest version of the mysql driver package
go get -u github.com/go-sql-driver/mysql
go: downloading github.com/go-sql-driver/mysql v1.5.0
go: github.com/go-sql-driver/mysql upgrade => v1.5.0

# composable middleware chains
go get github.com/justinas/alice
# pattern based mux
go get github.com/bmizerany/pat
# cookie-based session store -> limited (to 4KB)
# uses encrypted and authenticated cookies
go get github.com/golangcollege/sessions

go mod tidy     # removes unused packages from go.mod and go.sum
go mod verify   # verify checksums of the downloaded packages
go mod download # errors if mismatch between dependencies and checksums
go get -u ...   # upgrade to latest available minor or patch release of a package
go get -u github.com/foo/bar@v2.0.0
```

## DB preparation

```bash
docker-compose up -d

docker-compose ps
#    Name                Command             State                 Ports
# ------------------------------------------------------------------------------------
# snptx_db_1   docker-entrypoint.sh mysqld   Up      0.0.0.0:3306->3306/tcp, 33060/tcp

docker-compose exec db mysql -u root -p
```

```sql
CREATE DATABASE snptx CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'web'@'localhost';
GRANT SELECT, INSERT, UPDATE ON snptx.* TO 'web'@'localhost';
ALTER USER 'web'@'localhost' IDENTIFIED BY 'snptx';

CREATE USER 'web'@'172.19.0.1';
GRANT SELECT, INSERT, UPDATE ON snptx.* TO 'web'@'172.19.0.1';
ALTER USER 'web'@'172.19.0.1' IDENTIFIED BY 'snptx';

USE snptx;

CREATE TABLE snippets (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    created DATETIME NOT NULL,
    expires DATETIME NOT NULL
);

CREATE INDEX idx_snippets_created ON snippets(created);

INSERT INTO snippets (title, content, created, expires) VALUES (
    'An old silent pond',
    'An old silent pond...\nA frog jumps into the pond,\nsplash! Silence again.\n\n– Matsuo Bashō',
    UTC_TIMESTAMP(),
    DATE_ADD(UTC_TIMESTAMP(), INTERVAL 365 DAY)
);

INSERT INTO snippets (title, content, created, expires) VALUES (
    'Over the wintry forest',
    'Over the wintry\nforest, winds howl in rage\nwith no leaves to blow.\n\n– Natsume Soseki',
    UTC_TIMESTAMP(),
    DATE_ADD(UTC_TIMESTAMP(), INTERVAL 365 DAY)
);

INSERT INTO snippets (title, content, created, expires) VALUES (
    'First autumn morning',
    'First autumn morning\nthe mirror I stare into\nshows my father''s face.\n\n– Murakami Kijo',
    UTC_TIMESTAMP(),
    DATE_ADD(UTC_TIMESTAMP(), INTERVAL 7 DAY)
);

CREATE TABLE users (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    hashed_password CHAR(60) NOT NULL,
    created DATETIME NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email);

```

```bash
docker-compose exec db mysql -D snptx -u web -p
```

```sql
SELECT id, title, expires FROM snippets;
+----+------------------------+---------------------+
| id | title                  | expires             |
+----+------------------------+---------------------+
|  1 | An old silent pond     | 2021-03-10 14:47:14 |
|  2 | Over the wintry forest | 2021-03-10 14:47:29 |
|  3 | First autumn morning   | 2020-03-17 14:47:43 |
+----+------------------------+---------------------+
```

### [XSS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection) (Cross Site Scripting)

Hint for older browser implementations

```bash
# 1; mode=block - the browser will prevent rendering of the page if an attack is detected

# HEAD request to get http headers only
$ curl -I http://localhost:4200/
HTTP/1.1 200 OK
X-Frame-Options: deny
X-Xss-Protection: 1; mode=block
...
```

Modern browser: Use a strong [Content-Security-Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy) that disables the use of inline JavaScript ('unsafe-inline')

### Panic Recovery

Closes client connection automatically

```bash
$ curl -I http://localhost:4200/
HTTP/1.1 500 Internal Server Error
Connection: close
...
```

### Generate selv signed TLS certificate

```bash
echo 'tls/' >> .gitignore
mkdir tls && cd tls
minica --domains localhost
```

### Security/Server Side TLS

[Recommended configurations](https://wiki.mozilla.org/Security/Server_Side_TLS) for modern clients that support TLS 1.3, with no need for backwards compatibility.

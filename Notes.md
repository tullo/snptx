# Notes

go run

* `go run ./cmd/snptx`
* `go run ./cmd/snptx -addr :80`
* `go run ./cmd/snptx -addr=:80`
* `export SNPTX_ADDR=":9999"`
* `go run ./cmd/snptx -addr=$SNPTX_ADDR`
* `go run github.com/tullo/snptx`

go build

directories named `testdata` `_` `.`  will be ignored when compiling Go binaries

* `go build ./cmd/snptx`
* `./web`
* `go build -o app ./cmd/snptx`
* `./app`

go test

* `go test ./...`
* `go test -v ./cmd/snptx`
* `go test -v github.com/tullo/snptx/cmd/snptx`
* `go test -failfast -v ./cmd/snptx`
* `go test ./cmd/snptx -run TestSecureHeaders`
* `go test ./cmd/snptx -run="^TestHumanDate$/^UTC|CET$"`
* `go test github.com/tullo/snptx/cmd/snptx -run TestSecureHeaders`
* skip long running tests if the `-short` flag is provided
  * `if testing.Short() { t.Skip("xyz") }`
  * `go test -short ./pkg/models`
  * might consider to only run integration tests before committing a change
* enabling the race detector
  * flags data races at runtime - no static analysis
  * increases overall running time of tests
  * `go test -race ./cmd/snptx`
* parallel tests marked using `t.Parallel()`
  * uses all available processors per default
  * `go test -parallel 4 ./cmd/snptx`
* antidote for cached test results
  * go test -run="TestSignupUser" -count=1 ./cmd/snptx
  * go clean -testcache

## Debug Mode

Activate debug mode to get detailed errors and stack traces in the http response.

```bash
go run ./cmd/snptx -debug
```

## Decoupled Logging

```bash
# redirect the stdout and stderr streams to on-disk files
go run ./cmd/snptx >>/tmp/info.log 2>>/tmp/error.log
```

## Serving static content

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

# csrf protection middleware
go get github.com/justinas/nosurf

go mod tidy     # removes unused packages from go.mod and go.sum
go mod verify   # verify checksums of the downloaded packages
go mod download # errors if mismatch between dependencies and checksums
go list -mod=readonly -m all # to view final versions that will be used in a build for all direct and indirect dependencies
go mod why -m google.golang.org/appengine
go list -u -m all # to view available minor and patch upgrades for all direct and indirect dependencies
go get -u ...   # to upgrade to latest available minor or patch release of a package
go get -u github.com/foo/bar@v2.0.0
go get foo@v1.6.2, go get foo@e3702bed2, go get foo@'<v1.6.2' # @version suffix or "module query"
go get -u=patch ./... # to use the latest patch releases (-t to also upgrade test dependencies)
go list -mod=mod -m all
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

## [XSS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection) (Cross Site Scripting)

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

## Panic Recovery

Closes client connection automatically

```bash
$ curl -I http://localhost:4200/
HTTP/1.1 500 Internal Server Error
Connection: close
...
```

## Generate selv signed TLS certificate

```bash
echo 'tls/' >> .gitignore
mkdir tls && cd tls
minica --domains localhost
```

## Security/Server Side TLS

[Recommended configurations](https://wiki.mozilla.org/Security/Server_Side_TLS) for modern clients that support TLS 1.3, with no need for backwards compatibility.

## Require authentication for specific routes

```bash
# === [GET /snippet/create] ===================================================
$ curl --include --insecure https://localhost:4200/snippet/create
HTTP/2 303
content-type: text/html; charset=utf-8
location: /user/login
...
<a href="/user/login">See Other</a>.
# === [POST /snippet/create] ==================================================
$ curl --include --insecure -X POST https://localhost:4200/snippet/create
HTTP/2 303
location: /user/login
...
# === [POST /user/logout] =====================================================
$ curl --include --insecure -X POST https://localhost:4200/user/logout
HTTP/2 303
location: /user/login
...
```

## SameSite Cookies

To prevent CSRF attacks set the SameSite attribute on our session cookie.

* Works with 84% of the browsers out there https://caniuse.com/#feat=same-site-cookie-attribute
* [Cross-Site Request Forgery (CSRF) Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)

## Integration Test

Database preparation:

```sql
CREATE DATABASE test_snptx CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER 'test_web'@'localhost';
GRANT CREATE, DROP, ALTER, INDEX, SELECT, INSERT, UPDATE, DELETE ON test_snptx.* TO 'test_web'@'localhost';
ALTER USER 'test_web'@'localhost' IDENTIFIED BY 'pass';

CREATE USER 'test_web'@'172.21.0.1';
GRANT CREATE, DROP, ALTER, INDEX, SELECT, INSERT, UPDATE, DELETE ON test_snptx.* TO 'test_web'@'172.21.0.1';
ALTER USER 'test_web'@'172.21.0.1' IDENTIFIED BY 'pass';

-- pkg/models/mysql/testdata/setup.sql
-- pkg/models/mysql/testdata/teardown.sql
```

```bash
$ go test -v ./pkg/models/mysql
=== RUN   TestUserModelGet
=== RUN   TestUserModelGet/Valid_ID
=== RUN   TestUserModelGet/Zero_ID
=== RUN   TestUserModelGet/Non-existent_ID
--- PASS: TestUserModelGet (1.35s)
    --- PASS: TestUserModelGet/Valid_ID (0.47s)
    --- PASS: TestUserModelGet/Zero_ID (0.44s)
    --- PASS: TestUserModelGet/Non-existent_ID (0.45s)
PASS
ok      github.com/tullo/snptx/pkg/models/mysql 1.356s


$ go test -v -short ./pkg/models/mysql
=== RUN   TestUserModelGet
    TestUserModelGet: users_test.go:14: mysql: skipping integration test
--- SKIP: TestUserModelGet (0.00s)
PASS
ok      github.com/tullo/snptx/pkg/models/mysql (cached)
```

## Profiling Test Coverage

Get a test coverage summary

```bash
go test -cover ./...
```

Get a detailed breakdown of test coverage by method and function

```bash
$ go test -coverprofile=/tmp/profile.out ./...

ok  github.com/tullo/snptx/cmd/snptx  0.020s  coverage: 51.0% of statements
?   github.com/tullo/snptx/pkg/forms    [no test files]
?   github.com/tullo/snptx/pkg/models   [no test files]
?   github.com/tullo/snptx/pkg/models/mock  [no test files]
ok  github.com/tullo/snptx/pkg/models/mysql 1.530s  coverage: 10.6% of statements


$ go tool cover -func=/tmp/profile.out

github.com/tullo/snptx/cmd/snptx/handlers.go:13:          home                    0.0%
github.com/tullo/snptx/cmd/snptx/handlers.go:25:          showSnippet             100.0%
github.com/tullo/snptx/cmd/snptx/handlers.go:50:          createSnippetForm       0.0%
github.com/tullo/snptx/cmd/snptx/handlers.go:57:          createSnippet           0.0%
github.com/tullo/snptx/cmd/snptx/handlers.go:93:          signupUserForm          100.0%
github.com/tullo/snptx/cmd/snptx/handlers.go:99:          signupUser              86.4%
github.com/tullo/snptx/cmd/snptx/handlers.go:134:         loginUserForm           0.0%
github.com/tullo/snptx/cmd/snptx/handlers.go:140:         loginUser               0.0%
github.com/tullo/snptx/cmd/snptx/handlers.go:168:         logoutUser              0.0%
github.com/tullo/snptx/cmd/snptx/helpers.go:13:           ping                    100.0%
github.com/tullo/snptx/cmd/snptx/helpers.go:17:           serverError             100.0%
github.com/tullo/snptx/cmd/snptx/helpers.go:24:           clientError             100.0%
github.com/tullo/snptx/cmd/snptx/helpers.go:28:           notFound                100.0%
github.com/tullo/snptx/cmd/snptx/helpers.go:32:           addDefaultData          85.7%
github.com/tullo/snptx/cmd/snptx/helpers.go:52:           isAuthenticated         75.0%
github.com/tullo/snptx/cmd/snptx/helpers.go:61:           render                  60.0%
github.com/tullo/snptx/cmd/snptx/main.go:41:              main                    0.0%
github.com/tullo/snptx/cmd/snptx/main.go:104:             openDB                  0.0%
github.com/tullo/snptx/cmd/snptx/middleware.go:13:        secureHeaders           100.0%
github.com/tullo/snptx/cmd/snptx/middleware.go:23:        noSurf                  100.0%
github.com/tullo/snptx/cmd/snptx/middleware.go:34:        logRequest              100.0%
github.com/tullo/snptx/cmd/snptx/middleware.go:43:        recoverPanic            66.7%
github.com/tullo/snptx/cmd/snptx/middleware.go:62:        requireAuthentication   16.7%
github.com/tullo/snptx/cmd/snptx/middleware.go:77:        authenticate            33.3%
github.com/tullo/snptx/cmd/snptx/routes.go:10:            routes                  100.0%
github.com/tullo/snptx/cmd/snptx/templates.go:22:         humanDate               100.0%
github.com/tullo/snptx/cmd/snptx/templates.go:35:         newTemplateCache        76.5%
github.com/tullo/snptx/pkg/models/mysql/snippets.go:16: Insert                  0.0%
github.com/tullo/snptx/pkg/models/mysql/snippets.go:37: Get                     0.0%
github.com/tullo/snptx/pkg/models/mysql/snippets.go:57: Latest                  0.0%
github.com/tullo/snptx/pkg/models/mysql/users.go:19:    Insert                  0.0%
github.com/tullo/snptx/pkg/models/mysql/users.go:43:    Authenticate            0.0%
github.com/tullo/snptx/pkg/models/mysql/users.go:75:    Get                     87.5%
total:                          (statements)    41.1%


# A more visual way to view the coverage profile is to use the -html flag
$ go tool cover -html=/tmp/profile.out


# using `-covermode=count` or `-covermode=atomic` makes the coverage profile record the exact
# number of times that each statement is executed during the tests.
$ go test -covermode=count -coverprofile=/tmp/profile.out ./...
# statements which are executed more frequently are then shown in a more saturated shade of green
$ go tool cover -html=/tmp/profile.out
```

## Git

```bash
git lg
git log -u -1 43029e0
```

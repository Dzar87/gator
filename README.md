# Gator

Rudimental CLI RSS-Feed aggregator

## Install

```sh
$ go install github.com/Dzar87/gator
```

## Usage

Create `~/.gatorconfig`
```json
{
    "db_url": "postgres://<DB_NAME>:<DB_PASS>@localhost:<DB_PORT>/gator?sslmode=disable",
}
```

```
Commands:

gator register <username>  # Register a user, implicitly logs in
gator login <username>     # Changes user
gator reset                # Clears the database
gator users                # List users
gator agg                  # Start RSS-Feed aggregation
gator addfeed <name> <url> # Add a RSS-Feed
gator feeds                # List added RSS-feeds
gator follow <url>         # Follow RSS-feed
gator following            # List followed RSS-feeds
gator unfollow <url>       # Unfollow RSS-feed
gator browse <limit>       # Browse RSS-feed posts
```

## Development

### Dependencies

Tools:
*   Go
    version: 1.26.1
*   Docker (or Podman)
*   [Goose](https://pkg.go.dev/github.com/pressly/goose#section-documentation)
    ```
    go install github.com/pressly/goose/v3/cmd/goose@latest
    ```
*   [sqlc](https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html)
    ```
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    ```

Packages:
*   github.com/google/uuid
    version: v1.6.0 
*   github.com/lib/pq
    version: v1.12.3

## Makefile

```sh
Makefile targets for:
File: .../gator/Makefile

help                           Show this help message
test                           Run go test
test-cov                       Run go test with coverage
vet                            Run go vet
fmt                            Run go fmt
build                          Build the project
up                             Run compose up in detached mode
down                           Run compose down
psql                           Run psql on the postgres container
psql-notty                     Run psql on the postgres container (no-tty)
psql-url                       Print the Postgres connection string
migrate-up                     Run goose up
migrate-down                   Run goose down
```

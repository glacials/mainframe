# mainframe

Mainframe is a monolith containing some personal automations and convenience
code that I run 24/7 and pipe data into and out of. It's a replacement for a
series of brittle shell scripts and cron jobs.

If you can find some use out of it, great!

## Running

First copy `.envrc.example` to `.envrc` and fill it out. Source it or use [direnv](https://github.com/direnv/direnv) to source it automatically.

```sh
go build .
./mainframe
```

## Running as a daemon

To run mainframe how it's meant to be run, e.g. on an old machine or Raspberry
Pi in a closet, first install dependencies:

```sh
sudo apt-get install moreutils
```

Then set up a user cron like so:

```cron
0 0 * * * chronic bash /path/to/mainframe/supervisor.sh
```

The supervisor script will handle the rest, including auto-updating.

## Development

### Migrations

Prerequisites: Install
[`golang-migrate`](https://github.com/golang-migrate/migrate) with the
[`sqlite`](https://modernc.org/sqlite) driver.

```sh
go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

[`sqlite3`][https://github.com/mattn/go-sqlite3] would probably work too, but is
less friendly to cross-compiling.

**WARNING:** Do not install `golang-migrate` from Homebrew, as that version does
not include any SQLite drivers.

#### Creating a new migration

```sh

migrate -path db/migrations create -seq -ext sql name_of_migration
$EDITOR db/migrations/*name_of_migration.{up,down}.sql
```

#### Running migrations

```sh
migrate -path db/migrations -database sqlite://mainframe.db up
```

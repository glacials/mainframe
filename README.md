# mainframe

Mainframe is a monolith containing some personal automations and convenience
code that I run 24/7 and pipe data into and out of. It's a replacement for a
series of brittle shell scripts and cron jobs.

If you can find some use out of it, great!

## Running for development

### Prerequisites

- [Go](https://golang.org/) 1.19+
- [`just`](https://github.com/casey/just) (`brew install just`)
- [`gow`](https://github.com/mitranim/gow) (`go install github.com/mitranim/gow`)

### Developing

First copy `.envrc.example` to `.envrc` and fill it out. Source it or use
[direnv](https://github.com/direnv/direnv) to source it automatically.

Then to develop, run:

```sh
just serve
```

This will contnually rebuild and then reboot `mainframe` when a source file
changes.

## Running as a daemon

To run mainframe how it's meant to be run in production, i.e. on an old machine
or Raspberry Pi in a closet, first install dependencies:

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

[`sqlite3`](https://github.com/mattn/go-sqlite3) would probably work too, but is
less friendly to cross-compiling.

**Warning:** Do not install `golang-migrate` from Homebrew, as that version does
not include any SQLite drivers.

#### Running migrations

```sh
migrate -path db/migrations -database sqlite://mainframe.db up
```

#### Creating a new migration

```sh
just create-migration NAME
$EDITOR db/migrations/*name_of_migration.{up,down}.sql
```

#### Forcing a migration version

If a migration errors, you may have to force the migration engine back to the previous
version, possibly undoing any partial steps manually. Forcing a migration version does
not run any migration files.

```sh
just force-migration VERSION
```

set positional-arguments

default:
  @just --list

run:
  gow -e go,html,tmpl,css run .

migrate:
  migrate -path db/migrations -database sqlite://mainframe.db up

create-migration name:
  migrate create -dir db/migrations -seq -ext sql {{name}}

force-migration version:
  migrate -path db/migrations -database sqlite://mainframe.db force {{version}}

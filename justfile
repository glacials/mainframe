set positional-arguments

serve:
  gow -e go,html,tmpl,css run .

migrate:
  migrate -path db/migrations -database sqlite://mainframe.db up

create-migration:
  migrate create -dir db/migrations -seq -ext sql $0

force-migration:
  migrate -path db/migrations -database sqlite://mainframe.db force $0

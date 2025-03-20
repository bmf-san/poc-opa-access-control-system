module github.com/bmf-san/poc-opa-access-control-system/cmd/employee

go 1.24

require (
	github.com/bmf-san/poc-opa-access-control-system/internal v0.0.0
	github.com/jackc/pgx/v5 v5.7.2
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace github.com/bmf-san/poc-opa-access-control-system/internal => ../../internal

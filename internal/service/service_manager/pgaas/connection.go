package pgaas

import (
	"strings"

	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/domain"
)

const (
	maskPlaceholder = "MASKEDPWD"
	maskedPassword  = "••••"
)

func buildPgConnection(host string, port uint64, user, password, dbName string) domain.PgConnectionView {
	conn := resources.Postgres{
		Host:    host,
		Port:    port,
		User:    user,
		Pwd:     password,
		DbName:  dbName,
		SslMode: "disable",
	}

	masked := conn

	masked.Pwd = maskPlaceholder

	return domain.PgConnectionView{
		Dsn:       conn.ConnectionString(),
		MaskedDsn: strings.ReplaceAll(masked.ConnectionString(), maskPlaceholder, maskedPassword),
	}
}

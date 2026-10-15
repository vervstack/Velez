package pgaas

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
)

func Test_BuildPgConnection_MasksPassword(t *testing.T) {
	password := "p@ss/w:rd"

	conn := buildPgConnection("pg-host", 5432, "user", password, "db")

	require.Contains(t, conn.Dsn, password)
	require.NotContains(t, conn.MaskedDsn, password)
	require.Contains(t, conn.MaskedDsn, "••••")
	require.NotContains(t, conn.MaskedDsn, maskPlaceholder)
	require.Contains(t, conn.MaskedDsn, "pg-host:5432/db")
}

func Test_ExposedPort_Selection(t *testing.T) {
	published := publishedContainer{ports: []container.Port{
		{PrivatePort: 5432, PublicPort: 0, Type: "tcp"},
		{PrivatePort: 5432, PublicPort: 15432, Type: "udp"},
		{PrivatePort: 5432, PublicPort: 15433, Type: "tcp"},
	}}

	got := published.exposedPort(5432)

	require.NotNil(t, got)
	require.Equal(t, uint32(15433), *got)
	require.Nil(t, published.exposedPort(6000))
	require.Nil(t, publishedContainer{}.exposedPort(5432))
}

package cluster_steps

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

type createPgUserStep struct {
	dsn *string

	schema   string
	nodeName string
	pwd      string
}

func CreatePgUserForNode(rootDsn *string,
	schema, nodeName, pwd string,
) steps.Step {
	return &createPgUserStep{
		rootDsn,
		schema,
		nodeName,
		pwd,
	}
}

func (c *createPgUserStep) Do(ctx context.Context) error {
	conn, err := sql.Open("postgres", *c.dsn)
	if err != nil {
		return rerrors.Wrap(err, "error opening connection to database")
	}
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Err(closeErr).
				Msg("error closing database connection for root dsn when creating new user")
		}
	}()

	_, err = conn.Exec(
		fmt.Sprintf(`
		CREATE USER %[1]s WITH PASSWORD '%[2]s';
		
		GRANT working_node TO %[1]s;
`, c.nodeName, c.pwd, c.nodeName))
	if err != nil {
		return rerrors.Wrap(err, "error creating database user")
	}

	log.Info().
		Str("pwd", c.pwd).
		Msg("user with created")

	return nil
}

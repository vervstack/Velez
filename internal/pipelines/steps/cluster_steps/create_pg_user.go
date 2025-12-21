package cluster_steps

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"text/template"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

var (
	//go:embed templates/pg_create_user.pattern
	createUserPattern string
	createUserTmplt   = template.New("create-pg-user")
)

func init() {
	var err error
	createUserTmplt, err = createUserTmplt.Parse(createUserPattern)
	if err != nil {
		panic(rerrors.Wrap(err, "error parsing create user template"))
	}
}

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

	buf := bytes.Buffer{}

	err = createUserTmplt.Execute(&buf,
		map[string]interface{}{
			"username": c.nodeName,
			"password": c.pwd,
			"schema":   c.schema,
		})
	if err != nil {
		return rerrors.Wrap(err, "error compiling create user template")
	}

	_, err = conn.Exec(buf.String())
	if err != nil {
		return rerrors.Wrap(err, "error creating database user")
	}

	log.Info().
		Str("pwd", c.pwd).
		Msg("user with created")

	return nil
}

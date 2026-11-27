package jobs

import (
	"context"
	"encoding/json"

	"github.com/containerd/errdefs"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	pgFailoverCredentialsPath = pgVolumePath + "/.velez-failover-credentials"
	pgFailoverCredentialsMode = 0o600
)

type pgFailoverLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DbName   string `json:"db_name"`
}

// ReadPgFailoverCredentials reads the copied cluster's login from its own
// managed volume. The original primary can be dropped after failover.
func ReadPgFailoverCredentials(
	ctx context.Context, docker node_clients.Docker, name string,
) (domain.PgInstanceCredentials, error) {
	id, err := FindPgInstanceContainerId(ctx, docker, name)
	if rerrors.Is(err, user_errors.ErrContainerNotFoundInEnvironment) {
		return domain.PgInstanceCredentials{}, nil
	}

	if err != nil {
		return domain.PgInstanceCredentials{}, rerrors.Wrap(err)
	}

	return readPgFailoverCredentialsById(ctx, docker, id)
}

func readPgFailoverCredentialsById(
	ctx context.Context, docker node_clients.Docker, id string,
) (domain.PgInstanceCredentials, error) {
	content, err := readFileFromContainer(ctx, docker.Client(), id, pgFailoverCredentialsPath)
	if errdefs.IsNotFound(err) {
		return domain.PgInstanceCredentials{}, nil
	}

	if err != nil {
		return domain.PgInstanceCredentials{}, rerrors.Wrap(err, "error reading promoted PostgreSQL credentials")
	}

	var credentials pgFailoverLogin

	err = json.Unmarshal(content, &credentials)
	if err != nil {
		return domain.PgInstanceCredentials{}, rerrors.Wrap(err, "error decoding promoted PostgreSQL credentials")
	}

	login := domain.PgInstanceCredentials{
		Username: credentials.Username, Password: credentials.Password, DbName: credentials.DbName,
	}

	return login, nil
}

func (j *pgMasterChange) savePrimaryCredentials(ctx context.Context, endpoint replicationEndpoint) error {
	credentials := pgFailoverLogin{
		Username: endpoint.user, Password: endpoint.password, DbName: endpoint.dbName,
	}

	content, err := json.Marshal(credentials)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = writeFileToContainer(ctx, j.handler.docker.Client(), j.payload.GetNewMasterContainerId(),
		pgFailoverCredentialsPath, content, pgFailoverCredentialsMode)
	if err != nil {
		return rerrors.Wrap(err, "error preserving promoted PostgreSQL credentials")
	}

	return nil
}

package pgaas

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *PgaasService) GetPgInstanceCredentials(
	ctx context.Context, name string,
) (domain.PgInstanceCredentials, error) {
	svc, err := s.dataStorage.Services().GetByName(ctx, name)
	if err != nil {
		return domain.PgInstanceCredentials{}, rerrors.Wrap(err, "error getting pg instance service")
	}

	instance, err := s.dataStorage.PgInstances().GetPgInstanceByServiceID(ctx, svc.ID)
	if err != nil {
		return domain.PgInstanceCredentials{}, rerrors.Wrap(err, "error getting pg instance row")
	}

	secretRef, err := domain.ParseSecretRef(instance.SecretRef)
	if err != nil {
		return domain.PgInstanceCredentials{}, rerrors.Wrap(err, "error parsing pg instance secret ref")
	}

	password, err := s.secrets.Get(ctx, secretRef)
	if err != nil {
		return domain.PgInstanceCredentials{}, rerrors.Wrap(err, "error resolving pg instance password")
	}

	// Host is the instance's own service/container name - Velez's Docker
	// network resolves container names as hostnames, the same convention
	// internal/jobs/enable_statefull.go's getRootDsnJob uses (Host: j.pgName).
	conn := resources.Postgres{
		Host:    name,
		Port:    uint64(instance.Port),
		User:    instance.Username,
		Pwd:     password,
		DbName:  instance.DbName,
		SslMode: "disable",
	}

	credentials := domain.PgInstanceCredentials{
		DbName:   instance.DbName,
		Username: instance.Username,
		Password: password,
		Dsn:      conn.ConnectionString(),
	}

	return credentials, nil
}

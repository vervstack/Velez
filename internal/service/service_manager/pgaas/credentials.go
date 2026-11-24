package pgaas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (s *PostgresServiceImpl) GetPgInstanceCredentials(
	ctx context.Context, name, requestHost string,
) (domain.PgInstanceCredentials, error) {
	instance, err := s.getPgInstance(ctx, name)
	if err != nil {
		return domain.PgInstanceCredentials{}, rerrors.Wrap(err)
	}

	instance, password, err := s.effectivePgLogin(ctx, name, instance)
	if err != nil {
		return domain.PgInstanceCredentials{}, rerrors.Wrap(err)
	}

	// Host is the instance's own service/container name - Velez's Docker
	// network resolves container names as hostnames, the same convention
	// internal/jobs/enable_statefull.go's getRootDsnJob uses (Host: j.pgName).
	inNetwork := buildPgConnection(name, uint64(instance.Port), instance.Username, password, instance.DbName)

	credentials := domain.PgInstanceCredentials{
		DbName:    instance.DbName,
		Username:  instance.Username,
		Password:  password,
		Dsn:       inNetwork.Dsn,
		InNetwork: inNetwork,
	}

	external, hasExternal, err := s.externalConnection(ctx, name, requestHost, instance, password)
	if err != nil {
		return domain.PgInstanceCredentials{}, rerrors.Wrap(err)
	}

	if hasExternal {
		credentials.External = &external
	}

	return credentials, nil
}

func (s *PostgresServiceImpl) effectivePgLogin(
	ctx context.Context, name string, instance domain.PgInstance,
) (domain.PgInstance, string, error) {
	source, err := jobs.ReadPgReplicaSource(ctx, s.docker, name)
	if err != nil {
		return domain.PgInstance{}, "", rerrors.Wrap(err, "error reading pg replica source")
	}

	loginInstance := name
	if source != "" {
		loginInstance = source
	}

	promoted, err := jobs.ReadPgFailoverCredentials(ctx, s.docker, loginInstance)
	if err != nil {
		return domain.PgInstance{}, "", rerrors.Wrap(err)
	}

	if promoted.Username != "" {
		instance.Username = promoted.Username
		instance.DbName = promoted.DbName

		return instance, promoted.Password, nil
	}

	if source != "" {
		upstream, upstreamErr := s.getPgInstance(ctx, source)
		if upstreamErr != nil {
			return domain.PgInstance{}, "", rerrors.Wrap(upstreamErr, "error getting source pg instance")
		}

		instance.Username = upstream.Username
		instance.DbName = upstream.DbName
		instance.SecretRef = upstream.SecretRef
	}

	secretRef, err := domain.ParseSecretRef(instance.SecretRef)
	if err != nil {
		return domain.PgInstance{}, "", rerrors.Wrap(err, "error parsing pg instance secret ref")
	}

	password, err := s.secrets.Get(ctx, secretRef)
	if err != nil {
		return domain.PgInstance{}, "", rerrors.Wrap(err, "error resolving pg instance password")
	}

	return instance, password, nil
}

func (s *PostgresServiceImpl) externalConnection(
	ctx context.Context, name, requestHost string, instance domain.PgInstance, password string,
) (domain.PgConnectionView, bool, error) {
	filter := map[string]string{labels.VervServiceLabel: name}

	published, err := s.publishedContainers(ctx, filter)
	if err != nil {
		return domain.PgConnectionView{}, false, rerrors.Wrap(err, "error resolving published pg ports")
	}

	pub, ok := published[name]
	if !ok {
		return domain.PgConnectionView{}, false, nil
	}

	exposedPort := pub.exposedPort(instance.Port)
	if exposedPort == nil {
		return domain.PgConnectionView{}, false, nil
	}

	host := pub.environment.PublishedHost(s.docker.Host())
	if host == "" {
		host = requestHost
	}

	if host == "" {
		return domain.PgConnectionView{}, false, nil
	}

	external := buildPgConnection(host, uint64(*exposedPort), instance.Username, password, instance.DbName)

	return external, true, nil
}

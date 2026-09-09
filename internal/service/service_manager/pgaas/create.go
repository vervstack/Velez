package pgaas

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	pgSecretScope = "pgaas"
	pgSecretKey   = "password"
)

func (s *PgaasService) CreatePgInstance(
	ctx context.Context, req domain.CreatePgInstanceReq,
) (domain.PgInstanceView, error) {
	if req.Isolation == velez_api.PgInstanceIsolation_PG_INSTANCE_ISOLATION_SHARED_POOL {
		return domain.PgInstanceView{}, rerrors.Wrap(user_errors.ErrPgSharedPoolIsolationNotSupported)
	}

	dbName := sanitizeIdentifier(req.Name)

	creds := pgCredentials{
		dbName:   dbName,
		username: dbName + "_user",
		password: string(toolbox.RandomBase64(pgPasswordLength)),
	}

	secretRef := domain.SecretRef{Scope: pgSecretScope, Owner: req.Name, Key: pgSecretKey}

	err := s.secrets.Put(ctx, secretRef, creds.password)
	if err != nil {
		return domain.PgInstanceView{}, rerrors.Wrap(err, "error storing pg instance password")
	}

	descriptor, smerdRequest, err := buildDeployRequest(ctx, s.boxes(), req, creds)
	if err != nil {
		return domain.PgInstanceView{}, rerrors.Wrap(err, "error building pg instance deploy request")
	}

	deployReq := domain.CreateDeployReq{
		ServiceName:    req.Name,
		VervDescriptor: &descriptor,
		LaunchSmerd:    domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	// CreateNewDeploy upserts req.Name before looking it up, so no separate
	// UpsertService call is needed here (see verv_services/deploy.go).
	err = s.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return domain.PgInstanceView{}, rerrors.Wrap(err, "error creating pg instance deploy")
	}

	svc, err := s.dataStorage.Services().GetByName(ctx, req.Name)
	if err != nil {
		return domain.PgInstanceView{}, rerrors.Wrap(err, "error getting pg instance service")
	}

	upsertReq := domain.UpsertPgInstanceReq{
		ServiceID: svc.ID,
		DbName:    creds.dbName,
		Username:  creds.username,
		SecretRef: secretRef.String(),
		Port:      pgDefaultPort,
	}

	instance, err := s.dataStorage.PgInstances().UpsertPgInstance(ctx, upsertReq)
	if err != nil {
		return domain.PgInstanceView{}, rerrors.Wrap(err, "error upserting pg instance row")
	}

	if req.OwnerService != "" {
		err = s.dataStorage.ServiceResources().UpsertResource(ctx, req.OwnerService, req.Name, pgResourceType)
		if err != nil {
			return domain.PgInstanceView{}, rerrors.Wrap(err, "error binding pg instance to owner service")
		}
	}

	view := domain.PgInstanceView{
		Name:         req.Name,
		DbName:       instance.DbName,
		Username:     instance.Username,
		Port:         instance.Port,
		Environment:  req.Environment,
		OwnerService: req.OwnerService,
		CreatedAt:    instance.CreatedAt,
		UpdatedAt:    instance.UpdatedAt,
	}

	return view, nil
}

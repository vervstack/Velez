package pgaas

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (s *PostgresServiceImpl) ReplicatePgDatabase(
	ctx context.Context, req domain.ReplicatePgDatabaseReq,
) (string, error) {
	if req.SourceInstanceId != "" {
		credentials, err := s.GetPgInstanceCredentials(ctx, req.SourceInstanceId, "")
		if err != nil {
			return "", rerrors.Wrap(err, "error resolving source pg instance")
		}

		endpoint, err := replicationEndpointFromCredentials(credentials)
		if err != nil {
			return "", rerrors.Wrap(err, "error parsing source pg instance connection")
		}

		req.SourceHost = endpoint.host
		req.SourcePort = endpoint.port
		req.SourceUser = endpoint.user
		req.SourcePassword = endpoint.password
		req.SourceDbName = endpoint.dbName
	}

	if req.TargetInstanceId != "" {
		credentials, err := s.GetPgInstanceCredentials(ctx, req.TargetInstanceId, "")
		if err != nil {
			return "", rerrors.Wrap(err, "error resolving target pg instance")
		}

		endpoint, err := replicationEndpointFromCredentials(credentials)
		if err != nil {
			return "", rerrors.Wrap(err, "error parsing target pg instance connection")
		}

		req.TargetHost = endpoint.host
		req.TargetPort = endpoint.port
		req.TargetUser = endpoint.user
		req.TargetPassword = endpoint.password
	}

	if req.SourceHost == req.TargetHost && req.SourcePort == req.TargetPort {
		return "", rerrors.New("source and target database must differ")
	}

	payload := &velez_api.ReplicatePgDatabaseTaskPayload{
		SourceInstanceId: req.SourceInstanceId,
		SourceHost:       req.SourceHost,
		SourcePort:       req.SourcePort,
		SourceUser:       req.SourceUser,
		SourcePassword:   req.SourcePassword,
		SourceDbName:     req.SourceDbName,
		TargetInstanceId: req.TargetInstanceId,
		TargetHost:       req.TargetHost,
		TargetPort:       req.TargetPort,
		TargetUser:       req.TargetUser,
		TargetPassword:   req.TargetPassword,
	}

	entityId := replicationEntityId(req)

	_, err := s.jobsEngine.EnqueueReplacing(ctx, entityId, jobs.ReplicatePgDatabaseAction, payload)
	if err != nil {
		return "", rerrors.Wrap(err, "error enqueuing replicate pg database task")
	}

	return entityId, nil
}

type replicationServiceEndpoint struct {
	host     string
	port     uint32
	user     string
	password string
	dbName   string
}

func replicationEndpointFromCredentials(credentials domain.PgInstanceCredentials) (replicationServiceEndpoint, error) {
	parsed, err := url.Parse(credentials.InNetwork.Dsn)
	if err != nil {
		return replicationServiceEndpoint{}, rerrors.Wrap(err, "error parsing pg instance dsn")
	}

	portValue, err := strconv.ParseUint(parsed.Port(), 10, 32)
	if err != nil {
		return replicationServiceEndpoint{}, rerrors.Wrap(err, "error parsing pg instance port")
	}

	host, _, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		return replicationServiceEndpoint{}, rerrors.Wrap(err, "error parsing pg instance host")
	}

	password, _ := parsed.User.Password()

	endpoint := replicationServiceEndpoint{
		host:     host,
		port:     uint32(portValue),
		user:     parsed.User.Username(),
		password: password,
		dbName:   parsed.Path[1:],
	}

	return endpoint, nil
}

func replicationEntityId(req domain.ReplicatePgDatabaseReq) string {
	target := req.TargetInstanceId
	if target == "" {
		target = fmt.Sprintf("%s:%d", req.TargetHost, req.TargetPort)
	}

	entityId := "replicate_pg_database:" + target + ":" + req.SourceDbName

	return entityId
}

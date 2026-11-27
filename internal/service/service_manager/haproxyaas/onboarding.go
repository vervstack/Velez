package haproxyaas

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/proto"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	bootstrapWaitTimeout = 2 * time.Minute
)

func (s *HaproxyServiceImpl) BootstrapVcnSidecar(
	ctx context.Context, req *velez_api.BootstrapVcnSidecar_Request,
) (*velez_api.BootstrapVcnSidecar_Response, error) {
	err := validateBootstrapRequest(req)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	name := req.GetName()

	_, err = s.dataStorage.Services().GetByName(ctx, domain.HaproxyServiceName(name))
	if err == nil {
		return nil, rerrors.Wrap(user_errors.ErrHaproxyInstanceExists)
	}

	if !rerrors.Is(err, user_errors.ErrStorageNotFound) {
		return nil, rerrors.Wrap(err, "error checking HAProxy service name")
	}

	payload := &velez_api.BootstrapVcnSidecarTaskPayload{Request: req}

	task, err := s.jobsEngine.EnqueueReplacing(ctx, name, jobs.BootstrapVcnSidecarAction, payload)
	if err != nil {
		return nil, rerrors.Wrap(err, "error enqueuing VCN bootstrap")
	}

	err = checkActiveBootstrapRequest(task, req)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	resp := &velez_api.BootstrapVcnSidecar_Response{EntityId: name, Action: jobs.BootstrapVcnSidecarAction}
	if req.GetIsAsync() {
		return resp, nil
	}

	ctx, cancel := context.WithTimeout(ctx, bootstrapWaitTimeout)
	defer cancel()

	for update := range s.jobsEngine.WatchAfter(ctx, name, jobs.BootstrapVcnSidecarAction, task.ID-1) {
		switch update.Status {
		case tasks_queries.VelezTaskStatusDONE:
			resp.ContainerId, err = jobs.BootstrapSidecarResult(update)
			if err != nil {
				return nil, rerrors.Wrap(err)
			}

			return resp, nil
		case tasks_queries.VelezTaskStatusFAILED:
			return nil, user_errors.New(update.Error.String)
		case tasks_queries.VelezTaskStatusPENDING, tasks_queries.VelezTaskStatusRUNNING:
			continue
		}
	}

	return nil, user_errors.New("VCN sidecar bootstrap ended without a result")
}

func validateBootstrapRequest(req *velez_api.BootstrapVcnSidecar_Request) error {
	err := domain.ValidateInstanceName(req.GetName())
	if err != nil {
		return rerrors.Wrap(err)
	}

	loginUrl := req.GetLoginServerUrl()
	server, err := url.ParseRequestURI(loginUrl)

	if err != nil || server.Hostname() == "" || (server.Scheme != "http" && server.Scheme != "https") ||
		strings.ContainsAny(loginUrl, " \t\r\n\"'\\") {
		return user_errors.New("Enter a valid HTTP or HTTPS login server URL")
	}

	if strings.TrimSpace(req.GetAuthKey()) == "" {
		return user_errors.New("Enter a VCN auth key")
	}

	ports := make(map[uint32]struct{})

	for _, port := range req.GetPorts() {
		if port.GetServicePortNumber() == 0 || port.GetServicePortNumber() > domain.HaproxyMaxPort ||
			port.GetExposedTo() == 0 || port.GetExposedTo() > domain.HaproxyMaxPort ||
			(port.GetProtocol() != velez_api.Port_tcp && port.GetProtocol() != velez_api.Port_udp) {
			return rerrors.Wrap(user_errors.ErrHaproxyInvalidPort)
		}

		if _, isDuplicate := ports[port.GetServicePortNumber()]; isDuplicate {
			return user_errors.New("Duplicate VCN port")
		}

		ports[port.GetServicePortNumber()] = struct{}{}
	}

	return nil
}

func (s *HaproxyServiceImpl) ValidatePgBackend(
	ctx context.Context, req *velez_api.ValidatePgBackend_Request,
) error {
	err := jobs.ValidatePgBackend(ctx, s.dataStorage, s.secrets, s.runtimes, s.docker, req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func checkActiveBootstrapRequest(task tasks_queries.VelezTask, req *velez_api.BootstrapVcnSidecar_Request) error {
	var active velez_api.BootstrapVcnSidecarTaskPayload

	err := json.Unmarshal(task.Context.RawMessage, &active)
	if err != nil {
		return rerrors.Wrap(err, "error decoding active VCN bootstrap")
	}

	activeRequest := active.GetRequest()
	if activeRequest == nil {
		return user_errors.New("The active VCN bootstrap request is unavailable")
	}

	activeRequest.IsAsync = req.GetIsAsync()
	if !proto.Equal(activeRequest, req) {
		return user_errors.New("VCN bootstrap is already running with different settings")
	}

	return nil
}

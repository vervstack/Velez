/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";
import * as VelezApiVelezApi from "./velez_api.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };

type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (keyof T extends infer K
      ? K extends string & keyof T
        ? { [k in K]: T[K] } & Absent<T, K>
        : never
      : never);

export enum DeploymentStatus {
  DEPLOYMENT_STATUS_UNKNOWN = "DEPLOYMENT_STATUS_UNKNOWN",
  SCHEDULED_DEPLOYMENT = "SCHEDULED_DEPLOYMENT",
  SCHEDULED_DELETION = "SCHEDULED_DELETION",
  SCHEDULED_UPGRADE = "SCHEDULED_UPGRADE",
  RUNNING = "RUNNING",
  FAILED = "FAILED",
  DELETED = "DELETED",
  STOPPED = "STOPPED",
}

export type CreateServiceRequest = {
  name?: string;
};

export type CreateServiceResponse = Record<string, never>;

export type CreateService = Record<string, never>;

export type GetServiceRequest = {
  name?: string;
};

type BaseGetServiceResponse = {
};

export type GetServiceResponse = BaseGetServiceResponse &
  OneOf<{
    vervService: VervAppService;
  }>;

export type GetService = Record<string, never>;

export type VervAppService = {
  name?: string;
  currentDeploymentId?: string;
  status?: DeploymentStatus;
};

export type CreateDeployRequestUpgrade = {
  deploymentId?: string;
  image?: string;
};

type BaseCreateDeployRequest = {
  serviceName?: string;
};

export type CreateDeployRequest = BaseCreateDeployRequest &
  OneOf<{
    new: VelezApiVelezApi.CreateSmerdRequest;
    upgrade: CreateDeployRequestUpgrade;
  }>;

export type CreateDeployResponse = Record<string, never>;

export type CreateDeploy = Record<string, never>;

export type DeploymentInfo = {
  id?: string;
  status?: DeploymentStatus;
  specId?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
  image?: string;
  triggeredBy?: string;
  imageDigest?: string;
};

export type ListDeploymentsRequest = {
  paging?: VelezApiVelezCommon.Paging;
  serviceName?: string;
};

export type ListDeploymentsResponse = {
  deployments?: DeploymentInfo[];
  total?: string;
};

export type ListDeployments = Record<string, never>;

export type ListServicesRequest = {
  paging?: VelezApiVelezCommon.Paging;
  searchPattern?: string;
};

export type ListServicesResponse = {
  Total?: string;
  services?: ServiceBaseInfo[];
};

export type ListServices = Record<string, never>;

export type ServiceBaseInfo = {
  name?: string;
  lastDeployedAt?: GoogleProtobufTimestamp.Timestamp;
  imageName?: string;
  status?: string;
  env?: string;
};

export type StopServiceRequest = {
  name?: string;
};

export type StopServiceResponse = Record<string, never>;

export type StopService = Record<string, never>;

export type RestartServiceRequest = {
  name?: string;
};

export type RestartServiceResponse = Record<string, never>;

export type RestartService = Record<string, never>;

export type GetServiceMetricsRequest = {
  serviceName?: string;
};

export type GetServiceMetricsResponse = {
  cpuPercent?: number;
  memMi?: string;
  memMaxMi?: string;
  replicasRunning?: number;
  replicasDesired?: number;
  uptimeSeconds?: string;
};

export type GetServiceMetrics = Record<string, never>;

export type BoundResource = {
  name?: string;
  resourceType?: string;
  status?: string;
};

export type GetServiceResourcesRequest = {
  serviceName?: string;
};

export type GetServiceResourcesResponse = {
  resources?: BoundResource[];
};

export type GetServiceResources = Record<string, never>;

export type ServiceDependencyInfo = {
  serviceName?: string;
  proto?: string;
  requestRate?: number;
};

export type GetServiceGraphRequest = {
  serviceName?: string;
};

export type GetServiceGraphResponse = {
  callers?: ServiceDependencyInfo[];
  dependencies?: ServiceDependencyInfo[];
};

export type GetServiceGraph = Record<string, never>;

export type ServiceEnvironmentInfo = {
  env?: string;
  status?: string;
  deployedVersion?: string;
  deployedAt?: GoogleProtobufTimestamp.Timestamp;
  health?: string;
};

export type GetServiceEnvironmentsRequest = {
  serviceName?: string;
};

export type GetServiceEnvironmentsResponse = {
  environments?: ServiceEnvironmentInfo[];
};

export type GetServiceEnvironments = Record<string, never>;

export type GetVervonomiconRequest = {
  serviceName?: string;
};

export type GetVervonomiconResponse = Record<string, never>;

export type GetVervonomicon = Record<string, never>;

export class ServiceApi {
  static CreateService(this:void, req: CreateServiceRequest, initReq?: fm.InitReq): Promise<CreateServiceResponse> {
    return fm.fetchRequest<CreateServiceResponse>(`/api/service/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetService(this:void, req: GetServiceRequest, initReq?: fm.InitReq): Promise<GetServiceResponse> {
    return fm.fetchRequest<GetServiceResponse>(`/api/service/get`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static CreateDeploy(this:void, req: CreateDeployRequest, initReq?: fm.InitReq): Promise<CreateDeployResponse> {
    return fm.fetchRequest<CreateDeployResponse>(`/api/service/deploy/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListDeployments(this:void, req: ListDeploymentsRequest, initReq?: fm.InitReq): Promise<ListDeploymentsResponse> {
    return fm.fetchRequest<ListDeploymentsResponse>(`/api/service/deploy/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListServices(this:void, req: ListServicesRequest, initReq?: fm.InitReq): Promise<ListServicesResponse> {
    return fm.fetchRequest<ListServicesResponse>(`/api/service/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static StopService(this:void, req: StopServiceRequest, initReq?: fm.InitReq): Promise<StopServiceResponse> {
    return fm.fetchRequest<StopServiceResponse>(`/api/service/stop`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static RestartService(this:void, req: RestartServiceRequest, initReq?: fm.InitReq): Promise<RestartServiceResponse> {
    return fm.fetchRequest<RestartServiceResponse>(`/api/service/restart`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetServiceMetrics(this:void, req: GetServiceMetricsRequest, initReq?: fm.InitReq): Promise<GetServiceMetricsResponse> {
    return fm.fetchRequest<GetServiceMetricsResponse>(`/api/service/metrics`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetServiceResources(this:void, req: GetServiceResourcesRequest, initReq?: fm.InitReq): Promise<GetServiceResourcesResponse> {
    return fm.fetchRequest<GetServiceResourcesResponse>(`/api/service/resources`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetServiceGraph(this:void, req: GetServiceGraphRequest, initReq?: fm.InitReq): Promise<GetServiceGraphResponse> {
    return fm.fetchRequest<GetServiceGraphResponse>(`/api/service/graph`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetServiceEnvironments(this:void, req: GetServiceEnvironmentsRequest, initReq?: fm.InitReq): Promise<GetServiceEnvironmentsResponse> {
    return fm.fetchRequest<GetServiceEnvironmentsResponse>(`/api/service/environments`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetVervonomicon(this:void, req: GetVervonomiconRequest, initReq?: fm.InitReq): Promise<GetVervonomiconResponse> {
    return fm.fetchRequest<GetVervonomiconResponse>(`/api/service/vervonomicon`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
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

export type CreateServiceRequest = {
  name?: string;
};

export type CreateServiceResponse = Record<string, never>;

export type CreateService = Record<string, never>;

type BaseGetServiceRequest = {
};

export type GetServiceRequest = BaseGetServiceRequest &
  OneOf<{
    id: string;
    name: string;
  }>;

type BaseGetServiceResponse = {
};

export type GetServiceResponse = BaseGetServiceResponse &
  OneOf<{
    vervService: VervAppService;
  }>;

export type GetService = Record<string, never>;

export type VervAppService = {
  id?: string;
  name?: string;
};

export type CreateDeployRequestUpgrade = {
  deploymentId?: string;
  image?: string;
};

type BaseCreateDeployRequest = {
  serviceId?: string;
};

export type CreateDeployRequest = BaseCreateDeployRequest &
  OneOf<{
    new: VelezApiVelezApi.CreateSmerdRequest;
    upgrade: CreateDeployRequestUpgrade;
  }>;

export type CreateDeployResponse = Record<string, never>;

export type CreateDeploy = Record<string, never>;

export type ListDeploymentsRequest = {
  paging?: VelezApiVelezCommon.Paging;
  serviceId?: string;
};

export type ListDeploymentsResponse = Record<string, never>;

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
};

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
}
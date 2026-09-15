/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";


export type RegistryInstance = {
  name?: string;
  port?: number;
  uiPort?: number;
  username?: string;
  environment?: string;
  status?: string;
  ownerService?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
  updatedAt?: GoogleProtobufTimestamp.Timestamp;
};

export type ListRegistryInstancesRequest = {
  paging?: VelezApiVelezCommon.Paging;
};

export type ListRegistryInstancesResponse = {
  instances?: RegistryInstance[];
  total?: string;
};

export type ListRegistryInstances = Record<string, never>;

export type CreateRegistryInstanceRequest = {
  name?: string;
  environment?: string;
  box?: string;
  exposeToPort?: number;
  ownerService?: string;
  enableUi?: boolean;
};

export type CreateRegistryInstanceResponse = {
  instance?: RegistryInstance;
  entityId?: string;
  action?: string;
};

export type CreateRegistryInstance = Record<string, never>;

export type DropRegistryInstanceRequest = {
  name?: string;
};

export type DropRegistryInstanceResponse = Record<string, never>;

export type DropRegistryInstance = Record<string, never>;

export type GetRegistryInstanceCredentialsRequest = {
  name?: string;
};

export type GetRegistryInstanceCredentialsResponse = {
  username?: string;
  password?: string;
  registryUrl?: string;
};

export type GetRegistryInstanceCredentials = Record<string, never>;

export class ContainerRegistryAPI {
  static ListRegistryInstances(this:void, req: ListRegistryInstancesRequest, initReq?: fm.InitReq): Promise<ListRegistryInstancesResponse> {
    return fm.fetchRequest<ListRegistryInstancesResponse>(`/api/container-registry/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static CreateRegistryInstance(this:void, req: CreateRegistryInstanceRequest, initReq?: fm.InitReq): Promise<CreateRegistryInstanceResponse> {
    return fm.fetchRequest<CreateRegistryInstanceResponse>(`/api/container-registry/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DropRegistryInstance(this:void, req: DropRegistryInstanceRequest, initReq?: fm.InitReq): Promise<DropRegistryInstanceResponse> {
    return fm.fetchRequest<DropRegistryInstanceResponse>(`/api/container-registry/drop`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetRegistryInstanceCredentials(this:void, req: GetRegistryInstanceCredentialsRequest, initReq?: fm.InitReq): Promise<GetRegistryInstanceCredentialsResponse> {
    return fm.fetchRequest<GetRegistryInstanceCredentialsResponse>(`/api/container-registry/${req.name}/credentials?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"});
  }
}
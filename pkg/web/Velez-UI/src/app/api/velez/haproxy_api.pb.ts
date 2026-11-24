/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";


export type HaproxyInstance = {
  name?: string;
  masterPgInstance?: string;
  port?: number;
  status?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
  environment?: string;
};

export type CreateHaproxyInstanceRequest = {
  name?: string;
  masterPgInstance?: string;
  port?: number;
  environment?: string;
  box?: string;
};

export type CreateHaproxyInstanceResponse = {
  entityId?: string;
  action?: string;
};

export type CreateHaproxyInstance = Record<string, never>;

export type ListHaproxyInstancesRequest = {
  paging?: VelezApiVelezCommon.Paging;
};

export type ListHaproxyInstancesResponse = {
  instances?: HaproxyInstance[];
  total?: string;
  provisioning?: VelezApiVelezCommon.ProvisioningTask[];
};

export type ListHaproxyInstances = Record<string, never>;

export type DropHaproxyInstanceRequest = {
  name?: string;
};

export type DropHaproxyInstanceResponse = {
  entityId?: string;
  action?: string;
};

export type DropHaproxyInstance = Record<string, never>;

export type SelectHaproxyMasterRequest = {
  name?: string;
  masterPgInstance?: string;
};

export type SelectHaproxyMasterResponse = {
  entityId?: string;
  action?: string;
};

export type SelectHaproxyMaster = Record<string, never>;

export class HaproxyAPI {
  static CreateHaproxyInstance(this:void, req: CreateHaproxyInstanceRequest, initReq?: fm.InitReq): Promise<CreateHaproxyInstanceResponse> {
    return fm.fetchRequest<CreateHaproxyInstanceResponse>(`/api/haproxy/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListHaproxyInstances(this:void, req: ListHaproxyInstancesRequest, initReq?: fm.InitReq): Promise<ListHaproxyInstancesResponse> {
    return fm.fetchRequest<ListHaproxyInstancesResponse>(`/api/haproxy/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DropHaproxyInstance(this:void, req: DropHaproxyInstanceRequest, initReq?: fm.InitReq): Promise<DropHaproxyInstanceResponse> {
    return fm.fetchRequest<DropHaproxyInstanceResponse>(`/api/haproxy/drop`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static SelectHaproxyMaster(this:void, req: SelectHaproxyMasterRequest, initReq?: fm.InitReq): Promise<SelectHaproxyMasterResponse> {
    return fm.fetchRequest<SelectHaproxyMasterResponse>(`/api/haproxy/select-master`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
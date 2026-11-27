/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };

type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (keyof T extends infer K
      ? K extends string & keyof T
        ? { [k in K]: T[K] } & Absent<T, K>
        : never
      : never);

export enum PgBackendRole {
  PRIMARY = "PRIMARY",
  BACKUP = "BACKUP",
  BALANCED = "BALANCED",
}

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
  connectionTarget?: HaproxyConnectionTarget;
  publishPort?: boolean;
  backends?: PgBackend[];
  sidecarContainerId?: string;
};

export type CreateHaproxyInstanceResponse = {
  entityId?: string;
  action?: string;
};

export type CreateHaproxyInstance = Record<string, never>;

export type HaproxyConnectionTargetEndpoint = {
  host?: string;
  port?: number;
};

export type HaproxyConnectionTarget = {
  dockerNetwork?: string;
  publicEndpoint?: HaproxyConnectionTargetEndpoint;
  tailscaleEndpoint?: HaproxyConnectionTargetEndpoint;
};

type BasePgBackend = {
  name?: string;
  user?: string;
  password?: string;
  role?: PgBackendRole;
  connectionTarget?: HaproxyConnectionTarget;environment?: string;
};

export type PgBackend = BasePgBackend &
  OneOf<{
    velezService: string;
    endpoint: HaproxyConnectionTargetEndpoint;
  }>;

export type BootstrapVcnSidecarRequest = {
  name?: string;
  authKey?: string;
  loginServerUrl?: string;
  hostname?: string;
  ports?: VelezApiVelezCommon.Port[];
  isAsync?: boolean;
};

export type BootstrapVcnSidecarResponse = {
  containerId?: string;
  entityId?: string;
  action?: string;
};

export type BootstrapVcnSidecar = Record<string, never>;

export type ValidatePgBackendRequest = {
  backend?: PgBackend;
  sidecarContainerId?: string;
};

export type ValidatePgBackendResponse = {
  ok?: boolean;
  errorMessage?: string;
};

export type ValidatePgBackend = Record<string, never>;

export type ValidateHaproxyTargetRequest = {
  masterPgInstance?: string;
  connectionTarget?: HaproxyConnectionTarget;
  environment?: string;
};

export type ValidateHaproxyTargetResponse = {
  ok?: boolean;
  errorMessage?: string;
  dockerNetwork?: string;
};

export type ValidateHaproxyTarget = Record<string, never>;

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
  static BootstrapVcnSidecar(this:void, req: BootstrapVcnSidecarRequest, initReq?: fm.InitReq): Promise<BootstrapVcnSidecarResponse> {
    return fm.fetchRequest<BootstrapVcnSidecarResponse>(`/api/haproxy/vcn-bootstrap`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ValidatePgBackend(this:void, req: ValidatePgBackendRequest, initReq?: fm.InitReq): Promise<ValidatePgBackendResponse> {
    return fm.fetchRequest<ValidatePgBackendResponse>(`/api/haproxy/validate-backend`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ValidateHaproxyTarget(this:void, req: ValidateHaproxyTargetRequest, initReq?: fm.InitReq): Promise<ValidateHaproxyTargetResponse> {
    return fm.fetchRequest<ValidateHaproxyTargetResponse>(`/api/haproxy/validate-target`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
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
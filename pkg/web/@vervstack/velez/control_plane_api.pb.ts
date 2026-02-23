/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };

type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (keyof T extends infer K
      ? K extends string & keyof T
        ? { [k in K]: T[K] } & Absent<T, K>
        : never
      : never);

export enum VervServiceType {
  unknown_service_type = "unknown_service_type",
  matreshka = "matreshka",
  makosh = "makosh",
  webserver = "webserver",
  headscale = "headscale",
  portainer = "portainer",
  statefull_pg = "statefull_pg",
}

export enum VervServiceState {
  unknown = "unknown",
  running = "running",
  warning = "warning",
  dead = "dead",
  disabled = "disabled",
}

export type ListVervServicesRequest = Record<string, never>;

export type ListVervServicesResponse = {
  services?: VervService[];
};

export type ListVervServices = Record<string, never>;

export type VervService = {
  type?: VervServiceType;
  port?: number;
  state?: VervServiceState;
};

type BaseEnableServiceRequest = {
  service?: VervServiceType;
};

export type EnableServiceRequest = BaseEnableServiceRequest &
  OneOf<{
    statefullCluster: EnableStatefullCluster;
    headscaleServer: EnableHeadscaleServer;
  }>;

export type EnableServiceResponse = Record<string, never>;

export type EnableService = Record<string, never>;

export type InitMasterRequest = Record<string, never>;

export type InitMasterResponse = Record<string, never>;

export type InitMaster = Record<string, never>;

export type ConnectSlaveRequest = Record<string, never>;

export type ConnectSlaveResponse = Record<string, never>;

export type ConnectSlave = Record<string, never>;

export type EnableStatefullCluster = {
  isExposePort?: boolean;
  exposeToPort?: string;
};

export type EnableHeadscaleServerExternalHeadscaleConnection = {
  url?: string;
  token?: string;
};

export type EnableHeadscaleServerDeployHeadscaleConfig = {
  customPort?: number;
  customImage?: string;
};

type BaseEnableHeadscaleServer = {
};

export type EnableHeadscaleServer = BaseEnableHeadscaleServer &
  OneOf<{
    deployConfig: EnableHeadscaleServerDeployHeadscaleConfig;
    externalConnect: EnableHeadscaleServerExternalHeadscaleConnection;
  }>;

export class ControlPlaneAPI {
  static ListVervServices(this:void, req: ListVervServicesRequest, initReq?: fm.InitReq): Promise<ListVervServicesResponse> {
    return fm.fetchRequest<ListVervServicesResponse>(`/api/control_plane/services?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
  static EnableService(this:void, req: EnableServiceRequest, initReq?: fm.InitReq): Promise<EnableServiceResponse> {
    return fm.fetchRequest<EnableServiceResponse>(`/api/control_plane/service/enable`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ConnectSlave(this:void, req: ConnectSlaveRequest, initReq?: fm.InitReq): Promise<ConnectSlaveResponse> {
    return fm.fetchRequest<ConnectSlaveResponse>(`/api/slave/connect`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
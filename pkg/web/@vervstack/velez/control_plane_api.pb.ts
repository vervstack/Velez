/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as VelezApiVelezApi from "./velez_api.pb";


export enum VervServiceType {
  unknown_service_type = "unknown_service_type",
  matreshka = "matreshka",
  makosh = "makosh",
  webserver = "webserver",
  headscale = "headscale",
  portainer = "portainer",
  cluster_mode = "cluster_mode",
}

export type ListServicesRequest = Record<string, never>;

export type ListServicesResponse = {
  services?: Service[];
  inactiveServices?: Service[];
};

export type ListServices = Record<string, never>;

export type Service = {
  type?: VervServiceType;
  port?: number;
  constructor?: VelezApiVelezApi.CreateSmerdRequest;
  togglable?: boolean;
};

export type EnableServiceRequest = {
  service?: VervServiceType;
};

export type EnableServiceResponse = Record<string, never>;

export type EnableService = Record<string, never>;

export type InitMasterRequest = Record<string, never>;

export type InitMasterResponse = Record<string, never>;

export type InitMaster = Record<string, never>;

export type ConnectSlaveRequest = Record<string, never>;

export type ConnectSlaveResponse = Record<string, never>;

export type ConnectSlave = Record<string, never>;

export class ControlPlaneAPI {
  static ListServices(this:void, req: ListServicesRequest, initReq?: fm.InitReq): Promise<ListServicesResponse> {
    return fm.fetchRequest<ListServicesResponse>(`/api/control_plane/services?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
  static EnableService(this:void, req: EnableServiceRequest, initReq?: fm.InitReq): Promise<EnableServiceResponse> {
    return fm.fetchRequest<EnableServiceResponse>(`/api/control_plane/service/enable`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ConnectSlave(this:void, req: ConnectSlaveRequest, initReq?: fm.InitReq): Promise<ConnectSlaveResponse> {
    return fm.fetchRequest<ConnectSlaveResponse>(`/api/slave/connect`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
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
  svarog = "svarog",
  makosh = "makosh",
  webserver = "webserver",
  headscale = "headscale",
  portainer = "portainer",
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

export type EnableServicesRequest = {
    services?: VervServiceType[];
};

export type EnableServicesResponse = Record<string, never>;

export type EnableServices = Record<string, never>;

export class ControlPlaneAPI {
  static ListServices(this:void, req: ListServicesRequest, initReq?: fm.InitReq): Promise<ListServicesResponse> {
    return fm.fetchRequest<ListServicesResponse>(`/api/control_plane/services?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }

    static EnableServices(this: void, req: EnableServicesRequest, initReq?: fm.InitReq): Promise<EnableServicesResponse> {
        return fm.fetchRequest<EnableServicesResponse>(`/api/control_plane/services/enable`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
    }
}
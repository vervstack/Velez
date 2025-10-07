/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "../fetch.pb";


export enum ServiceType {
  unknown_service_type = "unknown_service_type",
  matreshka = "matreshka",
  svarog = "svarog",
  webserver = "webserver",
  makosh = "makosh",
  portainer = "portainer",
}

export type ListServicesRequest = Record<string, never>;

export type ListServicesResponse = {
  services?: Service[];
  inactiveServices?: Service[];
};

export type ListServices = Record<string, never>;

export type Service = {
  type?: ServiceType;
  port?: number;
};

export class ControlPlane {
  static ListServices(this:void, req: ListServicesRequest, initReq?: fm.InitReq): Promise<ListServicesResponse> {
    return fm.fetchRequest<ListServicesResponse>(`/api/control_plane/services?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
}
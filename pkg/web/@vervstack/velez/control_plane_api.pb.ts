/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";


export type ListServicesRequest = Record<string, never>;

export type ListServicesResponse = {
  matreshka?: Matreshka;
  makosh?: Makosh;
  svarog?: Svarog;
};

export type ListServices = Record<string, never>;

export type Matreshka = {
  uiUrl?: string;
};

export type Makosh = {
  uiUrl?: string;
};

export type Svarog = Record<string, never>;

export class ControlPlane {
  static ListServices(this:void, req: ListServicesRequest, initReq?: fm.InitReq): Promise<ListServicesResponse> {
    return fm.fetchRequest<ListServicesResponse>(`/api/control_plane/services?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
}
/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";


export enum AutoUpdateBumpLevel {
  AUTO_UPDATE_BUMP_LEVEL_UNSPECIFIED = "AUTO_UPDATE_BUMP_LEVEL_UNSPECIFIED",
  MAJOR = "MAJOR",
  MINOR = "MINOR",
  PATCH = "PATCH",
}

export type ListAutoUpdateServicesRequest = Record<string, never>;

export type ListAutoUpdateServicesResponseService = {
  name?: string;
  image?: string;
  isEnabled?: boolean;
  bumpLevel?: AutoUpdateBumpLevel;
};

export type ListAutoUpdateServicesResponse = {
  services?: ListAutoUpdateServicesResponseService[];
  checkPeriodSeconds?: string;
  isWorkerEnabled?: boolean;
};

export type ListAutoUpdateServices = Record<string, never>;

export type SetServiceAutoUpdateRequest = {
  name?: string;
  isEnabled?: boolean;
  bumpLevel?: AutoUpdateBumpLevel;
};

export type SetServiceAutoUpdateResponse = {
  entityId?: string;
  action?: string;
};

export type SetServiceAutoUpdate = Record<string, never>;

export class AutoUpdateAPI {
  static ListAutoUpdateServices(this:void, req: ListAutoUpdateServicesRequest, initReq?: fm.InitReq): Promise<ListAutoUpdateServicesResponse> {
    return fm.fetchRequest<ListAutoUpdateServicesResponse>(`/api/auto_update/services?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
  static SetServiceAutoUpdate(this:void, req: SetServiceAutoUpdateRequest, initReq?: fm.InitReq): Promise<SetServiceAutoUpdateResponse> {
    return fm.fetchRequest<SetServiceAutoUpdateResponse>(`/api/auto_update/set_service`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
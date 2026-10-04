/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";


export type Settings = {
  isSysboxEnabled?: boolean;
  isSysboxWhitelistIgnored?: boolean;
};

export type GetSettingsRequest = Record<string, never>;

export type GetSettingsResponse = {
  settings?: Settings;
};

export type GetSettings = Record<string, never>;

export type UpdateSettingsRequest = {
  isSysboxEnabled?: boolean;
  isSysboxWhitelistIgnored?: boolean;
};

export type UpdateSettingsResponse = {
  settings?: Settings;
};

export type UpdateSettings = Record<string, never>;

export type GetSysboxStatusRequest = Record<string, never>;

export type GetSysboxStatusResponse = {
  osType?: string;
  kernelVersion?: string;
  dockerVersion?: string;
  isRootless?: boolean;
  isSnap?: boolean;
  isRuntimeRegistered?: boolean;
  containersTotal?: number;
  containersOnSysbox?: number;
};

export type GetSysboxStatus = Record<string, never>;

export type RunSysboxSmokeTestRequest = Record<string, never>;

export type RunSysboxSmokeTestResponse = {
  isPassed?: boolean;
  failure?: string;
};

export type RunSysboxSmokeTest = Record<string, never>;

export class SettingsAPI {
  static GetSettings(this:void, req: GetSettingsRequest, initReq?: fm.InitReq): Promise<GetSettingsResponse> {
    return fm.fetchRequest<GetSettingsResponse>(`/api/settings/get?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
  static UpdateSettings(this:void, req: UpdateSettingsRequest, initReq?: fm.InitReq): Promise<UpdateSettingsResponse> {
    return fm.fetchRequest<UpdateSettingsResponse>(`/api/settings/update`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetSysboxStatus(this:void, req: GetSysboxStatusRequest, initReq?: fm.InitReq): Promise<GetSysboxStatusResponse> {
    return fm.fetchRequest<GetSysboxStatusResponse>(`/api/settings/sysbox_status?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
  static RunSysboxSmokeTest(this:void, req: RunSysboxSmokeTestRequest, initReq?: fm.InitReq): Promise<RunSysboxSmokeTestResponse> {
    return fm.fetchRequest<RunSysboxSmokeTestResponse>(`/api/settings/sysbox_smoke_test`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
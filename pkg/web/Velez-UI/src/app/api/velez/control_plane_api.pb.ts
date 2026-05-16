/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };

type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (keyof T extends infer K
      ? K extends string & keyof T
        ? { [k in K]: T[K] } & Absent<T, K>
        : never
      : never);

export enum VervPluginType {
  unknown_service_type = "unknown_service_type",
  matreshka = "matreshka",
  makosh = "makosh",
  webserver = "webserver",
  headscale = "headscale",
  portainer = "portainer",
  statefull_pg = "statefull_pg",
}

export enum VervPluginState {
  unknown = "unknown",
  running = "running",
  warning = "warning",
  dead = "dead",
  disabled = "disabled",
}

export type ListVervPluginsRequest = Record<string, never>;

export type ListVervPluginsResponse = {
  plugins?: VervPlugin[];
};

export type ListVervPlugins = Record<string, never>;

export type VervPlugin = {
  type?: VervPluginType;
  port?: number;
  state?: VervPluginState;
};

type BaseEnablePluginRequest = {
  plugin?: VervPluginType;
};

export type EnablePluginRequest = BaseEnablePluginRequest &
  OneOf<{
    statefullCluster: EnableStatefullCluster;
    headscaleServer: EnableHeadscaleServer;
  }>;

export type EnablePluginResponse = Record<string, never>;

export type EnablePlugin = Record<string, never>;

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

export type ListNodesRequest = {
  paging?: VelezApiVelezCommon.Paging;
};

export type ListNodesResponse = {
  nodes?: VelezApiVelezCommon.NodeBaseInfo[];
  total?: string;
};

export type ListNodes = Record<string, never>;

export type Plugin = {
  type?: VervPluginType;
  state?: VervPluginState;
  port?: number;
};

export type ListPluginsRequest = Record<string, never>;

export type ListPluginsResponse = {
  plugins?: Plugin[];
};

export type ListPlugins = Record<string, never>;

export class ControlPlaneAPI {
  static EnablePlugin(this:void, req: EnablePluginRequest, initReq?: fm.InitReq): Promise<EnablePluginResponse> {
    return fm.fetchRequest<EnablePluginResponse>(`/api/control_plane/plugin/enable`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ConnectSlave(this:void, req: ConnectSlaveRequest, initReq?: fm.InitReq): Promise<ConnectSlaveResponse> {
    return fm.fetchRequest<ConnectSlaveResponse>(`/api/control_plane/slave/connect`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListNodes(this:void, req: ListNodesRequest, initReq?: fm.InitReq): Promise<ListNodesResponse> {
    return fm.fetchRequest<ListNodesResponse>(`/api/control_plane/nodes/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListPlugins(this:void, req: ListPluginsRequest, initReq?: fm.InitReq): Promise<ListPluginsResponse> {
    return fm.fetchRequest<ListPluginsResponse>(`/api/control_plane/plugins?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
}
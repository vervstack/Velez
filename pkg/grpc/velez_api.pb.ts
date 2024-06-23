/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "../fetch.pb";
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb";


export enum PortBindingsProtocol {
  unknown = "unknown",
  tcp = "tcp",
  udp = "udp",
}

export enum SmerdStatus {
  unknown = "unknown",
  created = "created",
  restarting = "restarting",
  running = "running",
  removing = "removing",
  paused = "paused",
  exited = "exited",
  dead = "dead",
}

export type VersionRequest = Record<string, never>;

export type VersionResponse = {
  version?: string;
};

export type Version = Record<string, never>;

export type PortBindings = {
  host?: number;
  container?: number;
  protoc?: PortBindingsProtocol;
};

export type MountBindings = {
  host?: string;
  container?: string;
};

export type VolumeBindings = {
  volume?: string;
  containerPath?: string;
};

export type NetworkBind = {
  networkName?: string;
  aliases?: string[];
};

export type Image = {
  name?: string;
  tags?: string[];
  labels?: Record<string, string>;
};

export type Smerd = {
  uuid?: string;
  name?: string;
  imageName?: string;
  ports?: PortBindings[];
  volumes?: MountBindings[];
  status?: SmerdStatus;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
  networks?: NetworkBind[];
  labels?: Record<string, string>;
};

export type ContainerHardware = {
  cpuAmount?: number;
  ramMb?: number;
  memorySwapMb?: number;
};

export type ContainerSettings = {
  ports?: PortBindings[];
  mounts?: MountBindings[];
  networks?: NetworkBind[];
  volumes?: VolumeBindings[];
};

export type ContainerHealthcheck = {
  command?: string;
  intervalSecond?: number;
  timeoutSecond?: number;
  retries?: number;
};

export type Container = Record<string, never>;

export type CreateSmerdRequest = {
  name?: string;
  imageName?: string;
  hardware?: ContainerHardware;
  settings?: ContainerSettings;
  command?: string;
  env?: Record<string, string>;
  healthcheck?: ContainerHealthcheck;
};

export type CreateSmerd = Record<string, never>;

export type ListSmerdsRequest = {
  limit?: number;
  name?: string;
  id?: string;
};

export type ListSmerdsResponse = {
  smerds?: Smerd[];
};

export type ListSmerds = Record<string, never>;

export type DropSmerdRequest = {
  uuids?: string[];
  name?: string[];
};

export type DropSmerdResponseError = {
  uuid?: string;
  cause?: string;
};

export type DropSmerdResponse = {
  failed?: DropSmerdResponseError[];
  successful?: string[];
};

export type DropSmerd = Record<string, never>;

export type GetHardwareRequest = Record<string, never>;

export type GetHardwareResponseValue = {
  value?: string;
  err?: string;
};

export type GetHardwareResponse = {
  cpu?: GetHardwareResponseValue;
  diskMem?: GetHardwareResponseValue;
  ram?: GetHardwareResponseValue;
  portsAvailable?: number[];
  portsOccupied?: number[];
};

export type GetHardware = Record<string, never>;

export type FetchConfigRequest = {
  imageName?: string;
  serviceName?: string;
};

export type FetchConfigResponse = Record<string, never>;

export type FetchConfig = Record<string, never>;

export class VelezAPI {
  static Version(this:void, req: VersionRequest, initReq?: fm.InitReq): Promise<VersionResponse> {
    return fm.fetchRequest<VersionResponse>(`/version?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
  static CreateSmerd(this:void, req: CreateSmerdRequest, initReq?: fm.InitReq): Promise<Smerd> {
    return fm.fetchRequest<Smerd>(`/smerd/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListSmerds(this:void, req: ListSmerdsRequest, initReq?: fm.InitReq): Promise<ListSmerdsResponse> {
    return fm.fetchRequest<ListSmerdsResponse>(`/smerd/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DropSmerd(this:void, req: DropSmerdRequest, initReq?: fm.InitReq): Promise<DropSmerdResponse> {
    return fm.fetchRequest<DropSmerdResponse>(`/smerd/drop`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetHardware(this:void, req: GetHardwareRequest, initReq?: fm.InitReq): Promise<GetHardwareResponse> {
    return fm.fetchRequest<GetHardwareResponse>(`/hardware?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
  static FetchConfig(this:void, req: FetchConfigRequest, initReq?: fm.InitReq): Promise<FetchConfigResponse> {
    return fm.fetchRequest<FetchConfigResponse>(`/smerd/fetch-config`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
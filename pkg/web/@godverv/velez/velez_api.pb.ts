/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";


export enum PortProtocol {
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

export type Port = {
  servicePortNumber?: number;
  protocol?: PortProtocol;
  exposedTo?: number;
};

export type Volume = {
  volumeName?: string;
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
  ports?: Port[];
  volumes?: Volume[];
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
  ports?: Port[];
  volumes?: Volume[];
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
  labels?: Record<string, string>;
  ignoreConfig?: boolean;
  useImagePorts?: boolean;
};

export type CreateSmerd = Record<string, never>;

export type ListSmerdsRequest = {
  limit?: number;
  name?: string;
  id?: string;
  label?: Record<string, string>;
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

export type FetchConfigResponse = {
  config?: Uint8Array;
};

export type FetchConfig = Record<string, never>;

export class VelezAPI {
  static Version(this:void, req: VersionRequest, initReq?: fm.InitReq): Promise<VersionResponse> {
    return fm.fetchRequest<VersionResponse>(`/api/version?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
  static CreateSmerd(this:void, req: CreateSmerdRequest, initReq?: fm.InitReq): Promise<Smerd> {
    return fm.fetchRequest<Smerd>(`/api/smerd/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListSmerds(this:void, req: ListSmerdsRequest, initReq?: fm.InitReq): Promise<ListSmerdsResponse> {
    return fm.fetchRequest<ListSmerdsResponse>(`/api/smerd/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DropSmerd(this:void, req: DropSmerdRequest, initReq?: fm.InitReq): Promise<DropSmerdResponse> {
    return fm.fetchRequest<DropSmerdResponse>(`/api/smerd/drop`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetHardware(this:void, req: GetHardwareRequest, initReq?: fm.InitReq): Promise<GetHardwareResponse> {
    return fm.fetchRequest<GetHardwareResponse>(`/api/hardware?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
  static FetchConfig(this:void, req: FetchConfigRequest, initReq?: fm.InitReq): Promise<FetchConfigResponse> {
    return fm.fetchRequest<FetchConfigResponse>(`/api/smerd/fetch-config`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
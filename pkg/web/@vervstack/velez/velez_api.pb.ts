/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";


export type VersionRequest = Record<string, never>;

export type VersionResponse = {
  version?: string;
};

export type Version = Record<string, never>;

export type CreateSmerdRequest = {
  name?: string;
  imageName?: string;
    hardware?: VelezApiVelezCommon.ContainerHardware;
    settings?: VelezApiVelezCommon.ContainerSettings;
  command?: string;
  env?: Record<string, string>;
    healthcheck?: VelezApiVelezCommon.ContainerHealthcheck;
  labels?: Record<string, string>;
  ignoreConfig?: boolean;
  useImagePorts?: boolean;
  autoUpgrade?: boolean;
    restart?: VelezApiVelezCommon.RestartPolicy;
    config?: VelezApiVelezCommon.MatreshkaConfigSpec;
};

export type CreateSmerd = Record<string, never>;

export type ListSmerdsRequest = {
  limit?: number;
  name?: string;
  id?: string;
  label?: Record<string, string>;
};

export type ListSmerdsResponse = {
    smerds?: VelezApiVelezCommon.Smerd[];
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

export type AssembleConfigRequest = {
  imageName?: string;
  serviceName?: string;
};

export type AssembleConfigResponse = {
  config?: Uint8Array;
};

export type AssembleConfig = Record<string, never>;

export type UpgradeSmerdRequest = {
  name?: string;
  image?: string;
};

export type UpgradeSmerdResponse = Record<string, never>;

export type UpgradeSmerd = Record<string, never>;

export type MakeConnectionsRequest = {
    connections?: VelezApiVelezCommon.Connection[];
};

export type MakeConnectionsResponse = Record<string, never>;

export type MakeConnections = Record<string, never>;

export type BreakConnectionsRequest = {
    connections?: VelezApiVelezCommon.Connection[];
};

export type BreakConnectionsResponse = Record<string, never>;

export type BreakConnections = Record<string, never>;

export type ListImagesRequest = {
    useRegistry?: boolean;
};

export type ListImagesResponse = {
    images?: VelezApiVelezCommon.ImageListItem[];
};

export type ListImages = Record<string, never>;

export class VelezAPI {
  static Version(this:void, req: VersionRequest, initReq?: fm.InitReq): Promise<VersionResponse> {
    return fm.fetchRequest<VersionResponse>(`/api/version?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }

    static CreateSmerd(this: void, req: CreateSmerdRequest, initReq?: fm.InitReq): Promise<VelezApiVelezCommon.Smerd> {
        return fm.fetchRequest<VelezApiVelezCommon.Smerd>(`/api/smerd/create`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
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
  static UpgradeSmerd(this:void, req: UpgradeSmerdRequest, initReq?: fm.InitReq): Promise<UpgradeSmerdResponse> {
    return fm.fetchRequest<UpgradeSmerdResponse>(`/api/smerd/upgrade`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static AssembleConfig(this:void, req: AssembleConfigRequest, initReq?: fm.InitReq): Promise<AssembleConfigResponse> {
    return fm.fetchRequest<AssembleConfigResponse>(`/api/config/assemble`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static MakeConnections(this:void, req: MakeConnectionsRequest, initReq?: fm.InitReq): Promise<MakeConnectionsResponse> {
    return fm.fetchRequest<MakeConnectionsResponse>(`/api/smerd/connect`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static BreakConnections(this:void, req: BreakConnectionsRequest, initReq?: fm.InitReq): Promise<BreakConnectionsResponse> {
    return fm.fetchRequest<BreakConnectionsResponse>(`/api/smerd/disconnect`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }

    static ListImages(this: void, req: ListImagesRequest, initReq?: fm.InitReq): Promise<ListImagesResponse> {
        return fm.fetchRequest<ListImagesResponse>(`/api/images?${fm.renderURLSearchParams(req, [])}`, {
            ...initReq,
            method: "GET"
        });
    }
}
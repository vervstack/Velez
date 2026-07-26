/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as VelezApiControlPlaneApi from "./control_plane_api.pb";
import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";
import * as VelezApiVelezApi from "./velez_api.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";


export enum TaskStatusStatus {
  UNKNOWN = "UNKNOWN",
  PENDING = "PENDING",
  RUNNING = "RUNNING",
  DONE = "DONE",
  FAILED = "FAILED",
}

export type WatchTaskRequest = {
  entityId?: string;
  action?: string;
};

export type WatchTask = Record<string, never>;

export type TaskStatus = {
  status?: TaskStatusStatus;
  error?: string;
  updatedAt?: GoogleProtobufTimestamp.Timestamp;
};

export type CreateSmerdTaskPayload = {
  request?: VelezApiVelezApi.CreateSmerdRequest;
  imageId?: string;
  containerId?: string;
  imageLabels?: Record<string, string>;
  imageTags?: string[];
  imageExposedPorts?: string[];
  pathToFiles?: Record<string, Uint8Array>;
};

export type CreateServiceTaskPayload = {
  name?: string;
};

export type AssembleConfigTaskPayload = {
  serviceName?: string;
  imageName?: string;
  imageLabels?: Record<string, string>;
  imageTags?: string[];
  containerId?: string;
  configName?: string;
  configVersion?: string;
  confType?: string;
  configFormat?: VelezApiVelezCommon.ConfigFormat;
  contentRaw?: Uint8Array;
  content?: Uint8Array;
};

export type CopyToVolumeTaskPayload = {
  volumeName?: string;
  pathToFiles?: Record<string, Uint8Array>;
  containerId?: string;
};

export type ConnectServiceToVpnTaskPayload = {
  serviceName?: string;
  namespaceId?: string;
  clientKey?: string;
  loginServerUrl?: string;
  containerId?: string;
};

export type EnableStatefullTaskPayload = {
  request?: VelezApiControlPlaneApi.EnableStatefullCluster;
  rootPwd?: string;
  userPwd?: string;
  containerId?: string;
  rootDsn?: string;
};

export type DropSmerdTaskPayload = {
  request?: VelezApiVelezApi.DropSmerdRequest;
  failed?: VelezApiVelezApi.DropSmerdResponseError[];
  successful?: string[];
};

export type UpgradeSmerdTaskPayload = {
  upgradeRequest?: VelezApiVelezApi.UpgradeSmerdRequest;
  request?: VelezApiVelezApi.CreateSmerdRequest;
  oldContainerId?: string;
  imageLabels?: Record<string, string>;
  imageTags?: string[];
  containerId?: string;
};

export class TasksApi {
  static WatchTask(this:void, req: WatchTaskRequest, entityNotifier?: fm.NotifyStreamEntityArrival<TaskStatus>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<TaskStatus>(`/api/tasks/watch?${fm.renderURLSearchParams(req, [])}`, entityNotifier, {...initReq, method: "GET"});
  }
  static CreateSmerdStream(this:void, req: VelezApiVelezApi.CreateSmerdRequest, entityNotifier?: fm.NotifyStreamEntityArrival<TaskStatus>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<TaskStatus>(`/api/smerd/create/stream`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
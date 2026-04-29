/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";


export enum RestartPolicyType {
  unless_stopped = "unless_stopped",
  no = "no",
  always = "always",
  on_failure = "on_failure",
}

export enum ConfigFormat {
  yaml = "yaml",
  env = "env",
}

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

export type SearchImageItem = {
  name?: string;
  imageUrl?: string;
  latestTag?: string;
};

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
  env?: Record<string, string>;
};

export type ContainerHardware = {
  cpu?: number;
  ramMb?: number;
  memorySwapMb?: number;
};

export type ContainerSettings = {
  ports?: Port[];
  network?: NetworkBind[];
  volumes?: Volume[];
};

export type ContainerHealthcheck = {
  command?: string;
  intervalSecond?: number;
  timeoutSecond?: number;
  retries?: number;
};

export type Container = Record<string, never>;

export type RestartPolicy = {
  type?: RestartPolicyType;
  FailureCount?: number;
};

export type MatreshkaConfigSpec = {
  configName?: string;
  configVersion?: string;
  configFormat?: ConfigFormat;
  systemPath?: string;
};

export type Connection = {
  serviceName?: string;
  targetNetwork?: string;
  aliases?: string[];
};

export type Paging = {
  limit?: string;
  offset?: string;
};
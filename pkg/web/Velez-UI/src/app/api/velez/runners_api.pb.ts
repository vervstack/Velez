/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";


export enum RunnerProvider {
  RUNNER_PROVIDER_UNSPECIFIED = "RUNNER_PROVIDER_UNSPECIFIED",
  GITHUB = "GITHUB",
}

export enum RunnerScope {
  RUNNER_SCOPE_UNSPECIFIED = "RUNNER_SCOPE_UNSPECIFIED",
  REPO = "REPO",
  ORG = "ORG",
}

export type Runner = {
  name?: string;
  provider?: RunnerProvider;
  scope?: RunnerScope;
  target?: string;
  labels?: string[];
  environment?: string;
  status?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
  updatedAt?: GoogleProtobufTimestamp.Timestamp;
};

export type ListRunnersRequest = {
  paging?: VelezApiVelezCommon.Paging;
};

export type ListRunnersResponse = {
  runners?: Runner[];
  total?: string;
};

export type ListRunners = Record<string, never>;

export type CreateRunnerRequest = {
  name?: string;
  provider?: RunnerProvider;
  scope?: RunnerScope;
  target?: string;
  labels?: string[];
  accessToken?: string;
  environment?: string;
};

export type CreateRunnerResponse = {
  runner?: Runner;
};

export type CreateRunner = Record<string, never>;

export type DropRunnerRequest = {
  name?: string;
};

export type DropRunnerResponse = Record<string, never>;

export type DropRunner = Record<string, never>;

export class RunnersAPI {
  static ListRunners(this:void, req: ListRunnersRequest, initReq?: fm.InitReq): Promise<ListRunnersResponse> {
    return fm.fetchRequest<ListRunnersResponse>(`/api/runners/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static CreateRunner(this:void, req: CreateRunnerRequest, initReq?: fm.InitReq): Promise<CreateRunnerResponse> {
    return fm.fetchRequest<CreateRunnerResponse>(`/api/runners/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DropRunner(this:void, req: DropRunnerRequest, initReq?: fm.InitReq): Promise<DropRunnerResponse> {
    return fm.fetchRequest<DropRunnerResponse>(`/api/runners/drop`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
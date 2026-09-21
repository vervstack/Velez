/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };

type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (keyof T extends infer K
      ? K extends string & keyof T
        ? { [k in K]: T[K] } & Absent<T, K>
        : never
      : never);

export enum RunnerProvider {
  RUNNER_PROVIDER_UNSPECIFIED = "RUNNER_PROVIDER_UNSPECIFIED",
  GITHUB = "GITHUB",
  GITLAB = "GITLAB",
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

type BaseCreateRunnerRequest = {
  name?: string;
  scope?: RunnerScope;
  target?: string;
  labels?: string[];environment?: string;dockerSocketAddress?: string;
};

export type CreateRunnerRequest = BaseCreateRunnerRequest &
  OneOf<{
    github: GithubConfig;
    gitlab: GitlabConfig;
  }>;

export type CreateRunnerResponse = {
  runner?: Runner;
  entityId?: string;
  action?: string;
};

export type CreateRunner = Record<string, never>;

export type GithubConfig = {
  accessToken?: string;
};

export type GitlabConfig = {
  accessToken?: string;
  baseUrl?: string;
  dockerImage?: string;
};

export type DropRunnerRequest = {
  name?: string;
};

export type DropRunnerResponse = Record<string, never>;

export type DropRunner = Record<string, never>;

export type GetRunnerCredentialsRequest = {
  name?: string;
};

export type GetRunnerCredentialsResponse = {
  token?: string;
  target?: string;
  provider?: RunnerProvider;
};

export type GetRunnerCredentials = Record<string, never>;

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
  static GetRunnerCredentials(this:void, req: GetRunnerCredentialsRequest, initReq?: fm.InitReq): Promise<GetRunnerCredentialsResponse> {
    return fm.fetchRequest<GetRunnerCredentialsResponse>(`/api/runners/${req.name}/credentials?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"});
  }
}
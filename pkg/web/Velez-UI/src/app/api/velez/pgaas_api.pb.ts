/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";


export enum PgInstanceIsolation {
  PG_INSTANCE_ISOLATION_UNSPECIFIED = "PG_INSTANCE_ISOLATION_UNSPECIFIED",
  PG_INSTANCE_ISOLATION_SEPARATE_INSTANCE = "PG_INSTANCE_ISOLATION_SEPARATE_INSTANCE",
  PG_INSTANCE_ISOLATION_SHARED_POOL = "PG_INSTANCE_ISOLATION_SHARED_POOL",
}

export type PgInstance = {
  name?: string;
  dbName?: string;
  username?: string;
  port?: number;
  environment?: string;
  status?: string;
  ownerService?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
  updatedAt?: GoogleProtobufTimestamp.Timestamp;
};

export type ListPgInstancesRequest = {
  paging?: VelezApiVelezCommon.Paging;
};

export type ListPgInstancesResponse = {
  instances?: PgInstance[];
  total?: string;
};

export type ListPgInstances = Record<string, never>;

export type CreatePgInstanceRequest = {
  name?: string;
  environment?: string;
  box?: string;
  exposeToPort?: number;
  ownerService?: string;
};

export type CreatePgInstanceResponse = {
  instance?: PgInstance;
};

export type CreatePgInstance = Record<string, never>;

export type DropPgInstanceRequest = {
  name?: string;
};

export type DropPgInstanceResponse = Record<string, never>;

export type DropPgInstance = Record<string, never>;

export type GetPgInstanceCredentialsRequest = {
  name?: string;
};

export type GetPgInstanceCredentialsResponse = {
  dbName?: string;
  username?: string;
  password?: string;
  dsn?: string;
};

export type GetPgInstanceCredentials = Record<string, never>;

export class PostgresAPI {
  static ListPgInstances(this:void, req: ListPgInstancesRequest, initReq?: fm.InitReq): Promise<ListPgInstancesResponse> {
    return fm.fetchRequest<ListPgInstancesResponse>(`/api/postgres/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static CreatePgInstance(this:void, req: CreatePgInstanceRequest, initReq?: fm.InitReq): Promise<CreatePgInstanceResponse> {
    return fm.fetchRequest<CreatePgInstanceResponse>(`/api/postgres/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DropPgInstance(this:void, req: DropPgInstanceRequest, initReq?: fm.InitReq): Promise<DropPgInstanceResponse> {
    return fm.fetchRequest<DropPgInstanceResponse>(`/api/postgres/drop`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetPgInstanceCredentials(this:void, req: GetPgInstanceCredentialsRequest, initReq?: fm.InitReq): Promise<GetPgInstanceCredentialsResponse> {
    return fm.fetchRequest<GetPgInstanceCredentialsResponse>(`/api/postgres/${req.name}/credentials?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"});
  }
}
/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";


export type DindInfo = {
  name?: string;
  address?: string;
  isSysboxEnabled?: boolean;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
};

export type CreateDindRequest = {
  name?: string;
  environment?: string;
  isSysboxEnabled?: boolean;
};

export type CreateDindResponse = {
  entityId?: string;
  action?: string;
};

export type CreateDind = Record<string, never>;

export type ListDindsRequest = Record<string, never>;

export type ListDindsResponse = {
  dinds?: DindInfo[];
  provisioning?: VelezApiVelezCommon.ProvisioningTask[];
};

export type ListDinds = Record<string, never>;

export type DropDindRequest = {
  name?: string;
};

export type DropDindResponse = {
  entityId?: string;
  action?: string;
};

export type DropDind = Record<string, never>;

export class DindAPI {
  static CreateDind(this:void, req: CreateDindRequest, initReq?: fm.InitReq): Promise<CreateDindResponse> {
    return fm.fetchRequest<CreateDindResponse>(`/api/dind/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListDinds(this:void, req: ListDindsRequest, initReq?: fm.InitReq): Promise<ListDindsResponse> {
    return fm.fetchRequest<ListDindsResponse>(`/api/dind/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DropDind(this:void, req: DropDindRequest, initReq?: fm.InitReq): Promise<DropDindResponse> {
    return fm.fetchRequest<DropDindResponse>(`/api/dind/drop`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };

type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (keyof T extends infer K
      ? K extends string & keyof T
        ? { [k in K]: T[K] } & Absent<T, K>
        : never
      : never);

export enum TransferStatus {
  TRANSFER_STATUS_UNSPECIFIED = "TRANSFER_STATUS_UNSPECIFIED",
  TRANSFER_STATUS_PENDING = "TRANSFER_STATUS_PENDING",
  TRANSFER_STATUS_RUNNING = "TRANSFER_STATUS_RUNNING",
  TRANSFER_STATUS_FAILED = "TRANSFER_STATUS_FAILED",
  TRANSFER_STATUS_DONE = "TRANSFER_STATUS_DONE",
  TRANSFER_STATUS_CANCELLED = "TRANSFER_STATUS_CANCELLED",
}

export enum TransferItemKind {
  TRANSFER_ITEM_KIND_UNSPECIFIED = "TRANSFER_ITEM_KIND_UNSPECIFIED",
  TRANSFER_ITEM_KIND_IMAGE = "TRANSFER_ITEM_KIND_IMAGE",
  TRANSFER_ITEM_KIND_VOLUME = "TRANSFER_ITEM_KIND_VOLUME",
}

export type CreateTransferTokenRequest = Record<string, never>;

export type CreateTransferTokenResponse = {
  token?: string;
  expiresAt?: GoogleProtobufTimestamp.Timestamp;
};

export type CreateTransferToken = Record<string, never>;

export type PlanTransferRequest = {
  containerName?: string;
};

export type PlanTransferResponse = {
  sidecars?: PlanTransferSidecarCandidate[];
  consumers?: PlanTransferNetworkConsumer[];
  unsupportedMounts?: string[];
};

export type PlanTransferSidecarCandidate = {
  containerName?: string;
  imageName?: string;
};

export type PlanTransferNetworkConsumer = {
  containerName?: string;
  networkName?: string;
};

export type PlanTransfer = Record<string, never>;

export type StartTransferRequest = {
  containerName?: string;
  sidecarNames?: string[];
  address?: string;
  token?: string;
  isConsumersAcknowledged?: boolean;
};

export type StartTransferResponse = {
  transferId?: string;
};

export type StartTransfer = Record<string, never>;

export type CancelTransferRequest = {
  transferId?: string;
};

export type CancelTransferResponse = Record<string, never>;

export type CancelTransfer = Record<string, never>;

export type ListTransfersRequest = Record<string, never>;

export type ListTransfersResponse = {
  transfers?: TransferInfo[];
};

export type ListTransfers = Record<string, never>;

export type TransferInfo = {
  id?: string;
  isSource?: boolean;
  containerName?: string;
  peerAddress?: string;
  status?: TransferStatus;
  error?: string;
  totalBytes?: string;
  doneBytes?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
  updatedAt?: GoogleProtobufTimestamp.Timestamp;
};

type BaseTransferRequest = {
};

export type TransferRequest = BaseTransferRequest &
  OneOf<{
    begin: TransferBegin;
    itemStart: TransferItemStart;
    batch: TransferBatch;
    end: TransferEnd;
  }>;

type BaseTransferResponse = {
};

export type TransferResponse = BaseTransferResponse &
  OneOf<{
    accepted: TransferAccepted;
    itemApplied: TransferItemApplied;
    finished: TransferFinished;
  }>;

export type TransferBegin = {
  token?: string;
  containers?: TransferContainerManifest[];
  items?: TransferItemManifest[];
};

export type TransferContainerManifest = {
  name?: string;
  networkOwner?: string;
  inspectJson?: Uint8Array;
};

export type TransferItemManifest = {
  name?: string;
  kind?: TransferItemKind;
  totalBytes?: string;
};

export type TransferItemStart = {
  name?: string;
};

export type TransferBatch = {
  data?: Uint8Array;
};

export type TransferEnd = Record<string, never>;

export type TransferAccepted = Record<string, never>;

export type TransferItemApplied = {
  name?: string;
};

export type TransferFinished = Record<string, never>;

export type Transfer = Record<string, never>;

export class TransferAPI {
  static CreateTransferToken(this:void, req: CreateTransferTokenRequest, initReq?: fm.InitReq): Promise<CreateTransferTokenResponse> {
    return fm.fetchRequest<CreateTransferTokenResponse>(`/api/transfer/token/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static PlanTransfer(this:void, req: PlanTransferRequest, initReq?: fm.InitReq): Promise<PlanTransferResponse> {
    return fm.fetchRequest<PlanTransferResponse>(`/api/transfer/plan`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static StartTransfer(this:void, req: StartTransferRequest, initReq?: fm.InitReq): Promise<StartTransferResponse> {
    return fm.fetchRequest<StartTransferResponse>(`/api/transfer/start`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static CancelTransfer(this:void, req: CancelTransferRequest, initReq?: fm.InitReq): Promise<CancelTransferResponse> {
    return fm.fetchRequest<CancelTransferResponse>(`/api/transfer/cancel`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListTransfers(this:void, req: ListTransfersRequest, initReq?: fm.InitReq): Promise<ListTransfersResponse> {
    return fm.fetchRequest<ListTransfersResponse>(`/api/transfer/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
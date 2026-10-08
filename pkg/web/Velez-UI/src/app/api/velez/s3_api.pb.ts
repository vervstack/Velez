/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";
import * as GoogleProtobufTimestamp from "./google/protobuf/timestamp.pb";
import * as VelezApiVelezCommon from "./velez_common.pb";


export type S3Instance = {
  name?: string;
  s3Port?: number;
  webUiPort?: number;
  replicationFactor?: number;
  region?: string;
  environment?: string;
  status?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
};

export type S3BucketAccess = {
  bucketName?: string;
  accessKeyId?: string;
  keyName?: string;
  isRead?: boolean;
  isWrite?: boolean;
  isOwner?: boolean;
};

export type S3Bucket = {
  id?: string;
  name?: string;
  access?: S3BucketAccess[];
  objectCount?: string;
  bytes?: string;
  ownerService?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
};

export type S3Key = {
  accessKeyId?: string;
  name?: string;
  access?: S3BucketAccess[];
  ownerService?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
};

export type CreateS3InstanceRequest = {
  name?: string;
  environment?: string;
  box?: string;
  exposeToPort?: number;
  replicationFactor?: number;
  region?: string;
  enableWebUi?: boolean;
};

export type CreateS3InstanceResponse = {
  entityId?: string;
  action?: string;
};

export type CreateS3Instance = Record<string, never>;

export type ListS3InstancesRequest = {
  paging?: VelezApiVelezCommon.Paging;
};

export type ListS3InstancesResponse = {
  instances?: S3Instance[];
  total?: string;
  provisioning?: VelezApiVelezCommon.ProvisioningTask[];
};

export type ListS3Instances = Record<string, never>;

export type DropS3InstanceRequest = {
  name?: string;
};

export type DropS3InstanceResponse = Record<string, never>;

export type DropS3Instance = Record<string, never>;

export type GetS3InstanceCredentialsRequest = {
  name?: string;
};

export type GetS3InstanceCredentialsResponse = {
  adminToken?: string;
  s3Endpoint?: string;
  internalS3Endpoint?: string;
  region?: string;
  webUiUrl?: string;
  webUiUsername?: string;
  webUiPassword?: string;
};

export type GetS3InstanceCredentials = Record<string, never>;

export type ListS3BucketsRequest = {
  instanceName?: string;
};

export type ListS3BucketsResponse = {
  buckets?: S3Bucket[];
};

export type ListS3Buckets = Record<string, never>;

export type CreateS3BucketRequest = {
  instanceName?: string;
  bucketName?: string;
  ownerService?: string;
};

export type CreateS3BucketResponse = {
  bucket?: S3Bucket;
};

export type CreateS3Bucket = Record<string, never>;

export type DeleteS3BucketRequest = {
  instanceName?: string;
  bucketName?: string;
};

export type DeleteS3BucketResponse = Record<string, never>;

export type DeleteS3Bucket = Record<string, never>;

export type SetS3BucketAccessRequest = {
  instanceName?: string;
  access?: S3BucketAccess;
};

export type SetS3BucketAccessResponse = {
  bucket?: S3Bucket;
};

export type SetS3BucketAccess = Record<string, never>;

export type ListS3KeysRequest = {
  instanceName?: string;
};

export type ListS3KeysResponse = {
  keys?: S3Key[];
};

export type ListS3Keys = Record<string, never>;

export type CreateS3KeyRequest = {
  instanceName?: string;
  keyName?: string;
  access?: S3BucketAccess[];
};

export type CreateS3KeyResponse = {
  accessKeyId?: string;
  secretAccessKey?: string;
};

export type CreateS3Key = Record<string, never>;

export type DeleteS3KeyRequest = {
  instanceName?: string;
  accessKeyId?: string;
};

export type DeleteS3KeyResponse = Record<string, never>;

export type DeleteS3Key = Record<string, never>;

export type GetS3KeyCredentialsRequest = {
  instanceName?: string;
  accessKeyId?: string;
};

export type GetS3KeyCredentialsResponse = {
  accessKeyId?: string;
  secretAccessKey?: string;
  s3Endpoint?: string;
  internalS3Endpoint?: string;
  region?: string;
};

export type GetS3KeyCredentials = Record<string, never>;

export class S3API {
  static CreateS3Instance(this:void, req: CreateS3InstanceRequest, initReq?: fm.InitReq): Promise<CreateS3InstanceResponse> {
    return fm.fetchRequest<CreateS3InstanceResponse>(`/api/s3/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListS3Instances(this:void, req: ListS3InstancesRequest, initReq?: fm.InitReq): Promise<ListS3InstancesResponse> {
    return fm.fetchRequest<ListS3InstancesResponse>(`/api/s3/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DropS3Instance(this:void, req: DropS3InstanceRequest, initReq?: fm.InitReq): Promise<DropS3InstanceResponse> {
    return fm.fetchRequest<DropS3InstanceResponse>(`/api/s3/drop`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetS3InstanceCredentials(this:void, req: GetS3InstanceCredentialsRequest, initReq?: fm.InitReq): Promise<GetS3InstanceCredentialsResponse> {
    return fm.fetchRequest<GetS3InstanceCredentialsResponse>(`/api/s3/${req.name}/credentials?${fm.renderURLSearchParams(req, ["name"])}`, {...initReq, method: "GET"});
  }
  static ListS3Buckets(this:void, req: ListS3BucketsRequest, initReq?: fm.InitReq): Promise<ListS3BucketsResponse> {
    return fm.fetchRequest<ListS3BucketsResponse>(`/api/s3/bucket/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static CreateS3Bucket(this:void, req: CreateS3BucketRequest, initReq?: fm.InitReq): Promise<CreateS3BucketResponse> {
    return fm.fetchRequest<CreateS3BucketResponse>(`/api/s3/bucket/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DeleteS3Bucket(this:void, req: DeleteS3BucketRequest, initReq?: fm.InitReq): Promise<DeleteS3BucketResponse> {
    return fm.fetchRequest<DeleteS3BucketResponse>(`/api/s3/bucket/delete`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static SetS3BucketAccess(this:void, req: SetS3BucketAccessRequest, initReq?: fm.InitReq): Promise<SetS3BucketAccessResponse> {
    return fm.fetchRequest<SetS3BucketAccessResponse>(`/api/s3/bucket/access`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListS3Keys(this:void, req: ListS3KeysRequest, initReq?: fm.InitReq): Promise<ListS3KeysResponse> {
    return fm.fetchRequest<ListS3KeysResponse>(`/api/s3/key/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static CreateS3Key(this:void, req: CreateS3KeyRequest, initReq?: fm.InitReq): Promise<CreateS3KeyResponse> {
    return fm.fetchRequest<CreateS3KeyResponse>(`/api/s3/key/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DeleteS3Key(this:void, req: DeleteS3KeyRequest, initReq?: fm.InitReq): Promise<DeleteS3KeyResponse> {
    return fm.fetchRequest<DeleteS3KeyResponse>(`/api/s3/key/delete`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetS3KeyCredentials(this:void, req: GetS3KeyCredentialsRequest, initReq?: fm.InitReq): Promise<GetS3KeyCredentialsResponse> {
    return fm.fetchRequest<GetS3KeyCredentialsResponse>(`/api/s3/${req.instanceName}/key/${req.accessKeyId}/credentials?${fm.renderURLSearchParams(req, ["instanceName", "accessKeyId"])}`, {...initReq, method: "GET"});
  }
}
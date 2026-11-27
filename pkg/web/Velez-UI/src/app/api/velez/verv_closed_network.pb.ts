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

export type CreateVcnNamespaceRequest = {
  name?: string;
};

export type CreateVcnNamespaceResponse = {
  namespace?: Namespace;
};

export type CreateVcnNamespace = Record<string, never>;

export type ListVcnNamespacesRequest = Record<string, never>;

export type ListVcnNamespacesResponse = {
  namespaces?: Namespace[];
};

export type ListVcnNamespaces = Record<string, never>;

export type DeleteVcnNamespaceRequest = {
  id?: string;
};

export type DeleteVcnNamespaceResponse = Record<string, never>;

export type DeleteVcnNamespace = Record<string, never>;

export type ConnectServiceRequest = {
  serviceName?: string;
  domainName?: string;
  authKey?: string;
  loginServerUrl?: string;
  hostname?: string;
  resolveServiceContainer?: boolean;
  networkId?: string;
};

export type ConnectServiceResponse = Record<string, never>;

export type ConnectService = Record<string, never>;

export type Namespace = {
  id?: string;
  name?: string;
};

export type ConnectUserRequest = {
  key?: string;
  username?: string;
};

export type ConnectUserResponse = Record<string, never>;

export type ConnectUser = Record<string, never>;

export type VcnPeer = {
  id?: string;
  host?: string;
  type?: string;
  status?: string;
  latency?: string;
  rx?: string;
  tx?: string;
  ipAddresses?: string[];
};

export type ListPeersRequest = {
  namespaceKey?: string;
  networkId?: string;
};

export type ListPeersResponse = {
  peers?: VcnPeer[];
};

export type ListPeers = Record<string, never>;

export type VcnNetwork = {
  id?: string;
  name?: string;
  provider?: string;
  source?: string;
  url?: string;
  serviceName?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
};

export type NetworkApiKey = {
  prefix?: string;
  createdAt?: GoogleProtobufTimestamp.Timestamp;
  expiration?: GoogleProtobufTimestamp.Timestamp;
  lastSeen?: GoogleProtobufTimestamp.Timestamp;
};

export type CreateVcnNetworkDeploy = {
  customPort?: number;
  customImage?: string;
};

export type CreateVcnNetworkExternal = {
  url?: string;
  token?: string;
};

type BaseCreateVcnNetworkRequest = {
  name?: string;
};

export type CreateVcnNetworkRequest = BaseCreateVcnNetworkRequest &
  OneOf<{
    deploy: CreateVcnNetworkDeploy;
    external: CreateVcnNetworkExternal;
  }>;

export type CreateVcnNetworkResponse = {
  network?: VcnNetwork;
};

export type CreateVcnNetwork = Record<string, never>;

export type ListVcnNetworksRequest = Record<string, never>;

export type ListVcnNetworksResponse = {
  networks?: VcnNetwork[];
};

export type ListVcnNetworks = Record<string, never>;

export type GetVcnNetworkRequest = {
  id?: string;
};

export type GetVcnNetworkResponse = {
  network?: VcnNetwork;
  peers?: VcnPeer[];
  apiKeys?: NetworkApiKey[];
};

export type GetVcnNetwork = Record<string, never>;

export type DropVcnNetworkRequest = {
  id?: string;
};

export type DropVcnNetworkResponse = Record<string, never>;

export type DropVcnNetwork = Record<string, never>;

export type IssueNetworkApiKeyRequest = {
  networkId?: string;
};

export type IssueNetworkApiKeyResponse = {
  key?: string;
  apiKey?: NetworkApiKey;
};

export type IssueNetworkApiKey = Record<string, never>;

export type ExpireNetworkApiKeyRequest = {
  networkId?: string;
  prefix?: string;
};

export type ExpireNetworkApiKeyResponse = Record<string, never>;

export type ExpireNetworkApiKey = Record<string, never>;

export class VcnApi {
  static CreateNamespace(this:void, req: CreateVcnNamespaceRequest, initReq?: fm.InitReq): Promise<CreateVcnNamespaceResponse> {
    return fm.fetchRequest<CreateVcnNamespaceResponse>(`/api/vcn/namespaces/new`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListNamespaces(this:void, req: ListVcnNamespacesRequest, initReq?: fm.InitReq): Promise<ListVcnNamespacesResponse> {
    return fm.fetchRequest<ListVcnNamespacesResponse>(`/api/vcn/namespaces/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ConnectService(this:void, req: ConnectServiceRequest, initReq?: fm.InitReq): Promise<ConnectServiceResponse> {
    return fm.fetchRequest<ConnectServiceResponse>(`/api/vcn/services/connect`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ConnectUser(this:void, req: ConnectUserRequest, initReq?: fm.InitReq): Promise<ConnectUserResponse> {
    return fm.fetchRequest<ConnectUserResponse>(`/api/vcn/users/connect`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DeleteNamespace(this:void, req: DeleteVcnNamespaceRequest, initReq?: fm.InitReq): Promise<DeleteVcnNamespaceResponse> {
    return fm.fetchRequest<DeleteVcnNamespaceResponse>(`/api/vcn/namespaces/delete`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListPeers(this:void, req: ListPeersRequest, initReq?: fm.InitReq): Promise<ListPeersResponse> {
    return fm.fetchRequest<ListPeersResponse>(`/api/vcn/peers/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static CreateNetwork(this:void, req: CreateVcnNetworkRequest, initReq?: fm.InitReq): Promise<CreateVcnNetworkResponse> {
    return fm.fetchRequest<CreateVcnNetworkResponse>(`/api/vcn/networks/new`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ListNetworks(this:void, req: ListVcnNetworksRequest, initReq?: fm.InitReq): Promise<ListVcnNetworksResponse> {
    return fm.fetchRequest<ListVcnNetworksResponse>(`/api/vcn/networks/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetNetwork(this:void, req: GetVcnNetworkRequest, initReq?: fm.InitReq): Promise<GetVcnNetworkResponse> {
    return fm.fetchRequest<GetVcnNetworkResponse>(`/api/vcn/networks/${req.id}?${fm.renderURLSearchParams(req, ["id"])}`, {...initReq, method: "GET"});
  }
  static DropNetwork(this:void, req: DropVcnNetworkRequest, initReq?: fm.InitReq): Promise<DropVcnNetworkResponse> {
    return fm.fetchRequest<DropVcnNetworkResponse>(`/api/vcn/networks/drop`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static IssueNetworkApiKey(this:void, req: IssueNetworkApiKeyRequest, initReq?: fm.InitReq): Promise<IssueNetworkApiKeyResponse> {
    return fm.fetchRequest<IssueNetworkApiKeyResponse>(`/api/vcn/networks/api_keys/issue`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ExpireNetworkApiKey(this:void, req: ExpireNetworkApiKeyRequest, initReq?: fm.InitReq): Promise<ExpireNetworkApiKeyResponse> {
    return fm.fetchRequest<ExpireNetworkApiKeyResponse>(`/api/vcn/networks/api_keys/expire`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
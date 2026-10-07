/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";


export enum NetworkProvider {
  NETWORK_PROVIDER_UNSPECIFIED = "NETWORK_PROVIDER_UNSPECIFIED",
  NETWORK_PROVIDER_DOCKER = "NETWORK_PROVIDER_DOCKER",
  NETWORK_PROVIDER_VCN = "NETWORK_PROVIDER_VCN",
}

export enum NetworkCapability {
  NETWORK_CAPABILITY_UNSPECIFIED = "NETWORK_CAPABILITY_UNSPECIFIED",
  NETWORK_CAPABILITY_CREATE = "NETWORK_CAPABILITY_CREATE",
  NETWORK_CAPABILITY_DELETE = "NETWORK_CAPABILITY_DELETE",
  NETWORK_CAPABILITY_ATTACH = "NETWORK_CAPABILITY_ATTACH",
  NETWORK_CAPABILITY_RESTRICT = "NETWORK_CAPABILITY_RESTRICT",
}

export type NetworkMember = {
  containerId?: string;
  containerName?: string;
  aliases?: string[];
  ipAddress?: string;
};

export type Network = {
  id?: string;
  name?: string;
  provider?: NetworkProvider;
  isManaged?: boolean;
  isInternal?: boolean;
  isIccEnabled?: boolean;
  subnet?: string;
  members?: NetworkMember[];
  capabilities?: NetworkCapability[];
};

export type NetworkProviderInfo = {
  provider?: NetworkProvider;
  capabilities?: NetworkCapability[];
};

export type GetNetworkStatusRequest = Record<string, never>;

export type GetNetworkStatusResponse = {
  isClusterMode?: boolean;
  isVcnConnected?: boolean;
  providers?: NetworkProviderInfo[];
};

export type GetNetworkStatus = Record<string, never>;

export type ListNetworksRequest = {
  environment?: string;
  isForeignIncluded?: boolean;
};

export type ListNetworksResponse = {
  networks?: Network[];
};

export type ListNetworks = Record<string, never>;

export type GetNetworkRequest = {
  id?: string;
  environment?: string;
};

export type GetNetworkResponse = {
  network?: Network;
};

export type GetNetwork = Record<string, never>;

export type CreateNetworkRequest = {
  name?: string;
  environment?: string;
  isInternal?: boolean;
  isIccEnabled?: boolean;
};

export type CreateNetworkResponse = {
  network?: Network;
};

export type CreateNetwork = Record<string, never>;

export type DeleteNetworkRequest = {
  id?: string;
  environment?: string;
};

export type DeleteNetworkResponse = Record<string, never>;

export type DeleteNetwork = Record<string, never>;

export type ConnectContainerRequest = {
  networkId?: string;
  containerName?: string;
  aliases?: string[];
  environment?: string;
};

export type ConnectContainerResponse = Record<string, never>;

export type ConnectContainer = Record<string, never>;

export type DisconnectContainerRequest = {
  networkId?: string;
  containerName?: string;
  environment?: string;
};

export type DisconnectContainerResponse = Record<string, never>;

export type DisconnectContainer = Record<string, never>;

export class NetworkAPI {
  static GetNetworkStatus(this:void, req: GetNetworkStatusRequest, initReq?: fm.InitReq): Promise<GetNetworkStatusResponse> {
    return fm.fetchRequest<GetNetworkStatusResponse>(`/api/network/status?${fm.renderURLSearchParams(req, [])}`, {...initReq, method: "GET"});
  }
  static ListNetworks(this:void, req: ListNetworksRequest, initReq?: fm.InitReq): Promise<ListNetworksResponse> {
    return fm.fetchRequest<ListNetworksResponse>(`/api/network/list`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static GetNetwork(this:void, req: GetNetworkRequest, initReq?: fm.InitReq): Promise<GetNetworkResponse> {
    return fm.fetchRequest<GetNetworkResponse>(`/api/network/details/${req.id}?${fm.renderURLSearchParams(req, ["id"])}`, {...initReq, method: "GET"});
  }
  static CreateNetwork(this:void, req: CreateNetworkRequest, initReq?: fm.InitReq): Promise<CreateNetworkResponse> {
    return fm.fetchRequest<CreateNetworkResponse>(`/api/network/create`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DeleteNetwork(this:void, req: DeleteNetworkRequest, initReq?: fm.InitReq): Promise<DeleteNetworkResponse> {
    return fm.fetchRequest<DeleteNetworkResponse>(`/api/network/delete`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static ConnectContainer(this:void, req: ConnectContainerRequest, initReq?: fm.InitReq): Promise<ConnectContainerResponse> {
    return fm.fetchRequest<ConnectContainerResponse>(`/api/network/connect`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
  static DisconnectContainer(this:void, req: DisconnectContainerRequest, initReq?: fm.InitReq): Promise<DisconnectContainerResponse> {
    return fm.fetchRequest<DisconnectContainerResponse>(`/api/network/disconnect`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)});
  }
}
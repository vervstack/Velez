/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";


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
}
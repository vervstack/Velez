/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";


export type CreateVpnNamespaceRequest = {
    name?: string;
};

export type CreateVpnNamespaceResponse = {
    namespace?: Namespace;
};

export type CreateVpnNamespace = Record<string, never>;

export type ListVpnNamespacesRequest = Record<string, never>;

export type ListVpnNamespacesResponse = {
    namespaces?: Namespace[];
};

export type ListVpnNamespaces = Record<string, never>;

export type DeleteVpnNamespaceRequest = {
    id?: string;
};

export type DeleteVpnNamespaceResponse = Record<string, never>;

export type DeleteVpnNamespace = Record<string, never>;

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

export class VpnApi {
    static CreateNamespace(this: void, req: CreateVpnNamespaceRequest, initReq?: fm.InitReq): Promise<CreateVpnNamespaceResponse> {
        return fm.fetchRequest<CreateVpnNamespaceResponse>(`/api/vpn/namespaces/new`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
    }

    static ListNamespaces(this: void, req: ListVpnNamespacesRequest, initReq?: fm.InitReq): Promise<ListVpnNamespacesResponse> {
        return fm.fetchRequest<ListVpnNamespacesResponse>(`/api/vpn/namespaces/list`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
    }

    static ConnectService(this: void, req: ConnectServiceRequest, initReq?: fm.InitReq): Promise<ConnectServiceResponse> {
        return fm.fetchRequest<ConnectServiceResponse>(`/api/vpn/services/connect`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
    }

    static ConnectUser(this: void, req: ConnectUserRequest, initReq?: fm.InitReq): Promise<ConnectUserResponse> {
        return fm.fetchRequest<ConnectUserResponse>(`/api/vpn/users/connect`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
    }

    static DeleteNamespace(this: void, req: DeleteVpnNamespaceRequest, initReq?: fm.InitReq): Promise<DeleteVpnNamespaceResponse> {
        return fm.fetchRequest<DeleteVpnNamespaceResponse>(`/api/vpn/namespaces/delete`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
    }
}
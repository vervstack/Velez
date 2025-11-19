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

export type Namespace = {
    id?: string;
    name?: string;
};

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
}
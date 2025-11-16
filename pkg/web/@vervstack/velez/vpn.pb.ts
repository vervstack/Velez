/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";


export type ListVpnNamespacesRequest = Record<string, never>;

export type ListVpnNamespacesResponse = Record<string, never>;

export type ListVpnNamespaces = Record<string, never>;

export class VpnApi {
    static ListNamespaces(this: void, req: ListVpnNamespacesRequest, initReq?: fm.InitReq): Promise<ListVpnNamespacesResponse> {
        return fm.fetchRequest<ListVpnNamespacesResponse>(`/api/vpn/namespaces`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
    }
}
/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";


export type CreateServiceRequest = {
    name?: string;
};

export type CreateServiceResponse = Record<string, never>;

export type CreateService = Record<string, never>;

export class ServiceApi {
    static CreateService(this: void, req: CreateServiceRequest, initReq?: fm.InitReq): Promise<CreateServiceResponse> {
        return fm.fetchRequest<CreateServiceResponse>(`/api/service/create`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
    }
}
/* eslint-disable */
// @ts-nocheck

/**
 * This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
 */

import * as fm from "./fetch.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };

type OneOf<T> =
    | { [k in keyof T]?: undefined }
    | (keyof T extends infer K
    ? K extends string & keyof T
        ? { [k in K]: T[K] } & Absent<T, K>
        : never
    : never);

export type CreateServiceRequest = {
    name?: string;
};

export type CreateServiceResponse = Record<string, never>;

export type CreateService = Record<string, never>;

type BaseGetServiceRequest = {};

export type GetServiceRequest = BaseGetServiceRequest &
    OneOf<{
        id: string;
        name: string;
    }>;

type BaseGetServiceResponse = {};

export type GetServiceResponse = BaseGetServiceResponse &
    OneOf<{
        vervService: VervAppService;
    }>;

export type GetService = Record<string, never>;

export type VervAppService = {
    id?: string;
    name?: string;
};

export class ServiceApi {
    static CreateService(this: void, req: CreateServiceRequest, initReq?: fm.InitReq): Promise<CreateServiceResponse> {
        return fm.fetchRequest<CreateServiceResponse>(`/api/service/create`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
    }

    static GetService(this: void, req: GetServiceRequest, initReq?: fm.InitReq): Promise<GetServiceResponse> {
        return fm.fetchRequest<GetServiceResponse>(`/api/service/get`, {
            ...initReq,
            method: "POST",
            body: JSON.stringify(req, fm.replacer)
        });
    }
}
import {HaproxyAPI} from "@/app/api/velez/haproxy_api.pb"
import type {
    BootstrapVcnSidecarRequest,
    CreateHaproxyInstanceRequest,
    DropHaproxyInstanceRequest,
    SelectHaproxyMasterRequest,
    ValidatePgBackendRequest,
} from "@/app/api/velez/haproxy_api.pb"
import {PortProtocol} from "@/app/api/velez/velez_common.pb"
import {ApiService} from "@/processes/ApiService.ts"

export {PgBackendRole} from "@/app/api/velez/haproxy_api.pb"

class HaproxyService extends ApiService {
    listHaproxyInstances() {
        return this.execute((initReq) => HaproxyAPI.ListHaproxyInstances({}, initReq))
    }

    createHaproxyInstance(req: CreateHaproxyInstanceRequest) {
        return this.mutate((initReq) => HaproxyAPI.CreateHaproxyInstance(req, initReq))
    }

    bootstrapVcnSidecar(req: BootstrapVcnSidecarRequest, publishedPort?: number) {
        const ports = publishedPort === undefined ? [] : [{
            servicePortNumber: publishedPort,
            exposedTo: publishedPort,
            protocol: PortProtocol.tcp,
        }]
        return this.mutate((initReq) => HaproxyAPI.BootstrapVcnSidecar({...req, ports}, initReq))
    }

    validatePgBackend(req: ValidatePgBackendRequest) {
        return this.mutate((initReq) => HaproxyAPI.ValidatePgBackend(req, initReq), 95_000)
    }

    dropHaproxyInstance(name: string) {
        const req: DropHaproxyInstanceRequest = {name}
        return this.mutate((initReq) => HaproxyAPI.DropHaproxyInstance(req, initReq))
    }

    selectHaproxyMaster(req: SelectHaproxyMasterRequest) {
        return this.mutate((initReq) => HaproxyAPI.SelectHaproxyMaster(req, initReq))
    }
}

export const haproxyService = new HaproxyService()

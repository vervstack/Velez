import {SettingsAPI} from "@/app/api/velez/settings_api.pb"
import type {
    GetSettingsResponse,
    GetSysboxStatusResponse,
    RebuildAddressesResponse,
    RunSysboxSmokeTestResponse,
    UpdateSettingsRequest,
    UpdateSettingsResponse,
} from "@/app/api/velez/settings_api.pb"
import type {AddressesRebuildStatus} from "@/model/settings/AddressesRebuildStatus.ts"
import {ApiService} from "@/processes/ApiService.ts"

const SMOKE_TEST_TIMEOUT_MS = 120000

class SettingsService extends ApiService {
    async getSettings(): Promise<GetSettingsResponse> {
        return this.execute((initReq) => SettingsAPI.GetSettings({}, initReq))
    }

    async updateSettings(req: UpdateSettingsRequest): Promise<UpdateSettingsResponse> {
        return this.mutate((initReq) => SettingsAPI.UpdateSettings(req, initReq))
    }

    async rebuildAddresses(): Promise<RebuildAddressesResponse> {
        return this.mutate((initReq) => SettingsAPI.RebuildAddresses({}, initReq))
    }

    async getAddressesRebuildStatus(): Promise<AddressesRebuildStatus> {
        return this.execute((initReq) => SettingsAPI.GetAddressesRebuildStatus({}, initReq))
            .then((response) => ({
                isRunning: response.isRunning ?? false,
                totalSteps: response.totalSteps ?? 0,
                doneSteps: response.doneSteps ?? 0,
                lastError: response.lastError ?? "",
            }))
    }

    async getSysboxStatus(): Promise<GetSysboxStatusResponse> {
        return this.execute((initReq) => SettingsAPI.GetSysboxStatus({}, initReq))
    }

    async runSysboxSmokeTest(): Promise<RunSysboxSmokeTestResponse> {
        return this.mutate((initReq) => SettingsAPI.RunSysboxSmokeTest({}, initReq), SMOKE_TEST_TIMEOUT_MS)
    }
}

export const settingsService = new SettingsService()

import {TransferAPI} from "@/app/api/velez/transfer_api.pb"
import type {
    CreateTransferTokenResponse,
    ListTransfersResponse,
    PlanTransferResponse,
    StartTransferRequest,
    StartTransferResponse,
} from "@/app/api/velez/transfer_api.pb"
import {ApiService} from "@/processes/ApiService.ts"

class TransferService extends ApiService {
    async listTransfers(): Promise<ListTransfersResponse> {
        return this.execute((initReq) => TransferAPI.ListTransfers({}, initReq))
    }

    async createTransferToken(): Promise<CreateTransferTokenResponse> {
        return this.mutate((initReq) => TransferAPI.CreateTransferToken({}, initReq))
    }

    async planTransfer(containerName: string): Promise<PlanTransferResponse> {
        return this.mutate((initReq) => TransferAPI.PlanTransfer({containerName}, initReq))
    }

    async startTransfer(req: StartTransferRequest): Promise<StartTransferResponse> {
        return this.mutate((initReq) => TransferAPI.StartTransfer(req, initReq))
    }

    async cancelTransfer(transferId: string): Promise<void> {
        return this.mutate((initReq) => TransferAPI.CancelTransfer({transferId}, initReq).then())
    }
}

export const transferService = new TransferService()

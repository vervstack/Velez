import {useCredentialsStore} from '@/app/settings/creds.ts'
import {parseGrpcError, ServiceError} from '@vervstack/chures'
import type {InitReq} from '@/app/settings/state.ts'

export class ApiService {
    protected async execute<T>(fn: (req: InitReq) => Promise<T>): Promise<T> {
        return withRetries(() => this.call(fn), 2)
    }

    protected async mutate<T>(fn: (req: InitReq) => Promise<T>): Promise<T> {
        return this.call(fn)
    }

    private async call<T>(fn: (req: InitReq) => Promise<T>): Promise<T> {
        const initReq = useCredentialsStore.getState().getInitReq()
        try {
            return await fn(initReq)
        } catch (e) {
            throw parseGrpcError(e)
        }
    }
}

function withRetries<T>(fn: () => Promise<T>, retries: number): Promise<T> {
    return fn().catch((err) => {
        if (err instanceof ServiceError && !err.isRetryable) throw err
        if (retries > 0) return withRetries(fn, retries - 1)
        throw err
    })
}
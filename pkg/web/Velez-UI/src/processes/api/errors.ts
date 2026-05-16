export enum Errors {
    OK = 0,
    CANCELLED = 1,
    UNKNOWN = 2,
    INVALID_ARGUMENT = 3,
    DEADLINE_EXCEEDED = 4,
    NOT_FOUND = 5,
    ALREADY_EXISTS = 6,
    PERMISSION_DENIED = 7,
    RESOURCE_EXHAUSTED = 8,
    FAILED_PRECONDITION = 9,
    ABORTED = 10,
    OUT_OF_RANGE = 11,
    UNIMPLEMENTED = 12,
    INTERNAL = 13,
    UNAVAILABLE = 14,
    DATA_LOSS = 15,
    UNAUTHENTICATED = 16
}

export type GrpcError = {
    message: string
    code: number
    details?: unknown[]
}

export class ServiceError extends Error {
    readonly isRetryable: boolean
    readonly grpcCode?: number

    constructor(opts: { title: string; isRetryable?: boolean; grpcCode?: number }) {
        super(opts.title)
        this.isRetryable = opts.isRetryable ?? false
        this.grpcCode = opts.grpcCode
    }
}

export function parseGrpcError(e: unknown): ServiceError {
    if (e instanceof ServiceError) return e

    if (e instanceof Error && e.message === 'Failed to fetch') {
        return new ServiceError({ title: 'Server unavailable. Try again later', isRetryable: true })
    }

    if (typeof e === 'object' && e !== null && 'code' in e) {
        const grpc = e as GrpcError
        const retryable = grpc.code === Errors.UNAVAILABLE || grpc.code === Errors.INTERNAL
        return new ServiceError({ title: grpc.message, isRetryable: retryable, grpcCode: grpc.code })
    }

    return new ServiceError({ title: String(e), isRetryable: false })
}

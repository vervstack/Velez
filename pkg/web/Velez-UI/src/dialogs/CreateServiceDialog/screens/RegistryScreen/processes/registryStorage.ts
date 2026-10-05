import type {RegistryS3Storage} from "@/app/api/velez"

export type RegistryStorageKind = "local" | "s3"

export interface RegistryStorageState {
    kind: RegistryStorageKind
    s3Instance: string
    s3Bucket: string
}

export const DEFAULT_REGISTRY_STORAGE: RegistryStorageState = {kind: "local", s3Instance: "", s3Bucket: ""}

export function isRegistryStorageValid(storage: RegistryStorageState): boolean {
    return storage.kind === "local" || storage.s3Instance !== ""
}

export function buildRegistryS3Storage(storage: RegistryStorageState): RegistryS3Storage | undefined {
    if (storage.kind !== "s3") return undefined

    return {
        instanceName: storage.s3Instance,
        bucketName: storage.s3Bucket.trim() || undefined,
    }
}

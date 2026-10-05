import type {CreateS3KeyRequest} from "@/app/api/velez/s3_api.pb"

export interface KeyPermissionRow {
    id: number
    bucketName: string
    isRead: boolean
    isWrite: boolean
    isOwner: boolean
}

export function newKeyPermissionRow(id: number): KeyPermissionRow {
    return {id, bucketName: "", isRead: true, isWrite: false, isOwner: false}
}

export function buildCreateS3KeyRequest(
    instanceName: string,
    keyName: string,
    rows: KeyPermissionRow[]
): CreateS3KeyRequest | null {
    const trimmedName = keyName.trim()
    if (!trimmedName) return null

    const access = rows
        .filter((row) => row.bucketName)
        .map((row) => ({
            bucketName: row.bucketName,
            isRead: row.isRead,
            isWrite: row.isWrite,
            isOwner: row.isOwner,
        }))

    return {instanceName, keyName: trimmedName, access}
}

import type {CreateS3BucketRequest} from "@/app/api/velez/s3_api.pb"

export function buildCreateS3BucketRequest(
    instanceName: string,
    bucketName: string,
    ownerService: string
): CreateS3BucketRequest | null {
    const trimmedName = bucketName.trim()
    if (!trimmedName) return null

    return {
        instanceName,
        bucketName: trimmedName,
        ownerService: ownerService || undefined,
    }
}

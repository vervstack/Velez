import {describe, expect, it} from "vitest"

import {
    buildRegistryS3Storage,
    DEFAULT_REGISTRY_STORAGE,
    isRegistryStorageValid,
} from "@/dialogs/CreateServiceDialog/screens/RegistryScreen/processes/registryStorage.ts"

describe("buildRegistryS3Storage", () => {
    it("returns undefined for local storage, even if S3 fields were typed before switching back", () => {
        expect(buildRegistryS3Storage({kind: "local", s3Instance: "s3", s3Bucket: "b"})).toBeUndefined()
    })

    it("sends the instance and omits an empty bucket so the backend defaults it", () => {
        expect(buildRegistryS3Storage({kind: "s3", s3Instance: "s3", s3Bucket: "  "})).toEqual({
            instanceName: "s3",
            bucketName: undefined,
        })
    })

    it("sends the trimmed bucket when one is typed", () => {
        expect(buildRegistryS3Storage({kind: "s3", s3Instance: "s3", s3Bucket: " images "})?.bucketName)
            .toBe("images")
    })
})

describe("isRegistryStorageValid", () => {
    it("accepts local storage and rejects S3 without an instance", () => {
        expect(isRegistryStorageValid(DEFAULT_REGISTRY_STORAGE)).toBe(true)
        expect(isRegistryStorageValid({kind: "s3", s3Instance: "", s3Bucket: ""})).toBe(false)
        expect(isRegistryStorageValid({kind: "s3", s3Instance: "s3", s3Bucket: ""})).toBe(true)
    })
})

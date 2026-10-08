import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import type {
    CreateS3BucketRequest,
    CreateS3InstanceRequest,
    CreateS3KeyRequest,
    DeleteS3BucketRequest,
    DeleteS3KeyRequest,
    SetS3BucketAccessRequest,
} from "@/app/api/velez/s3_api.pb"
import {s3Service} from "@/processes/api/s3aas.ts"
import {provisioningRefetchInterval} from "@/processes/mappings/provisioning.ts"

export const S3_INSTANCES_QUERY_KEY = ["s3-instances"]
const LIST_REQ = {paging: {limit: "50", offset: "0"}}

function bucketsKey(instanceName: string) {
    return ["s3-buckets", instanceName]
}

function keysKey(instanceName: string) {
    return ["s3-keys", instanceName]
}

export function useListS3InstancesQuery() {
    return useQuery({
        queryKey: S3_INSTANCES_QUERY_KEY,
        queryFn: () => s3Service.listS3Instances(LIST_REQ),
        refetchInterval: (query) => provisioningRefetchInterval(query.state.data?.provisioning),
    })
}

// The list is refetched by the progress screen once the create_s3_instance task reaches DONE.
export function CreateS3InstanceMutation() {
    return useMutation({
        mutationFn: (req: CreateS3InstanceRequest) => s3Service.createS3Instance(req),
    })
}

export function DropS3InstanceMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (name: string) => s3Service.dropS3Instance(name),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: S3_INSTANCES_QUERY_KEY})
            queryClient.invalidateQueries({queryKey: ["services"]})
        },
    })
}

// Disabled by default — the admin token and web UI password are only resolved when the caller
// explicitly triggers refetch().
export function GetS3InstanceCredentialsQuery(name: string) {
    return useQuery({
        queryKey: ["s3-instance-credentials", name],
        queryFn: () => s3Service.getS3InstanceCredentials(name),
        enabled: false,
    })
}

export function useListS3BucketsQuery(instanceName: string) {
    return useQuery({
        queryKey: bucketsKey(instanceName),
        queryFn: () => s3Service.listS3Buckets(instanceName),
    })
}

export function CreateS3BucketMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: CreateS3BucketRequest) => s3Service.createS3Bucket(req),
        onSuccess: (_data, req) => {
            queryClient.invalidateQueries({queryKey: bucketsKey(req.instanceName ?? "")})
            queryClient.invalidateQueries({queryKey: keysKey(req.instanceName ?? "")})
        },
    })
}

export function DeleteS3BucketMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: DeleteS3BucketRequest) => s3Service.deleteS3Bucket(req),
        onSuccess: (_data, req) => {
            queryClient.invalidateQueries({queryKey: bucketsKey(req.instanceName ?? "")})
            queryClient.invalidateQueries({queryKey: keysKey(req.instanceName ?? "")})
        },
    })
}

export function SetS3BucketAccessMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: SetS3BucketAccessRequest) => s3Service.setS3BucketAccess(req),
        onSuccess: (_data, req) => {
            queryClient.invalidateQueries({queryKey: bucketsKey(req.instanceName ?? "")})
            queryClient.invalidateQueries({queryKey: keysKey(req.instanceName ?? "")})
        },
    })
}

export function useListS3KeysQuery(instanceName: string) {
    return useQuery({
        queryKey: keysKey(instanceName),
        queryFn: () => s3Service.listS3Keys(instanceName),
    })
}

export function CreateS3KeyMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: CreateS3KeyRequest) => s3Service.createS3Key(req),
        onSuccess: (_data, req) => {
            queryClient.invalidateQueries({queryKey: keysKey(req.instanceName ?? "")})
            queryClient.invalidateQueries({queryKey: bucketsKey(req.instanceName ?? "")})
        },
    })
}

export function DeleteS3KeyMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: DeleteS3KeyRequest) => s3Service.deleteS3Key(req),
        onSuccess: (_data, req) => {
            queryClient.invalidateQueries({queryKey: keysKey(req.instanceName ?? "")})
            queryClient.invalidateQueries({queryKey: bucketsKey(req.instanceName ?? "")})
        },
    })
}

// Disabled by default — the secret is only resolved on an explicit Reveal / Copy.
export function GetS3KeyCredentialsQuery(instanceName: string, accessKeyId: string) {
    return useQuery({
        queryKey: ["s3-key-credentials", instanceName, accessKeyId],
        queryFn: () => s3Service.getS3KeyCredentials(instanceName, accessKeyId),
        enabled: false,
    })
}

import {S3API} from "@/app/api/velez/s3_api.pb"
import type {
    CreateS3BucketRequest,
    CreateS3BucketResponse,
    CreateS3InstanceRequest,
    CreateS3InstanceResponse,
    CreateS3KeyRequest,
    CreateS3KeyResponse,
    DeleteS3BucketRequest,
    DeleteS3KeyRequest,
    DropS3InstanceRequest,
    GetS3InstanceCredentialsRequest,
    GetS3InstanceCredentialsResponse,
    GetS3KeyCredentialsRequest,
    GetS3KeyCredentialsResponse,
    ListS3BucketsRequest,
    ListS3BucketsResponse,
    ListS3InstancesRequest,
    ListS3InstancesResponse,
    ListS3KeysRequest,
    ListS3KeysResponse,
    SetS3BucketAccessRequest,
    SetS3BucketAccessResponse,
} from "@/app/api/velez/s3_api.pb"
import {ApiService} from "@/processes/ApiService.ts"

class S3Service extends ApiService {
    async listS3Instances(req: ListS3InstancesRequest): Promise<ListS3InstancesResponse> {
        return this.execute((initReq) => S3API.ListS3Instances(req, initReq))
    }

    async createS3Instance(req: CreateS3InstanceRequest): Promise<CreateS3InstanceResponse> {
        return this.mutate((initReq) => S3API.CreateS3Instance(req, initReq))
    }

    async dropS3Instance(name: string): Promise<void> {
        return this.mutate((initReq) => {
            const payload: DropS3InstanceRequest = {name}
            return S3API.DropS3Instance(payload, initReq).then()
        })
    }

    async getS3InstanceCredentials(name: string): Promise<GetS3InstanceCredentialsResponse> {
        return this.execute((initReq) => {
            const payload: GetS3InstanceCredentialsRequest = {name}
            return S3API.GetS3InstanceCredentials(payload, initReq)
        })
    }

    async listS3Buckets(instanceName: string): Promise<ListS3BucketsResponse> {
        return this.execute((initReq) => {
            const payload: ListS3BucketsRequest = {instanceName}
            return S3API.ListS3Buckets(payload, initReq)
        })
    }

    async createS3Bucket(req: CreateS3BucketRequest): Promise<CreateS3BucketResponse> {
        return this.mutate((initReq) => S3API.CreateS3Bucket(req, initReq))
    }

    async deleteS3Bucket(req: DeleteS3BucketRequest): Promise<void> {
        return this.mutate((initReq) => S3API.DeleteS3Bucket(req, initReq).then())
    }

    async setS3BucketAccess(req: SetS3BucketAccessRequest): Promise<SetS3BucketAccessResponse> {
        return this.mutate((initReq) => S3API.SetS3BucketAccess(req, initReq))
    }

    async listS3Keys(instanceName: string): Promise<ListS3KeysResponse> {
        return this.execute((initReq) => {
            const payload: ListS3KeysRequest = {instanceName}
            return S3API.ListS3Keys(payload, initReq)
        })
    }

    async createS3Key(req: CreateS3KeyRequest): Promise<CreateS3KeyResponse> {
        return this.mutate((initReq) => S3API.CreateS3Key(req, initReq))
    }

    async deleteS3Key(req: DeleteS3KeyRequest): Promise<void> {
        return this.mutate((initReq) => S3API.DeleteS3Key(req, initReq).then())
    }

    async getS3KeyCredentials(instanceName: string, accessKeyId: string): Promise<GetS3KeyCredentialsResponse> {
        return this.execute((initReq) => {
            const payload: GetS3KeyCredentialsRequest = {instanceName, accessKeyId}
            return S3API.GetS3KeyCredentials(payload, initReq)
        })
    }
}

export const s3Service = new S3Service()

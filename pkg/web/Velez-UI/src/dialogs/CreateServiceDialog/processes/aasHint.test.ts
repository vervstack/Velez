import {describe, expect, it} from "vitest"

import {aasHintOf} from "@/dialogs/CreateServiceDialog/processes/aasHint.ts"

describe("aasHintOf", () => {
    it("returns the postgres hint when the image is postgres:18", () => {
        expect(aasHintOf("postgres:18")).toEqual({screen: "postgres", title: "Postgres"})
    })

    it("returns the registry hint when the image is registry:2", () => {
        expect(aasHintOf("registry:2")).toEqual({screen: "registry", title: "Container registry"})
    })

    it("returns the s3 hint when the image is dxflrs/garage:v2.4.0", () => {
        expect(aasHintOf("dxflrs/garage:v2.4.0")).toEqual({screen: "s3", title: "S3 storage"})
    })

    it("returns the gitlab runner hint when the image is gitlab/gitlab-runner:latest", () => {
        expect(aasHintOf("gitlab/gitlab-runner:latest")).toEqual({screen: "gitlabRunner", title: "GitLab runner"})
    })

    it("returns the github runner hint when the image is ghcr.io/actions/actions-runner:latest", () => {
        expect(aasHintOf("ghcr.io/actions/actions-runner:latest")).toEqual({
            screen: "githubRunner",
            title: "GitHub runner",
        })
    })

    it("returns the dind hint when the image is docker:dind", () => {
        expect(aasHintOf("docker:dind")).toEqual({screen: "dind", title: "Docker-in-Docker"})
    })

    it("returns the dind hint only when the docker tag starts with dind", () => {
        expect(aasHintOf("docker:27-dind")).toBeUndefined()
        expect(aasHintOf("docker:dind-rootless")).toEqual({screen: "dind", title: "Docker-in-Docker"})
    })

    it("returns undefined when the docker image has no dind tag", () => {
        expect(aasHintOf("docker")).toBeUndefined()
        expect(aasHintOf("docker:latest")).toBeUndefined()
    })

    it("returns the hint when the image is a bare repository name", () => {
        expect(aasHintOf("postgres")).toEqual({screen: "postgres", title: "Postgres"})
    })

    it("strips host, port and tag when the image has a private registry prefix", () => {
        expect(aasHintOf("my.registry:5000/team/postgres:16")).toEqual({screen: "postgres", title: "Postgres"})
    })

    it("strips the digest when the image is pinned by digest", () => {
        expect(aasHintOf("postgres:16@sha256:abc123")).toEqual({screen: "postgres", title: "Postgres"})
        expect(aasHintOf("registry@sha256:abc123")).toEqual({screen: "registry", title: "Container registry"})
    })

    it("ignores case and surrounding whitespace", () => {
        expect(aasHintOf("  Postgres:16 ")).toEqual({screen: "postgres", title: "Postgres"})
    })

    it("returns undefined when the image is unknown", () => {
        expect(aasHintOf("nginx:latest")).toBeUndefined()
        expect(aasHintOf("my.registry:5000/team/postgres-exporter:1")).toBeUndefined()
    })

    it("returns undefined when the image is empty", () => {
        expect(aasHintOf("")).toBeUndefined()
        expect(aasHintOf("   ")).toBeUndefined()
    })

    it("returns undefined when the repository name is an object prototype key", () => {
        expect(aasHintOf("constructor")).toBeUndefined()
    })
})

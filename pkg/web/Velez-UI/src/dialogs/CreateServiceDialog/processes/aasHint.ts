import {ProductScreen} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

export interface AasHint {
    screen: ProductScreen
    title: string
}

const HINT_BY_REPOSITORY = new Map<string, AasHint>([
    ["postgres", {screen: "postgres", title: "Postgres"}],
    ["registry", {screen: "registry", title: "Container registry"}],
    ["garage", {screen: "s3", title: "S3 storage"}],
    ["gitlab-runner", {screen: "gitlabRunner", title: "GitLab runner"}],
    ["actions-runner", {screen: "githubRunner", title: "GitHub runner"}],
])

const DIND_HINT: AasHint = {screen: "dind", title: "Docker-in-Docker"}

export function aasHintOf(image: string): AasHint | undefined {
    const withoutDigest = image.trim().toLowerCase().split("@")[0]
    const lastSlash = withoutDigest.lastIndexOf("/")
    const lastColon = withoutDigest.lastIndexOf(":")
    const hasTag = lastColon > lastSlash
    const repository = withoutDigest.slice(lastSlash + 1, hasTag ? lastColon : undefined)
    const tag = hasTag ? withoutDigest.slice(lastColon + 1) : ""

    if (repository === "docker") {
        return tag.startsWith("dind") ? DIND_HINT : undefined
    }

    return HINT_BY_REPOSITORY.get(repository)
}

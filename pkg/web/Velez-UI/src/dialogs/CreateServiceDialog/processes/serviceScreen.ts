import {ComponentType} from "react"

import {RunnerProvider, ServicePattern} from "@/app/api/velez"
import DindIcon from "@/components/base/icons/DindIcon.tsx"
import GithubIcon from "@/components/base/icons/GithubIcon.tsx"
import GitlabIcon from "@/components/base/icons/GitlabIcon.tsx"
import PostgresIcon from "@/components/base/icons/PostgresIcon.tsx"
import RegistryIcon from "@/components/base/icons/RegistryIcon.tsx"
import S3Icon from "@/components/base/icons/S3Icon.tsx"
import ServiceIcon from "@/components/base/icons/ServiceIcon.tsx"

export type ServiceScreen = "picker" | "generic" | "postgres" | "registry" | "s3" | "dind" | "githubRunner" | "gitlabRunner"

export type ProductScreen = Exclude<ServiceScreen, "picker">

export interface ProductCard {
    screen: ProductScreen
    title: string
    description: string
    icon: ComponentType<{className?: string}>
}

export const PRODUCT_CARDS: ProductCard[] = [
    {
        screen: "generic",
        title: "Generic container/image",
        description: "Run any Docker image, optionally from Git",
        icon: ServiceIcon,
    },
    {screen: "postgres", title: "PostgreSQL", description: "A managed PostgreSQL database instance", icon: PostgresIcon},
    {
        screen: "registry",
        title: "Container registry",
        description: "A private registry to store your images",
        icon: RegistryIcon,
    },
    {screen: "s3", title: "S3 storage", description: "An S3-compatible object storage instance", icon: S3Icon},
    {
        screen: "dind",
        title: "Docker in Docker",
        description: "An isolated Docker daemon for builds and runners",
        icon: DindIcon,
    },
    {screen: "githubRunner", title: "GitHub runner", description: "A self-hosted GitHub Actions runner", icon: GithubIcon},
    {screen: "gitlabRunner", title: "GitLab runner", description: "A self-hosted GitLab CI runner", icon: GitlabIcon},
]

const SCREEN_TITLES: Record<ServiceScreen, string> = {
    picker: "Create service",
    generic: "Create app",
    postgres: "Create database",
    registry: "Create registry",
    s3: "Create S3 instance",
    dind: "Create Docker daemon",
    githubRunner: "Create runner",
    gitlabRunner: "Create runner",
}

export function screenTitle(screen: ServiceScreen): string {
    return SCREEN_TITLES[screen]
}

export function runnerProviderOf(screen: ServiceScreen): RunnerProvider | undefined {
    if (screen === "githubRunner") return RunnerProvider.GITHUB
    if (screen === "gitlabRunner") return RunnerProvider.GITLAB
    return undefined
}

export const ADOPTABLE_SCREENS: ProductScreen[] = ["generic", "postgres", "registry", "githubRunner", "gitlabRunner"]

const SUGGESTED_SCREEN_BY_PATTERN: Partial<Record<ServicePattern, ServiceScreen>> = {
    [ServicePattern.SERVICE_PATTERN_POSTGRES]: "postgres",
    [ServicePattern.SERVICE_PATTERN_REGISTRY]: "registry",
    [ServicePattern.SERVICE_PATTERN_GITHUB_RUNNER]: "githubRunner",
    [ServicePattern.SERVICE_PATTERN_GITLAB_RUNNER]: "gitlabRunner",
}

export function suggestedScreenOf(pattern?: ServicePattern): ServiceScreen | undefined {
    if (!pattern) return undefined
    return SUGGESTED_SCREEN_BY_PATTERN[pattern]
}

export function adoptTitle(containerName?: string): string {
    return `Onboard container ${containerName ?? ""}`.trim()
}

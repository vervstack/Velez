import {RunnerProvider, ServicePattern} from "@/app/api/velez"

export type ServiceScreen = "picker" | "generic" | "postgres" | "registry" | "githubRunner" | "gitlabRunner"

export type ProductScreen = Exclude<ServiceScreen, "picker">

export interface ProductCard {
    screen: ProductScreen
    title: string
    description: string
}

export const PRODUCT_CARDS: ProductCard[] = [
    {screen: "generic", title: "Generic container/image", description: "Run any Docker image, optionally from Git"},
    {screen: "postgres", title: "PostgreSQL", description: "A managed PostgreSQL database instance"},
    {screen: "registry", title: "Container registry", description: "A private registry to store your images"},
    {screen: "githubRunner", title: "GitHub runner", description: "A self-hosted GitHub Actions runner"},
    {screen: "gitlabRunner", title: "GitLab runner", description: "A self-hosted GitLab CI runner"},
]

const SCREEN_TITLES: Record<ServiceScreen, string> = {
    picker: "Create service",
    generic: "Create app",
    postgres: "Create database",
    registry: "Create registry",
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
    return `Register container ${containerName ?? ""}`.trim()
}

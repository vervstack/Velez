import {RunnerProvider} from "@/app/api/velez"

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

import {Routes} from "@/app/router/Routes.ts"

// Derived labels the API attaches to GetService.labels — see
// internal/domain/service_labels.go's ClassifyService, the single source of
// truth these strings must match.
const AAS_ROUTES: {label: string; route: Routes}[] = [
    {label: "service-runner-gitlab", route: Routes.Runners},
    {label: "service-pgaas", route: Routes.Postgres},
    {label: "service-s3", route: Routes.S3},
    {label: "service-dind", route: Routes.Dinds},
]

export function aasRouteForLabels(labels: string[] | undefined): Routes | undefined {
    return AAS_ROUTES.find((entry) => labels?.includes(entry.label))?.route
}

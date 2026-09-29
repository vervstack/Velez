import {useParams} from "react-router-dom";

import {GetServiceByNameQuery} from "@/processes/queries/services.ts";
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx";
import ServiceInfoPage from "@/pages/service/ServiceInfoPage.tsx";
import VervCoreServicePage from "@/pages/service/VervCoreServicePage.tsx";
import GitlabRunnerServicePage from "@/pages/service/GitlabRunnerServicePage.tsx";

// Derived labels the API attaches to GetService.labels — see
// internal/domain/service_labels.go's ClassifyService, the single source of
// truth these strings must match.
const CORE_LABEL = "service-core";
const RUNNER_GITLAB_LABEL = "service-runner-gitlab";

// LABEL_PAGES is checked in order: the first matching label picks the detail
// page for /service/:key. Everything else falls through to the generic
// ServiceInfoPage. Add a row here (and a matching label in
// service_labels.go) for each new service type that needs its own page —
// e.g. a future "service-runner-github" entry.
const LABEL_PAGES: {label: string; Component: () => React.ReactElement}[] = [
    {label: RUNNER_GITLAB_LABEL, Component: GitlabRunnerServicePage},
    {label: CORE_LABEL, Component: VervCoreServicePage},
];

// ServiceRouteDispatch picks the detail page for /service/:key from the derived
// labels on GetService. Every candidate page re-runs the same query (React
// Query dedupes by key), so this only costs the branch.
export default function ServiceRouteDispatch() {
    const params = useParams<Record<string, string>>();
    const key = params["key"] || "";

    const serviceQuery = GetServiceByNameQuery(key);

    if (key !== "" && serviceQuery.isLoading) {
        return <SkeletonLoader shape="block" width="100%" height="12rem"/>;
    }

    const labels = serviceQuery.data?.labels ?? [];
    const match = LABEL_PAGES.find((entry) => labels.includes(entry.label));

    const Page = match?.Component ?? ServiceInfoPage;

    return <Page/>;
}

import {useParams} from "react-router-dom";

import {GetServiceByNameQuery} from "@/processes/queries/services.ts";
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx";

import ServiceInfoPage from "@/pages/service/ServiceInfoPage.tsx";
import VervCoreServicePage from "@/pages/service/VervCoreServicePage.tsx";

// service-core marks a Verv-stack infra service (velez, matreshka, …). Same
// string the API derives server-side and ServiceLabelBadge renders.
const CORE_LABEL = "service-core";

// ServiceRouteDispatch picks the detail page for /service/:key from the derived
// labels on GetService: core stack services get the read-only VervCoreServicePage,
// everything else the full ServiceInfoPage. Both target pages re-run the same
// query (React Query dedupes by key), so this only costs the branch.
export default function ServiceRouteDispatch() {
    const params = useParams<Record<string, string>>();
    const key = params["key"] || "";

    const serviceQuery = GetServiceByNameQuery(key);

    if (key !== "" && serviceQuery.isLoading) {
        return <SkeletonLoader shape="block" width="100%" height="12rem"/>;
    }

    const isCore = serviceQuery.data?.labels?.includes(CORE_LABEL) ?? false;

    return isCore ? <VervCoreServicePage/> : <ServiceInfoPage/>;
}

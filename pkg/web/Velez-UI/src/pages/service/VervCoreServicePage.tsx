import {useParams} from "react-router-dom";

import ServiceDetailLayout from "@/pages/service/widgets/ServiceDetailLayout.tsx";

// VervCoreServicePage is the read-only detail view for Verv-stack core services
// (velez, matreshka, makosh, headscale, portainer, angie). Same shared
// ServiceDetailLayout as ServiceInfoPage, minus every lifecycle action
// (Stop / Restart / Deploy / Remove) — these services aren't managed from this UI.

export default function VervCoreServicePage() {
    const params = useParams<Record<string, string>>();
    const key = params["key"] || "";

    return <ServiceDetailLayout serviceName={key}/>;
}

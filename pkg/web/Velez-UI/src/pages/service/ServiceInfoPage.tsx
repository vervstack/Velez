import {useParams} from "react-router-dom";

import ServiceDetailLayout from "@/pages/service/widgets/ServiceDetailLayout.tsx";
import ServiceLifecycleActions from "@/pages/service/widgets/ServiceLifecycleActions.tsx";

// ServiceInfoPage is the full detail view for user-managed Verv services —
// the shared ServiceDetailLayout plus the Stop / Restart / Deploy / Remove
// lifecycle actions in the header.

export default function ServiceInfoPage() {
    const params = useParams<Record<string, string>>();
    const key = params["key"] || "";

    return (
        <ServiceDetailLayout
            serviceName={key}
            headerActions={<ServiceLifecycleActions serviceName={key}/>}
        />
    );
}

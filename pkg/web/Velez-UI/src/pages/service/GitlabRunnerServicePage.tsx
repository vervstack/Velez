import {useParams} from "react-router-dom";

import ServiceDetailLayout from "@/pages/service/widgets/ServiceDetailLayout.tsx";
import GitlabRunnerActions from "@/pages/service/widgets/GitlabRunnerActions.tsx";
import GitlabRunnerSettings from "@/pages/service/widgets/GitlabRunnerSettings.tsx";

// GitlabRunnerServicePage is the detail view for a GitLab runner service —
// the shared ServiceDetailLayout plus runner-specific header actions
// (Stop / Restart / Rerun registration / Drop) and its settings segment
// (token, base URL, docker image, docker socket), instead of
// ServiceInfoPage's generic lifecycle actions.

export default function GitlabRunnerServicePage() {
    const params = useParams<Record<string, string>>();
    const key = params["key"] || "";

    return (
        <ServiceDetailLayout
            serviceName={key}
            headerActions={<GitlabRunnerActions serviceName={key}/>}
            extraContent={<GitlabRunnerSettings serviceName={key}/>}
        />
    );
}

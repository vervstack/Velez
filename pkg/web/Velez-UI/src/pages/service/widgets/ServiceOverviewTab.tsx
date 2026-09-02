import cls from "@/pages/service/widgets/ServiceOverviewTab.module.css";
import {GetServiceByNameQuery, ListDeploymentsByServiceNameQuery} from "@/processes/queries/services.ts";
import {ListSmerdsByServiceIdQuery} from "@/processes/queries/smerds.ts";
import ServiceHero from "@/widgets/service/ServiceHero/ServiceHero.tsx";
import ObservabilityTools from "@/widgets/service/ObservabilityTools/ObservabilityTools.tsx";
import ResourcesSection from "@/widgets/service/ResourcesSection/ResourcesSection.tsx";
import ServiceGraph from "@/widgets/service/ServiceGraph/ServiceGraph.tsx";
import DeploymentHistory from "@/widgets/service/DeploymentHistory/DeploymentHistory.tsx";
import Vervonomicon from "@/widgets/service/Vervonomicon/Vervonomicon.tsx";

interface Props {
    serviceName: string;
}

export default function ServiceOverviewTab({serviceName}: Props) {
    const serviceQuery = GetServiceByNameQuery(serviceName);
    const deploymentsQuery = ListDeploymentsByServiceNameQuery(serviceName);
    const smerdsQuery = ListSmerdsByServiceIdQuery(serviceName);

    const service = serviceQuery.data;
    const deployments = deploymentsQuery.data?.deployments || [];
    const currentSmerd = smerdsQuery.data?.smerds?.[0];

    return (
        <>
            <ServiceHero
                serviceName={serviceName}
                serviceStatus={service?.status as string | undefined}
                imageFromSmerd={currentSmerd?.imageName}
            />

            <div className={cls.ObservabilityWrapper}>
                <span className={cls.SectionTitle}>Observability &amp; Tools</span>
                <ObservabilityTools serviceName={serviceName}/>
            </div>

            <ResourcesSection serviceName={serviceName}/>
            <ServiceGraph serviceName={serviceName}/>
            <DeploymentHistory
                deployments={deployments}
                currentDeploymentId={service?.currentDeploymentId}
            />
            <Vervonomicon serviceName={serviceName}/>
        </>
    );
}

import cls from "@/pages/service/widgets/ServiceOverviewTab.module.css";
import {
    GetServiceByNameQuery,
    ListDeploymentsByServiceNameQuery,
    useGetServiceAboutQuery,
} from "@/processes/queries/services.ts";
import {s3InstanceNameOf} from "@/processes/mappings/s3.ts";
import {ListSmerdsByServiceIdQuery} from "@/processes/queries/smerds.ts";
import PgConnectionSection from "@/widgets/service/PgConnectionSection/PgConnectionSection.tsx";
import ServiceHero from "@/widgets/service/ServiceHero/ServiceHero.tsx";
import ObservabilityTools from "@/widgets/service/ObservabilityTools/ObservabilityTools.tsx";
import ResourcesSection from "@/widgets/service/ResourcesSection/ResourcesSection.tsx";
import ServiceGraph from "@/widgets/service/ServiceGraph/ServiceGraph.tsx";
import DeploymentHistory from "@/widgets/service/DeploymentHistory/DeploymentHistory.tsx";
import Vervonomicon from "@/widgets/service/Vervonomicon/Vervonomicon.tsx";
import S3Credentials from "@/widgets/S3Credentials/S3Credentials.tsx";

interface Props {
    serviceName: string;
}

export default function ServiceOverviewTab({serviceName}: Props) {
    const serviceQuery = GetServiceByNameQuery(serviceName);
    const deploymentsQuery = ListDeploymentsByServiceNameQuery(serviceName);
    const smerdsQuery = ListSmerdsByServiceIdQuery(serviceName);
    const aboutQuery = useGetServiceAboutQuery(serviceName);

    const service = serviceQuery.data;
    const deployments = deploymentsQuery.data?.deployments || [];
    const currentSmerd = smerdsQuery.data?.smerds?.[0];
    const s3InstanceName = s3InstanceNameOf(serviceName);

    return (
        <>
            <ServiceHero
                serviceName={serviceName}
                serviceStatus={service?.status as string | undefined}
                imageFromSmerd={currentSmerd?.imageName}
                containerId={currentSmerd?.uuid}
                containerStatus={currentSmerd?.status}
            />

            <PgConnectionSection serviceName={serviceName}/>

            <div className={cls.ObservabilityWrapper}>
                <span className={cls.SectionTitle}>Observability &amp; Tools</span>
                <ObservabilityTools serviceName={serviceName}/>
            </div>

            <ResourcesSection serviceName={serviceName}/>
            {s3InstanceName && <S3Credentials key={s3InstanceName} instanceName={s3InstanceName}/>}
            <ServiceGraph serviceName={serviceName} sidecars={aboutQuery.data?.sidecars}/>
            <DeploymentHistory
                deployments={deployments}
                currentDeploymentId={service?.currentDeploymentId}
            />
            <Vervonomicon serviceName={serviceName}/>
        </>
    );
}

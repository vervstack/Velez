import {ReactNode} from "react";

import cls from "@/pages/service/widgets/ServicePageHeader.module.css";
import {GetServiceByNameQuery} from "@/processes/queries/services.ts";
import EnvSwitcher from "@/widgets/service/EnvSwitcher/EnvSwitcher.tsx";
import {ServiceTab} from "@/pages/service/widgets/tabs.ts";
import TabStrip from "@/pages/service/widgets/TabStrip.tsx";
import ServiceTagsStrip from "@/pages/service/widgets/ServiceTagsStrip.tsx";

interface Props {
    serviceName: string;
    activeTab: ServiceTab;
    setActiveTab: (s: ServiceTab) => void;
    actions?: ReactNode;
}

export default function ServicePageHeader({serviceName, activeTab, setActiveTab, actions}: Props) {
    const serviceQuery = GetServiceByNameQuery(serviceName);
    const service = serviceQuery.data;

    if (!service || !service.name) return null;

    return (
        <div className={cls.ServicePageHeaderContainer}>
            <div className={cls.HeaderLeftWrapper}>
                <TabStrip activeTab={activeTab} setActiveTab={setActiveTab}/>
                {actions}
            </div>

            <div className={cls.HeaderRightWrapper}>
                <EnvSwitcher serviceName={serviceName}/>
                <ServiceTagsStrip serviceName={serviceName}/>
            </div>
        </div>
    );
}

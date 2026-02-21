import {useEffect, useState} from "react";
import cn from "classnames";

import cls from "@/pages/service/parts/DeployMenu.module.css";

import DeploymentWidget from "@/widgets/deploy/DeploymentWidget.tsx";
import {CreateNewDeployment} from "@/processes/api/service.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";

enum TabsOptions {
    New = 'New',
    Upgrade = 'Upgrade',
}

interface DeployMenuProps {
    serviceId: string;
    serviceName: string;
}

export default function DeployMenu(props: DeployMenuProps) {
    const [selectedTab, setSelectedTab] = useState<TabsOptions>(TabsOptions.New);

    const [content, setContent] = useState<React.ReactNode | null>(null);

    const toaster = useToaster()

    useEffect(() => {
        switch (selectedTab) {
            case TabsOptions.New:
                setContent(<DeploymentWidget
                    onCreate={(req) => {
                        req.name = props.serviceName
                        CreateNewDeployment(props.serviceId, req)
                            .catch(toaster.catchGrpc)
                    }}
                />)
                break
            case TabsOptions.Upgrade:
                setContent(<div> Not implemented yet</div>)
                break
            default:
                setContent(<div>Unknown deploy strategy {selectedTab}</div>)
        }
    }, [selectedTab]);


    const tabs = [
        {
            title: TabsOptions.New,
        },
        {
            title: TabsOptions.Upgrade,
        }
    ]

    return (
        <div className={cn(cls.DeployMenuContainer, {
            [cls.fullScreen]: selectedTab === TabsOptions.New,
        })}>
            <div className={cls.Tabs}>
                {tabs.map((tab, index) => (
                    <button
                        className={cn(cls.Tab, {
                            [cls.selected]: selectedTab === tab.title,
                        })}
                        key={index}
                        onClick={() => setSelectedTab(tab.title)}
                    >
                        {tab.title}
                    </button>
                ))}
            </div>
            <div className={cls.TabContent}>
                {content}
            </div>
        </div>
    )
}

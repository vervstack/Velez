import cls from "@/pages/service_info/parts/DeployMenu.module.css";
import {useEffect, useState} from "react";

import cn from "classnames";
import DeployWidget from "@/widgets/deploy/DeployWidget.tsx";

enum TabsOptions {
    New = 'New',
    Upgrade = 'Upgrade',
}

export default function DeployMenu() {
    const [selectedTab, setSelectedTab] = useState<TabsOptions>(TabsOptions.New);

    const [content, setContent] = useState<React.ReactNode | null>(null);


    useEffect(() => {
        switch (selectedTab) {
            case TabsOptions.New:
                setContent(<DeployWidget/>)
                break
            case TabsOptions.Upgrade:
                setContent(null)
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

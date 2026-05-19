import {useEffect, useState, useCallback} from "react";
import cn from "classnames";

import cls from "@/pages/service/parts/DeployMenu.module.css";

import DeploymentWidget from "@/widgets/deploy/DeploymentWidget.tsx";
import {serviceService} from "@/processes/api/service.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {ListDeploymentsByServiceNameQuery} from "@/processes/queries/services.ts";
import SkeletonDeploymentHistory from "@/components/deploy/SkeletonDeploymentHistory.tsx";

enum TabsOptions {
    New = 'New',
    Upgrade = 'Upgrade',
}

interface DeployMenuProps {
    serviceName: string;
    onDeploymentCreated?: () => void;
}

export default function DeployMenu(props: DeployMenuProps) {
    const [selectedTab, setSelectedTab] = useState<TabsOptions>(TabsOptions.New);
    const [content, setContent] = useState<React.ReactNode | null>(null);
    const toaster = useToaster();

    const deploymentsQuery = ListDeploymentsByServiceNameQuery(props.serviceName);
    useEffect(() => {
        if (deploymentsQuery.error) toaster.catchGrpc(deploymentsQuery.error);
    }, [deploymentsQuery.error]);

    const deployments = deploymentsQuery.data?.deployments || [];
    const latestDeployment = deployments.length > 0 ? deployments[0] : null;

    const handleUpgradeLatest = useCallback(function handleUpgradeLatest() {
        if (!latestDeployment?.id) {
            toaster.bake({
                title: "Cannot upgrade",
                description: "No previous deployment found",
                level: "Error",
            });
            return;
        }

        const confirmed = window.confirm(
            `Deploy latest from deployment ${latestDeployment.id}?\n` +
            `Image: ${latestDeployment.image || 'unknown'}`
        );

        if (!confirmed) {
            return;
        }

        serviceService.createNewDeployment(props.serviceName, {
            name: props.serviceName,
            imageName: latestDeployment.image,
        })
            .then(() => {
                toaster.bake({
                    title: "Deployment triggered",
                    description: `Upgrading to latest`,
                    level: "Info",
                });
                if (props.onDeploymentCreated) {
                    props.onDeploymentCreated();
                }
            })
            .catch(toaster.catchGrpc);
    }, [latestDeployment, props.serviceName, toaster]);

    useEffect(() => {
        switch (selectedTab) {
            case TabsOptions.New:
                setContent(<DeploymentWidget
                    onCreate={(req) => {
                        req.name = props.serviceName
                        serviceService.createNewDeployment(props.serviceName, req)
                            .then(() => {
                                toaster.bake({
                                    title: "Deployment triggered",
                                    description: `Creating new deployment for ${props.serviceName}`,
                                    level: "Info",
                                });
                                if (props.onDeploymentCreated) {
                                    props.onDeploymentCreated();
                                }
                            })
                            .catch(toaster.catchGrpc)
                    }}
                />)
                break
            case TabsOptions.Upgrade:
                setContent(<UpgradePanel
                    latestDeployment={latestDeployment}
                    onUpgrade={handleUpgradeLatest}
                    isLoading={deploymentsQuery.isLoading}
                />)
                break
            default:
                setContent(<div>Unknown deploy strategy {selectedTab}</div>)
        }
    }, [selectedTab, latestDeployment, handleUpgradeLatest, deploymentsQuery.isLoading]);

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

function UpgradePanel({latestDeployment, onUpgrade, isLoading}: {
    latestDeployment: any;
    onUpgrade: () => void;
    isLoading: boolean;
}) {
    if (isLoading) {
        return (
            <div className={cls.UpgradePanelContainer}>
                <SkeletonDeploymentHistory />
            </div>
        );
    }

    if (!latestDeployment) {
        return <div className={cls.UpgradePanelContainer}>No previous deployments found.</div>;
    }

    return (
        <div className={cls.UpgradePanelContainer}>
            <div className={cls.UpgradePanelContent}>
                <div className={cls.UpgradePanelTitle}>Deploy Latest</div>
                <div className={cls.UpgradePanelInfo}>
                    {latestDeployment.image && (
                        <div className={cls.UpgradePanelRow}>
                            <span className={cls.UpgradePanelLabel}>Image:</span>
                            <span className={cls.UpgradePanelValue}>{latestDeployment.image}</span>
                        </div>
                    )}
                    {latestDeployment.triggeredBy && (
                        <div className={cls.UpgradePanelRow}>
                            <span className={cls.UpgradePanelLabel}>Triggered by:</span>
                            <span className={cls.UpgradePanelValue}>{latestDeployment.triggeredBy}</span>
                        </div>
                    )}
                    <div className={cls.UpgradePanelRow}>
                        <span className={cls.UpgradePanelLabel}>Deployment ID:</span>
                        <span className={cls.UpgradePanelValue}>{latestDeployment.id}</span>
                    </div>
                </div>
                <div className={cls.UpgradePanelActions}>
                    <button onClick={onUpgrade} className={cls.UpgradeButton}>
                        Deploy Latest
                    </button>
                </div>
            </div>
        </div>
    );
}

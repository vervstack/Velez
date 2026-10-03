import {MouseEvent} from "react";

import cls from "@/pages/services/parts/ContainerCard/ContainerCard.module.css";
import type {DockerContainer} from "@/app/api/velez";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import {
    deriveContainerServicePresentation,
    mapContainerStatusToDot,
    containerStatusLabel,
    suggestedPatternHint,
} from "@/processes/mappings/containers";
import Button from "@/components/base/Button.tsx";
import StatusDot from "@/components/base/StatusDot";
import ServiceChip from "@/components/base/chips/ServiceChip";
import CreateServiceDialog from "@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx";
import FinishOnboardingButton from "@/pages/services/parts/FinishOnboardingButton/FinishOnboardingButton.tsx";

interface Props {
    container: DockerContainer;
    onOpen: (id: string) => void;
    onFilterByService: (serviceName: string) => void;
    isRegisterHidden?: boolean;
}

export default function ContainerCard({container, onOpen, onFilterByService, isRegisterHidden = false}: Props) {
    const {OpenDialog} = useDialog();

    function handleClick() {
        onOpen(container.id || "");
    }

    function handleActionsClick(e: MouseEvent) {
        e.stopPropagation();
    }

    function handleRegister() {
        OpenDialog(<CreateServiceDialog container={container}/>);
    }

    const isLeftover = Boolean(container.replacedByContainerId)
    const isRegisterOffered = !container.isRegistered && !container.linkedServiceName && !isRegisterHidden
        && !isLeftover

    const presentation = container.linkedServiceName
        ? deriveContainerServicePresentation(container)
        : null;
    const hint = suggestedPatternHint(container.suggestedPattern);

    return (
        <div className={cls.ContainerCardContainer} onClick={handleClick}>
            <div className={cls.NameRow}>
                <span className={cls.Name}>{container.name || container.id}</span>
                {!container.isRegistered && !isLeftover && <span className={cls.UnregisteredBadge}>not onboarded</span>}
                {isLeftover && <span className={cls.WaitingBadge}>waiting</span>}
            </div>
            <div className={cls.Image}>{container.imageName}</div>
            <div className={cls.StatusRow}>
                <StatusDot
                    status={mapContainerStatusToDot(container.status)}
                    tooltip={containerStatusLabel(container.status)}
                />
                {presentation && (
                    <ServiceChip
                        serviceName={presentation.displayName}
                        icon={presentation.icon}
                        onClick={onFilterByService}
                    />
                )}
            </div>
            {isLeftover && (
                <div className={cls.ActionsRow} onClick={handleActionsClick}>
                    <FinishOnboardingButton container={container}/>
                </div>
            )}
            {isRegisterOffered && (
                <div className={cls.ActionsRow} onClick={handleActionsClick}>
                    {hint && <span className={cls.Hint}>{hint}</span>}
                    <Button variant="secondary" sm onClick={handleRegister} tooltipContent="Onboard to Verv">
                        Onboard
                    </Button>
                </div>
            )}
        </div>
    );
}

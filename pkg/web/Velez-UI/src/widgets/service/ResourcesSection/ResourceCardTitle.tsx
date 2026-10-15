import cn from "classnames"

import type {ServiceResource} from "@/model/service_page/ServicePageModel"
import OnboardedIcon from "@/components/base/icons/OnboardedIcon.tsx"

import cls from "./ResourceCard.module.css"

interface Props {
    resource: ServiceResource
}

export default function ResourceCardTitle({resource}: Props) {
    const reconciliationLabel = reconciliationBadgeLabel(resource.reconciliation)

    return (
        <div className={cls.NameGroup}>
            <span className={cls.NameRow}>
                <span className={cls.ResourceName}>{resource.name}</span>
                {resource.isOnboarded && (
                    <span
                        className={cls.OnboardedBadge}
                        role="img"
                        aria-label="Onboarded in Velez"
                        data-tooltip-id="root-tooltip"
                        data-tooltip-content="Onboarded in Velez"
                        data-tooltip-place="top"
                    >
                        <OnboardedIcon className={cls.OnboardedIcon}/>
                    </span>
                )}
            </span>
            <span className={cls.KindDesc}>{resource.type}</span>
            {reconciliationLabel && (
                <span className={cn(cls.ReconciliationBadge, reconciliationBadgeClass(resource.reconciliation))}>
                    {reconciliationLabel}
                </span>
            )}
        </div>
    )
}

function reconciliationBadgeLabel(status: ServiceResource["reconciliation"]): string {
    switch (status) {
        case "already_connected":
            return "Existing connection"
        case "must_provision":
            return "Not yet created"
        default:
            return ""
    }
}

function reconciliationBadgeClass(status: ServiceResource["reconciliation"]): string {
    switch (status) {
        case "already_connected":
            return cls.alreadyConnected
        case "must_provision":
            return cls.mustProvision
        default:
            return ""
    }
}

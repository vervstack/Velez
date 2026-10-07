import {useQuery} from "@tanstack/react-query"
import cn from "classnames"

import cls from "@/dialogs/ResourceAddressesDialog/components/AddressRow/AddressRow.module.css"
import type {ResourceAddress} from "@/model/service_page/ServicePageModel"
import {pingAddress, resourceAddressUrl} from "@/processes/resourceAddresses.ts"

interface Props {
    address: ResourceAddress
}

const SCOPE_LABELS: Record<ResourceAddress["scope"], string> = {
    docker: "Public",
    vcn: "VCN",
}

export default function AddressRow({address}: Props) {
    const url = resourceAddressUrl(address, window.location)
    const pingQuery = useQuery({
        queryKey: ["address-ping", url],
        queryFn: () => pingAddress(url),
        retry: false,
        staleTime: 0,
        gcTime: 0,
    })

    const statusLabel = pingStatusLabel(pingQuery.isLoading, pingQuery.data)

    return (
        <div className={cls.AddressRowContainer}>
            <span
                className={cn(cls.StatusDot, statusDotClass(statusLabel))}
                role="img"
                aria-label={statusLabel}
                title={statusLabel}
            />
            <a className={cls.Url} href={url} target="_blank" rel="noreferrer">{url}</a>
            <span className={cls.ScopeLabel}>{SCOPE_LABELS[address.scope]}</span>
        </div>
    )
}

function pingStatusLabel(isLoading: boolean, isReachable?: boolean): string {
    if (isLoading) {
        return "Checking"
    }
    return isReachable ? "Reachable" : "Unreachable"
}

function statusDotClass(statusLabel: string): string {
    switch (statusLabel) {
        case "Reachable":
            return cls.healthy
        case "Unreachable":
            return cls.unhealthy
        default:
            return cls.unknown
    }
}

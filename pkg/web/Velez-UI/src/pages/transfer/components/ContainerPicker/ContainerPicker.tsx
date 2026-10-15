import {useMemo} from "react"
import {Dropdown} from "@vervstack/chures"
import type {DropdownOption} from "@vervstack/chures"

import type {Smerd} from "@/app/api/velez/velez_common.pb"
import {useListSmerdsQuery} from "@/processes/queries/smerds.ts"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"

function toOption(smerd: Smerd): DropdownOption {
    return {id: smerd.name ?? "", name: smerd.name ?? ""}
}

interface Props {
    value: string
    onChange(name: string): void
}

export default function ContainerPicker({value, onChange}: Props) {
    const smerdsQuery = useListSmerdsQuery()

    const options = useMemo(
        () => [...(smerdsQuery.data?.smerds ?? [])]
            .sort((a, b) => (a.name ?? "").localeCompare(b.name ?? ""))
            .map(toOption),
        [smerdsQuery.data]
    )

    function handleChange(ids: string[]) {
        onChange(ids[0] ?? "")
    }

    function handleRetry() {
        smerdsQuery.refetch()
    }

    if (smerdsQuery.isLoading) return <SkeletonLoader shape="line" height="2.5rem"/>
    if (smerdsQuery.isError) {
        return <QueryErrorState message="Failed to load containers." onRetry={handleRetry}/>
    }

    return (
        <Dropdown
            label="Container"
            placeholder="Select container"
            options={options}
            value={value ? [value] : []}
            onChange={handleChange}
            portal
        />
    )
}

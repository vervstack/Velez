import {Dropdown, DropdownOption} from "@vervstack/chures"

import cls
    from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/components/DockerDaemonPicker/DockerDaemonPicker.module.css"
import {useListDindsQuery} from "@/processes/queries/dinds.ts"
import Input from "@/components/base/Input.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import {
    EXTERNAL_DOCKER_CHOICE,
} from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/processes/dockerTarget.ts"

const EXTERNAL_OPTION: DropdownOption = {id: EXTERNAL_DOCKER_CHOICE, name: "External Docker daemon (advanced)"}

interface Props {
    choice: string
    externalAddress: string
    isDisabled: boolean
    onChoiceChange(choice: string): void
    onExternalAddressChange(address: string): void
    onCreateNew(): void
}

export default function DockerDaemonPicker({
    choice,
    externalAddress,
    isDisabled,
    onChoiceChange,
    onExternalAddressChange,
    onCreateNew,
}: Props) {
    const dindsQuery = useListDindsQuery()

    const options: DropdownOption[] = [
        ...(dindsQuery.data?.dinds ?? []).map(function toOption(dind) {
            return {id: dind.name ?? "", name: dind.name ?? ""}
        }),
        EXTERNAL_OPTION,
    ]

    function handleChange(ids: string[]) {
        onChoiceChange(ids[0] ?? "")
    }

    function handleRetry() {
        dindsQuery.refetch()
    }

    if (dindsQuery.isLoading) {
        return <SkeletonLoader shape="line" height="2.5rem"/>
    }

    if (dindsQuery.isError) {
        return <QueryErrorState message="Failed to load Docker daemons." onRetry={handleRetry}/>
    }

    return (
        <div className={cls.DockerDaemonPickerContainer}>
            <Dropdown
                label="Docker daemon"
                placeholder="Select Docker daemon"
                options={options}
                value={choice ? [choice] : []}
                onChange={handleChange}
                footerAction={{label: "Create new Docker daemon", onAction: onCreateNew}}
                portal
            />

            {choice === EXTERNAL_DOCKER_CHOICE && (
                <Input
                    label="Docker daemon address (tcp://host:port)"
                    inputValue={externalAddress}
                    onChange={onExternalAddressChange}
                    disabled={isDisabled}
                />
            )}
            {choice === EXTERNAL_DOCKER_CHOICE && (
                <div className={cls.RiskNotice}>
                    Jobs run on the daemon at this address. Only use a daemon you trust.
                </div>
            )}
        </div>
    )
}

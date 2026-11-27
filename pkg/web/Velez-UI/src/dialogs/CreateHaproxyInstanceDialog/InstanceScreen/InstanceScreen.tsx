import {Dropdown, Input, type DropdownOption} from "@vervstack/chures"

import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import type {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"
import {useListResourceBoxesQuery} from "@/processes/queries/settings.ts"
import ExposePortScreen from "@/dialogs/CreateHaproxyInstanceDialog/ExposePortScreen/ExposePortScreen.tsx"
import type {HaproxyWizardContext} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"

export default function InstanceScreen(props: StepScreenProps<HaproxyWizardContext>) {
    const {name, box, sidecarContainerId, updateContext} = props
    const query = useListResourceBoxesQuery()
    const boxes = query.data?.boxes ?? []
    const options = boxes.flatMap((item) => item.name ? [{id: item.name, name: item.name}] : [])

    function renderBoxOption(option: DropdownOption) {
        if (typeof option === "string") return option

        const item = boxes.find((entry) => entry.name === option.id)
        if (!item) return option.name

        return <span>{item.name} — cpu: {item.cpu}, ram: {item.ramMb} MB, disk: {item.diskMb} MB</span>
    }

    return (
        <div className={cls.Fields}>
            <Input label="Name" value={name} setValue={(value) => updateContext({name: value})}
                   error={validateInstanceName(name.trim())} disabled={Boolean(sidecarContainerId)}/>
            {query.isLoading && <SkeletonLoader shape="line"/>}
            {query.isError && <QueryErrorState message="Failed to load resource boxes." onRetry={query.refetch}/>}
            {query.isSuccess && (
                <Dropdown label="Box" placeholder="Default size" options={options} value={box ? [box] : []}
                          renderOption={renderBoxOption} onChange={(ids) => updateContext({box: ids[0] ?? ""})} portal/>
            )}
            <ExposePortScreen {...props}/>
        </div>
    )
}

import {Input} from "@vervstack/chures"

import type {VcnNetworkWizardContext} from "@/dialogs/CreateVcnNetworkDialog/processes/vcnNetworkWizard.ts"
import type {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"

type Props = StepScreenProps<VcnNetworkWizardContext>

export default function NetworkNameScreen({name, updateContext}: Props) {
    return (
        <Input label="Network name" value={name} setValue={(value) => updateContext({name: value})}
               error={validateInstanceName(name)} autoFocus/>
    )
}

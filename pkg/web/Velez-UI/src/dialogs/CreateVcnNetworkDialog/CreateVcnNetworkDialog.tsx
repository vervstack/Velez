import {
    canProceedFromName,
    canProceedFromSource,
    initialContext,
    type VcnNetworkWizardContext,
} from "@/dialogs/CreateVcnNetworkDialog/processes/vcnNetworkWizard.ts"
import NetworkCreationScreen
    from "@/dialogs/CreateVcnNetworkDialog/screens/NetworkCreationScreen/NetworkCreationScreen.tsx"
import NetworkNameScreen from "@/dialogs/CreateVcnNetworkDialog/screens/NetworkNameScreen/NetworkNameScreen.tsx"
import NetworkSourceScreen from "@/dialogs/CreateVcnNetworkDialog/screens/NetworkSourceScreen/NetworkSourceScreen.tsx"
import {OpenWizardDialog} from "@/dialogs/StepsDialog/openWizardDialog.tsx"
import type {Step} from "@/dialogs/StepsDialog/Step.ts"

const steps: Step<VcnNetworkWizardContext>[] = [
    {name: "name", label: "Name", component: NetworkNameScreen, canProceed: canProceedFromName},
    {
        name: "source",
        label: "Source",
        nextLabel: "Create",
        component: NetworkSourceScreen,
        canProceed: canProceedFromSource,
    },
]

export function openCreateVcnNetworkDialog() {
    OpenWizardDialog({
        steps,
        initialContext,
        header: {icon: "⇄", eyebrow: "Create VCN network", eyebrowDetail: "Headscale"},
        finalStep: NetworkCreationScreen,
    })
}

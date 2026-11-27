import CreateHaproxyProgress
    from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyProgress/CreateHaproxyProgress.tsx"
import InstanceScreen from "@/dialogs/CreateHaproxyInstanceDialog/InstanceScreen/InstanceScreen.tsx"
import PostgresScreen from "@/dialogs/CreateHaproxyInstanceDialog/PostgresScreen/PostgresScreen.tsx"
import VcnSetupScreen from "@/dialogs/CreateHaproxyInstanceDialog/VcnSetupScreen/VcnSetupScreen.tsx"
import {type HaproxyWizardContext, initialContext} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"
import type {Step} from "@/dialogs/StepsDialog/Step.ts"
import {OpenWizardDialog} from "@/dialogs/StepsDialog/openWizardDialog.tsx"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"

function canProceedFromInstance(context: HaproxyWizardContext): boolean {
    const name = context.name.trim()
    return Boolean(name && !validateInstanceName(name) && (!context.isPortExposed || context.isPortReady))
}

const STEPS: Step<HaproxyWizardContext>[] = [
    {name: "instance", label: "Instance", component: InstanceScreen, canProceed: canProceedFromInstance},
    {name: "vcn", label: "VCN setup", component: VcnSetupScreen, isNextHidden: true,
        isBusy: (context) => context.isBootstrapping},
    {name: "postgres", label: "PostgreSQL instances", nextLabel: "Create", component: PostgresScreen},
]

export function openCreateHaproxyInstanceDialog() {
    OpenWizardDialog({
        steps: STEPS,
        initialContext,
        railLabel: "Onboarding",
        header: {icon: "⇄", eyebrow: "Create HAProxy instance", eyebrowDetail: "PostgreSQL balancer"},
        finalStep: CreateHaproxyProgress,
    })
}

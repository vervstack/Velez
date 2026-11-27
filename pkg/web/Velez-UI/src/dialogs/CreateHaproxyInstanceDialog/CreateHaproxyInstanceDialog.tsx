import CreateHaproxyProgress
    from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyProgress/CreateHaproxyProgress.tsx"
import ExposePortScreen from "@/dialogs/CreateHaproxyInstanceDialog/ExposePortScreen/ExposePortScreen.tsx"
import HaproxyNetworkScreen from "@/dialogs/CreateHaproxyInstanceDialog/HaproxyNetworkScreen/HaproxyNetworkScreen.tsx"
import PostgresScreen from "@/dialogs/CreateHaproxyInstanceDialog/PostgresScreen/PostgresScreen.tsx"
import ReviewScreen from "@/dialogs/CreateHaproxyInstanceDialog/ReviewScreen/ReviewScreen.tsx"
import type {Step} from "@/dialogs/StepsDialog/Step.ts"
import {OpenWizardDialog} from "@/dialogs/StepsDialog/openWizardDialog.tsx"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"
import {type HaproxyWizardContext, initialContext} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"

function canSelectPostgres(context: HaproxyWizardContext): boolean {
    const name = context.name.trim()
    return Boolean(context.master && name && !validateInstanceName(name))
}

const STEPS: Step<HaproxyWizardContext>[] = [
    {name: "postgres", label: "PostgreSQL", component: PostgresScreen,
        canProceed: canSelectPostgres},
    {name: "network", label: "Network access", component: HaproxyNetworkScreen,
        canProceed: (context) => context.isNetworkReady},
    {name: "port", label: "Expose port", component: ExposePortScreen,
        canProceed: (context) => !context.isPortExposed || context.isPortReady},
    {name: "review", label: "Review", nextLabel: "Create", component: ReviewScreen},
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

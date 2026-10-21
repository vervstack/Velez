import type {DockerContainer} from "@/app/api/velez"
import OnboardConfirm from "@/dialogs/CreateServiceDialog/components/OnboardConfirm/OnboardConfirm.tsx"
import OnboardProgress from "@/dialogs/CreateServiceDialog/components/OnboardProgress/OnboardProgress.tsx"
import {
    initialOnboardContext,
    isTypeStepReady,
    OnboardContext,
} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import OnboardContainersScreen
    from "@/dialogs/CreateServiceDialog/screens/OnboardContainersScreen/OnboardContainersScreen.tsx"
import OnboardReviewScreen from "@/dialogs/CreateServiceDialog/screens/OnboardReviewScreen/OnboardReviewScreen.tsx"
import OnboardTypeScreen from "@/dialogs/CreateServiceDialog/screens/OnboardTypeScreen/OnboardTypeScreen.tsx"
import {Step} from "@/dialogs/StepsDialog/Step.ts"
import {OpenWizardDialog} from "@/dialogs/StepsDialog/openWizardDialog.tsx"

function isReviewReady(context: OnboardContext): boolean {
    return context.form?.request != null
}

const STEPS: Step<OnboardContext>[] = [
    {name: "containers", label: "Containers", component: OnboardContainersScreen},
    {name: "type", label: "Type", component: OnboardTypeScreen, canProceed: isTypeStepReady},
    {
        name: "review",
        label: "Review",
        nextLabel: "Onboard",
        component: OnboardReviewScreen,
        canProceed: isReviewReady,
    },
]

function shouldConfirmOnboard(context: OnboardContext): boolean {
    return context.form?.isRecreating ?? true
}

export function openOnboardDialog(container: DockerContainer, networkMembers: DockerContainer[] = []): void {
    OpenWizardDialog<OnboardContext>({
        steps: STEPS,
        initialContext: initialOnboardContext(container, networkMembers),
        header: {icon: "⇪", eyebrow: "Onboard container", eyebrowDetail: container.name},
        railLabel: "Onboarding",
        initialStepIndex: networkMembers.length > 0 ? 0 : 1,
        confirmFinish: OnboardConfirm,
        shouldConfirmFinish: shouldConfirmOnboard,
        finalStep: OnboardProgress,
    })
}

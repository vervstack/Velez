import RegisterProgress from "@/dialogs/CreateServiceDialog/components/RegisterProgress/RegisterProgress.tsx"
import {OnboardContext} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import {FinalStepProps} from "@/dialogs/StepsDialog/Step.ts"

export default function OnboardProgress({context}: FinalStepProps<OnboardContext>) {
    const request = context.form?.request
    if (!request) return null

    return <RegisterProgress request={request} containerName={context.container.name}/>
}

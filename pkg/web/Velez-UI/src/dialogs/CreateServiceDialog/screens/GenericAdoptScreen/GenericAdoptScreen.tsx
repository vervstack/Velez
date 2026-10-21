import type {DockerContainer} from "@/app/api/velez"
import AdoptForm from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx"
import type {OnboardFormState} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"

interface Props {
    container: DockerContainer
    onFormChange(form: OnboardFormState): void
}

export default function GenericAdoptScreen({container, onFormChange}: Props) {
    return <AdoptForm container={container} pattern="generic" onFormChange={onFormChange}/>
}

import type {DockerContainer} from "@/app/api/velez"
import AdoptForm from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.tsx"

interface Props {
    container: DockerContainer
    onBusyChange(isBusy: boolean): void
}

export default function GenericAdoptScreen({container, onBusyChange}: Props) {
    return <AdoptForm container={container} pattern="generic" onBusyChange={onBusyChange}/>
}

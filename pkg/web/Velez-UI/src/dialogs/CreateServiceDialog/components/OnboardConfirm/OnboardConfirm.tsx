import RestartConfirm from "@/dialogs/CreateServiceDialog/components/RestartConfirm/RestartConfirm.tsx"
import RootSwitchConfirm from "@/dialogs/CreateServiceDialog/components/RootSwitchConfirm/RootSwitchConfirm.tsx"
import {
    isRootChanged,
    nameOfContainer,
    OnboardContext,
} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import {ConfirmFinishProps} from "@/dialogs/StepsDialog/Step.ts"

export default function OnboardConfirm({context, onConfirm, onCancel}: ConfirmFinishProps<OnboardContext>) {
    if (isRootChanged(context)) {
        return (
            <RootSwitchConfirm
                oldRootName={nameOfContainer(context, context.container.id ?? "")}
                newRootName={nameOfContainer(context, context.roles.rootId)}
                onConfirm={onConfirm}
                onCancel={onCancel}
            />
        )
    }

    return (
        <RestartConfirm
            containerName={context.container.name}
            hasPublishedPorts={context.form?.hasPublishedPorts ?? false}
            onConfirm={onConfirm}
            onCancel={onCancel}
        />
    )
}

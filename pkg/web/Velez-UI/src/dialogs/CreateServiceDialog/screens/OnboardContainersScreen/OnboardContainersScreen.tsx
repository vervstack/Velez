import cls from "@/dialogs/CreateServiceDialog/screens/OnboardContainersScreen/OnboardContainersScreen.module.css"
import type {DockerContainer} from "@/app/api/velez"
import {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import ContainerOverviewRow from "@/dialogs/CreateServiceDialog/components/ContainerOverviewRow/ContainerOverviewRow.tsx"
import RootChangeWarning from "@/dialogs/CreateServiceDialog/components/RootChangeWarning/RootChangeWarning.tsx"
import {
    isRootChanged,
    nameOfContainer,
    networkModeOf,
    OnboardContext,
} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import {pickRoot, pickService, serviceIdOf} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"

export default function OnboardContainersScreen(props: StepScreenProps<OnboardContext>) {
    const {updateContext, ...context} = props
    const {containers, roles} = context
    const isReadOnly = containers.length === 1
    const memberIds = containers.map((member) => member.id ?? "")

    function handlePickRoot(containerId: string) {
        updateContext({roles: pickRoot(roles, containerId, memberIds)})
    }

    function handlePickService(containerId: string) {
        updateContext({roles: pickService(roles, containerId)})
    }

    function renderRow(member: DockerContainer) {
        const memberId = member.id ?? ""

        return (
            <ContainerOverviewRow
                key={memberId}
                container={member}
                networkMode={networkModeOf(member, context)}
                isRoot={roles.rootId === memberId}
                isService={serviceIdOf(roles) === memberId}
                isReadOnly={isReadOnly}
                onPickRoot={handlePickRoot}
                onPickService={handlePickService}
            />
        )
    }

    return (
        <div className={cls.OnboardContainersScreenContainer}>
            <p className={cls.Intro}>Pick the network root and the service. The service defaults to the root.</p>
            <div className={cls.Rows}>
                {containers.map(renderRow)}
            </div>
            {isReadOnly && <p className={cls.Note}>Only one container: nothing to choose.</p>}
            <RootChangeWarning
                isVisible={isRootChanged(context)}
                oldRootName={nameOfContainer(context, context.container.id ?? "")}
                newRootName={nameOfContainer(context, roles.rootId)}
            />
        </div>
    )
}

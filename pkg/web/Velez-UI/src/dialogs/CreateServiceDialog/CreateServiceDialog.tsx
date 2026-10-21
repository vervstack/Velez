import {useState} from "react"

import cls from "@/dialogs/CreateServiceDialog/CreateServiceDialog.module.css"
import {DockerContainer} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import NetworkGroupList from "@/dialogs/CreateServiceDialog/components/NetworkGroupList/NetworkGroupList.tsx"
import RootSwitchConfirm from "@/dialogs/CreateServiceDialog/components/RootSwitchConfirm/RootSwitchConfirm.tsx"
import {AdoptGroupContext} from "@/dialogs/CreateServiceDialog/hooks/AdoptGroup.ts"
import {useAdoptRoles} from "@/dialogs/CreateServiceDialog/hooks/AdoptRoles.ts"
import {
    ADOPTABLE_SCREENS,
    adoptTitle,
    registerPatternOf,
    runnerProviderOf,
    screenTitle,
    ServiceScreen,
    suggestedScreenOf,
} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"
import DindScreen from "@/dialogs/CreateServiceDialog/screens/DindScreen/DindScreen.tsx"
import GenericAdoptScreen from "@/dialogs/CreateServiceDialog/screens/GenericAdoptScreen/GenericAdoptScreen.tsx"
import GenericScreen from "@/dialogs/CreateServiceDialog/screens/GenericScreen/GenericScreen.tsx"
import PickerScreen from "@/dialogs/CreateServiceDialog/screens/PickerScreen/PickerScreen.tsx"
import PostgresAdoptScreen from "@/dialogs/CreateServiceDialog/screens/PostgresAdoptScreen/PostgresAdoptScreen.tsx"
import PostgresScreen from "@/dialogs/CreateServiceDialog/screens/PostgresScreen/PostgresScreen.tsx"
import RegistryAdoptScreen from "@/dialogs/CreateServiceDialog/screens/RegistryAdoptScreen/RegistryAdoptScreen.tsx"
import RegistryScreen from "@/dialogs/CreateServiceDialog/screens/RegistryScreen/RegistryScreen.tsx"
import RunnerAdoptScreen from "@/dialogs/CreateServiceDialog/screens/RunnerAdoptScreen/RunnerAdoptScreen.tsx"
import RunnerScreen from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/RunnerScreen.tsx"
import S3Screen from "@/dialogs/CreateServiceDialog/screens/S3Screen/S3Screen.tsx"

interface Props {
    initialScreen?: ServiceScreen
    suggestedScreen?: ServiceScreen
    container?: DockerContainer
    networkMembers?: DockerContainer[]
}

export default function CreateServiceDialog({
    initialScreen = "picker",
    suggestedScreen,
    container,
    networkMembers = [],
}: Props) {
    const [screen, setScreen] = useState<ServiceScreen>(initialScreen)
    const [isBusy, setIsBusy] = useState(false)
    const {CloseDialog} = useDialog()

    const adoptRoles = useAdoptRoles(container, networkMembers)
    const {isRolesChanged, pendingRootId} = adoptRoles
    const effectiveScreen: ServiceScreen = isRolesChanged && screen !== "picker" ? "generic" : screen

    const canGoBack = initialScreen === "picker" && screen !== "picker" && !isBusy

    function handleBack() {
        setScreen("picker")
    }

    function handleClose() {
        CloseDialog()
    }

    function renderAdoptScreen(adopted: DockerContainer) {
        if (effectiveScreen === "generic") return <GenericAdoptScreen container={adopted} onBusyChange={setIsBusy}/>
        if (effectiveScreen === "postgres") return <PostgresAdoptScreen container={adopted} onBusyChange={setIsBusy}/>
        if (effectiveScreen === "registry") return <RegistryAdoptScreen container={adopted} onBusyChange={setIsBusy}/>
        if (effectiveScreen === "githubRunner" || effectiveScreen === "gitlabRunner") {
            return (
                <RunnerAdoptScreen
                    container={adopted}
                    initialProvider={runnerProviderOf(effectiveScreen)}
                    onBusyChange={setIsBusy}
                />
            )
        }
        return (
            <PickerScreen
                suggestedScreen={isRolesChanged ? undefined : suggestedScreenOf(adopted.suggestedPattern)}
                enabledScreens={isRolesChanged ? ["generic"] : ADOPTABLE_SCREENS}
                onSelect={setScreen}
            />
        )
    }

    function renderAdopt(adopted: DockerContainer) {
        const serviceScreen = screen === "picker" ? suggestedScreenOf(adopted.suggestedPattern) : effectiveScreen

        return (
            <AdoptGroupContext.Provider value={adoptRoles.group}>
                <div className={cls.AdoptBodyWrapper} hidden={pendingRootId !== null}>
                    {!isBusy && (
                        <NetworkGroupList
                            group={adoptRoles.group}
                            pattern={registerPatternOf(effectiveScreen)}
                            serviceScreen={serviceScreen}
                            onPickService={adoptRoles.onPickService}
                            onRequestRoot={adoptRoles.onRequestRoot}
                        />
                    )}
                    {renderAdoptScreen(adopted)}
                </div>
                {pendingRootId !== null && (
                    <RootSwitchConfirm
                        oldRootName={adoptRoles.oldRootName}
                        newRootName={adoptRoles.pendingRootName}
                        onConfirm={adoptRoles.onConfirmRoot}
                        onCancel={adoptRoles.onCancelRoot}
                    />
                )}
            </AdoptGroupContext.Provider>
        )
    }

    function renderScreen() {
        if (container) return renderAdopt(container)
        if (screen === "generic") return <GenericScreen onBusyChange={setIsBusy} onOpenScreen={setScreen}/>
        if (screen === "postgres") return <PostgresScreen onBusyChange={setIsBusy}/>
        if (screen === "registry") return <RegistryScreen onBusyChange={setIsBusy}/>
        if (screen === "s3") return <S3Screen onBusyChange={setIsBusy}/>
        if (screen === "dind") return <DindScreen onBusyChange={setIsBusy}/>
        if (screen === "githubRunner" || screen === "gitlabRunner") {
            return <RunnerScreen initialProvider={runnerProviderOf(screen)} onBusyChange={setIsBusy}/>
        }
        return <PickerScreen suggestedScreen={suggestedScreen} onSelect={setScreen}/>
    }

    return (
        <div className={cls.CreateServiceDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>{container ? adoptTitle(container.name) : screenTitle(screen)}</h2>
                {canGoBack && <Button variant="ghost" sm onClick={handleBack}>Back</Button>}
                {!isBusy && <Button variant="ghost" sm onClick={handleClose}>✕</Button>}
            </div>
            {renderScreen()}
        </div>
    )
}

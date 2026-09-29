import {useState} from "react"

import cls from "@/dialogs/CreateServiceDialog/CreateServiceDialog.module.css"
import {DockerContainer} from "@/app/api/velez"
import Button from "@/components/base/Button.tsx"
import {
    ADOPTABLE_SCREENS,
    adoptTitle,
    runnerProviderOf,
    screenTitle,
    ServiceScreen,
    suggestedScreenOf,
} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"
import GenericAdoptScreen from "@/dialogs/CreateServiceDialog/screens/GenericAdoptScreen/GenericAdoptScreen.tsx"
import GenericScreen from "@/dialogs/CreateServiceDialog/screens/GenericScreen/GenericScreen.tsx"
import PickerScreen from "@/dialogs/CreateServiceDialog/screens/PickerScreen/PickerScreen.tsx"
import PostgresAdoptScreen from "@/dialogs/CreateServiceDialog/screens/PostgresAdoptScreen/PostgresAdoptScreen.tsx"
import PostgresScreen from "@/dialogs/CreateServiceDialog/screens/PostgresScreen/PostgresScreen.tsx"
import RegistryScreen from "@/dialogs/CreateServiceDialog/screens/RegistryScreen/RegistryScreen.tsx"
import RunnerScreen from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/RunnerScreen.tsx"

interface Props {
    initialScreen?: ServiceScreen
    suggestedScreen?: ServiceScreen
    container?: DockerContainer
}

export default function CreateServiceDialog({initialScreen = "picker", suggestedScreen, container}: Props) {
    const [screen, setScreen] = useState<ServiceScreen>(initialScreen)
    const [isBusy, setIsBusy] = useState(false)

    const canGoBack = initialScreen === "picker" && screen !== "picker" && !isBusy

    function handleBack() {
        setScreen("picker")
    }

    function renderAdoptScreen(adopted: DockerContainer) {
        if (screen === "generic") return <GenericAdoptScreen container={adopted} onBusyChange={setIsBusy}/>
        if (screen === "postgres") return <PostgresAdoptScreen container={adopted} onBusyChange={setIsBusy}/>
        return (
            <PickerScreen
                suggestedScreen={suggestedScreenOf(adopted.suggestedPattern)}
                enabledScreens={ADOPTABLE_SCREENS}
                onSelect={setScreen}
            />
        )
    }

    function renderScreen() {
        if (container) return renderAdoptScreen(container)
        if (screen === "generic") return <GenericScreen onBusyChange={setIsBusy}/>
        if (screen === "postgres") return <PostgresScreen onBusyChange={setIsBusy}/>
        if (screen === "registry") return <RegistryScreen onBusyChange={setIsBusy}/>
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
            </div>
            {renderScreen()}
        </div>
    )
}

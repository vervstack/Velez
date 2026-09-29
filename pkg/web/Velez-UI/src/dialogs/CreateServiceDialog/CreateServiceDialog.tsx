import {useState} from "react"

import cls from "@/dialogs/CreateServiceDialog/CreateServiceDialog.module.css"
import Button from "@/components/base/Button.tsx"
import {runnerProviderOf, screenTitle, ServiceScreen} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"
import GenericScreen from "@/dialogs/CreateServiceDialog/screens/GenericScreen/GenericScreen.tsx"
import PickerScreen from "@/dialogs/CreateServiceDialog/screens/PickerScreen/PickerScreen.tsx"
import PostgresScreen from "@/dialogs/CreateServiceDialog/screens/PostgresScreen/PostgresScreen.tsx"
import RegistryScreen from "@/dialogs/CreateServiceDialog/screens/RegistryScreen/RegistryScreen.tsx"
import RunnerScreen from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/RunnerScreen.tsx"

interface Props {
    initialScreen?: ServiceScreen
    suggestedScreen?: ServiceScreen
}

export default function CreateServiceDialog({initialScreen = "picker", suggestedScreen}: Props) {
    const [screen, setScreen] = useState<ServiceScreen>(initialScreen)
    const [isBusy, setIsBusy] = useState(false)

    const canGoBack = initialScreen === "picker" && screen !== "picker" && !isBusy

    function handleBack() {
        setScreen("picker")
    }

    function renderScreen() {
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
                <h2 className={cls.Title}>{screenTitle(screen)}</h2>
                {canGoBack && <Button variant="ghost" sm onClick={handleBack}>Back</Button>}
            </div>
            {renderScreen()}
        </div>
    )
}

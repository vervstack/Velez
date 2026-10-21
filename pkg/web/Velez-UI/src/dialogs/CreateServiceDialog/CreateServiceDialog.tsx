import {useState} from "react"

import cls from "@/dialogs/CreateServiceDialog/CreateServiceDialog.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import {runnerProviderOf, screenTitle, ServiceScreen} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"
import DindScreen from "@/dialogs/CreateServiceDialog/screens/DindScreen/DindScreen.tsx"
import GenericScreen from "@/dialogs/CreateServiceDialog/screens/GenericScreen/GenericScreen.tsx"
import PickerScreen from "@/dialogs/CreateServiceDialog/screens/PickerScreen/PickerScreen.tsx"
import PostgresScreen from "@/dialogs/CreateServiceDialog/screens/PostgresScreen/PostgresScreen.tsx"
import RegistryScreen from "@/dialogs/CreateServiceDialog/screens/RegistryScreen/RegistryScreen.tsx"
import RunnerScreen from "@/dialogs/CreateServiceDialog/screens/RunnerScreen/RunnerScreen.tsx"
import S3Screen from "@/dialogs/CreateServiceDialog/screens/S3Screen/S3Screen.tsx"

interface Props {
    initialScreen?: ServiceScreen
}

export default function CreateServiceDialog({initialScreen = "picker"}: Props) {
    const [screen, setScreen] = useState<ServiceScreen>(initialScreen)
    const [isBusy, setIsBusy] = useState(false)
    const {CloseDialog} = useDialog()

    const canGoBack = initialScreen === "picker" && screen !== "picker" && !isBusy

    function handleBack() {
        setScreen("picker")
    }

    function handleClose() {
        CloseDialog()
    }

    function renderScreen() {
        if (screen === "generic") return <GenericScreen onBusyChange={setIsBusy} onOpenScreen={setScreen}/>
        if (screen === "postgres") return <PostgresScreen onBusyChange={setIsBusy}/>
        if (screen === "registry") return <RegistryScreen onBusyChange={setIsBusy}/>
        if (screen === "s3") return <S3Screen onBusyChange={setIsBusy}/>
        if (screen === "dind") return <DindScreen onBusyChange={setIsBusy}/>
        if (screen === "githubRunner" || screen === "gitlabRunner") {
            return <RunnerScreen initialProvider={runnerProviderOf(screen)} onBusyChange={setIsBusy}/>
        }
        return <PickerScreen onSelect={setScreen}/>
    }

    return (
        <div className={cls.CreateServiceDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>{screenTitle(screen)}</h2>
                {canGoBack && <Button variant="ghost" sm onClick={handleBack}>Back</Button>}
                {!isBusy && <Button variant="ghost" sm onClick={handleClose}>✕</Button>}
            </div>
            {renderScreen()}
        </div>
    )
}

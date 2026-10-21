import {useEffect} from "react"

import cls from "@/dialogs/CreateServiceDialog/screens/OnboardTypeScreen/OnboardTypeScreen.module.css"
import {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import TypeHint from "@/dialogs/CreateServiceDialog/components/TypeHint/TypeHint.tsx"
import {
    enabledScreensOf,
    isRolesChanged,
    OnboardContext,
    selectedScreenOf,
} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import {ProductScreen, suggestedScreenOf} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"
import PickerScreen from "@/dialogs/CreateServiceDialog/screens/PickerScreen/PickerScreen.tsx"

const HINT_TEXT = "Choose the type of this service to continue."
const GENERIC_ONLY_NOTE = "The service or network root differs from the original container, "
    + "so only the Generic type is available and ports, bind-mount links and the image tag stay as they are."

export default function OnboardTypeScreen(props: StepScreenProps<OnboardContext>) {
    const {updateContext, ...context} = props
    const selectedScreen = selectedScreenOf(context)
    const isChanged = isRolesChanged(context)

    useEffect(() => {
        if (selectedScreen !== context.screen) {
            updateContext({screen: selectedScreen})
        }
    }, [selectedScreen, context.screen, updateContext])

    function handleSelect(screen: ProductScreen) {
        updateContext({screen})
    }

    return (
        <div className={cls.OnboardTypeScreenContainer}>
            <TypeHint isVisible={selectedScreen === undefined} text={HINT_TEXT}/>
            {isChanged && <p className={cls.Note}>{GENERIC_ONLY_NOTE}</p>}
            <PickerScreen
                suggestedScreen={isChanged ? undefined : suggestedScreenOf(context.container.suggestedPattern)}
                enabledScreens={enabledScreensOf(context)}
                selectedScreen={selectedScreen}
                onSelect={handleSelect}
            />
        </div>
    )
}

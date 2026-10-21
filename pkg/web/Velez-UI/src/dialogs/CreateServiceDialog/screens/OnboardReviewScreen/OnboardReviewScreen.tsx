import cls from "@/dialogs/CreateServiceDialog/screens/OnboardReviewScreen/OnboardReviewScreen.module.css"
import {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import NetworkGroupList from "@/dialogs/CreateServiceDialog/components/NetworkGroupList/NetworkGroupList.tsx"
import {AdoptGroup, AdoptGroupContext} from "@/dialogs/CreateServiceDialog/hooks/AdoptGroup.ts"
import {
    OnboardContext,
    OnboardFormState,
    selectedScreenOf,
    serviceNameOf,
} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import {registerPatternOf, runnerProviderOf} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"
import GenericAdoptScreen from "@/dialogs/CreateServiceDialog/screens/GenericAdoptScreen/GenericAdoptScreen.tsx"
import PostgresAdoptScreen from "@/dialogs/CreateServiceDialog/screens/PostgresAdoptScreen/PostgresAdoptScreen.tsx"
import RegistryAdoptScreen from "@/dialogs/CreateServiceDialog/screens/RegistryAdoptScreen/RegistryAdoptScreen.tsx"
import RunnerAdoptScreen from "@/dialogs/CreateServiceDialog/screens/RunnerAdoptScreen/RunnerAdoptScreen.tsx"

export default function OnboardReviewScreen(props: StepScreenProps<OnboardContext>) {
    const {updateContext, ...context} = props
    const screen = selectedScreenOf(context)

    function handleServiceNameChange(name: string) {
        updateContext({serviceNameDraft: name})
    }

    const group: AdoptGroup = {
        containers: context.containers,
        originalRootId: context.container.id ?? "",
        roles: context.roles,
        serviceName: serviceNameOf(context),
        onServiceNameChange: handleServiceNameChange,
    }

    function handleFormChange(form: OnboardFormState) {
        updateContext({form})
    }

    function renderForm() {
        if (screen === "postgres") {
            return <PostgresAdoptScreen container={context.container} onFormChange={handleFormChange}/>
        }
        if (screen === "registry") {
            return <RegistryAdoptScreen container={context.container} onFormChange={handleFormChange}/>
        }
        if (screen === "githubRunner" || screen === "gitlabRunner") {
            return (
                <RunnerAdoptScreen
                    container={context.container}
                    initialProvider={runnerProviderOf(screen)}
                    onFormChange={handleFormChange}
                />
            )
        }
        return <GenericAdoptScreen container={context.container} onFormChange={handleFormChange}/>
    }

    if (screen === undefined) return null

    return (
        <AdoptGroupContext.Provider value={group}>
            <div className={cls.OnboardReviewScreenContainer}>
                {renderForm()}
                <section className={cls.ResultSection}>
                    <h3 className={cls.SectionTitle}>Result</h3>
                    <NetworkGroupList group={group} pattern={registerPatternOf(screen)} serviceScreen={screen}/>
                </section>
            </div>
        </AdoptGroupContext.Provider>
    )
}

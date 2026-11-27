import type {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import {
    createRequest,
    formatEndpoint,
    type HaproxyWizardContext,
    networkLabels,
    networkMode,
} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"

type ScreenProps = StepScreenProps<HaproxyWizardContext>

export default function ReviewScreen(context: ScreenProps) {
    const mode = networkMode(context)
    const target = createRequest(context).connectionTarget
    const endpoint = target?.publicEndpoint ?? target?.tailscaleEndpoint
    const targetLabel = endpoint
        ? formatEndpoint(endpoint.host ?? "", String(endpoint.port ?? ""))
        : context.dockerNetwork
    const environment = mode === "docker" && context.environment ? context.environment : "Default"

    return (
        <dl className={cls.Review}>
            <dt>Name</dt><dd>{context.name.trim()}</dd>
            <dt>PostgreSQL</dt><dd>{context.masterLabel}</dd>
            <dt>Network access</dt><dd>{networkLabels[mode]}</dd>
            <dt>Target</dt><dd>{targetLabel}</dd>
            <dt>Deployment environment</dt><dd>{environment}</dd>
            {context.box.trim() && <><dt>Box</dt><dd>{context.box.trim()}</dd></>}
            <dt>Published port</dt><dd>{context.isPortExposed ? context.port : "Off"}</dd>
        </dl>
    )
}

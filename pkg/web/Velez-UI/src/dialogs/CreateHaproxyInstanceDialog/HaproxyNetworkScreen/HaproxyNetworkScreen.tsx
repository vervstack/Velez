
import {useEffect, useId, useState} from "react"
import {Input} from "@vervstack/chures"

import type {HaproxyConnectionTargetEndpoint} from "@/app/api/velez/haproxy_api.pb"
import TargetStatus from "@/dialogs/CreateHaproxyInstanceDialog/TargetStatus/TargetStatus.tsx"
import type {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import {useDockerHaproxyTargetQuery, useValidateHaproxyTargetQuery} from "@/processes/queries/haproxy.ts"
import {
    formatEndpoint,
    type HaproxyWizardContext,
    isIpAddress,
    networkLabels,
    networkMode,
    type NetworkMode,
    normalizeHost,
    parsePort,
} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"

const VALIDATION_DELAY_MS = 350
const MODES: NetworkMode[] = ["docker", "public", "tailscale"]

export default function HaproxyNetworkScreen(context: StepScreenProps<HaproxyWizardContext>) {
    const {host, targetPort, tailscaleIp, tailscalePort, updateContext} = context
    const mode = networkMode(context)
    const radioName = useId()
    const dockerTarget = useDockerHaproxyTargetQuery(context.master, context.environment)
    const normalizedHost = normalizeHost(host)
    const parsedPort = parsePort(targetPort)
    const [publicEndpoint, setPublicEndpoint] = useState<HaproxyConnectionTargetEndpoint>()
    const isPublicHostValid = isIpAddress(normalizedHost) || /^[a-z0-9][a-z0-9_.-]*$/i.test(normalizedHost)
    const isPublicTargetCurrent = mode === "public" && isPublicHostValid && parsedPort !== undefined &&
        publicEndpoint?.host === normalizedHost && publicEndpoint?.port === parsedPort

    useEffect(() => {
        if (mode !== "public" || !isPublicHostValid || parsedPort === undefined) return

        const timer = window.setTimeout(() => {
            setPublicEndpoint({host: normalizedHost, port: parsedPort})
        }, VALIDATION_DELAY_MS)
        return () => window.clearTimeout(timer)
    }, [mode, isPublicHostValid, normalizedHost, parsedPort])

    const publicTarget = useValidateHaproxyTargetQuery(isPublicTargetCurrent ? {
        masterPgInstance: context.master, connectionTarget: {publicEndpoint},
    } : undefined)
    const dockerNetwork = dockerTarget.data ? dockerTarget.data.dockerNetwork ?? "" : context.dockerNetwork
    const isDockerReady = Boolean(dockerTarget.data?.ok && dockerNetwork && !dockerTarget.isFetching)
    const isPublicReady = Boolean(isPublicTargetCurrent && publicTarget.data?.ok && !publicTarget.isFetching)
    const isTailscaleReady = isIpAddress(normalizeHost(tailscaleIp)) && parsePort(tailscalePort) !== undefined
    const isNetworkReady = mode === "docker" ? isDockerReady : mode === "public" ? isPublicReady : isTailscaleReady

    useEffect(() => {
        if (dockerNetwork !== context.dockerNetwork || isNetworkReady !== context.isNetworkReady) {
            updateContext({dockerNetwork, isNetworkReady})
        }
    }, [dockerNetwork, isNetworkReady, context.dockerNetwork, context.isNetworkReady, updateContext])

    function handlePublicHostChange(value: string) {
        updateContext({host: value, mode: "public", isNetworkReady: false})
    }

    function handlePublicPortChange(value: string) {
        updateContext({targetPort: value, mode: "public", isNetworkReady: false})
    }

    return (
        <div className={cls.Fields}>
            <fieldset className={cls.ModeOptions}>
                <legend className={cls.VisuallyHidden}>Network access</legend>
                {MODES.filter((option) => option !== "docker" || dockerNetwork).map((option) => (
                    <label className={cls.ModeOption} key={option}>
                        {/* eslint-disable-next-line no-restricted-syntax -- Chures has no radio input. */}
                        <input type="radio" name={radioName} checked={mode === option}
                               onChange={() => updateContext({mode: option, isNetworkReady: false})}/>
                        {networkLabels[option]}
                    </label>
                ))}
            </fieldset>
            {mode === "docker" && (
                <>
                    <span>Docker network · {dockerNetwork}</span>
                    <TargetStatus query={dockerTarget}/>
                </>
            )}
            {mode === "public" && (
                <>
                    <div className={cls.Endpoint}>
                        <Input label="PostgreSQL host" value={host}
                               setValue={handlePublicHostChange}/>
                        <Input label="PostgreSQL port" value={targetPort}
                               setValue={handlePublicPortChange}/>
                    </div>
                    {normalizedHost && parsedPort !== undefined && (
                        <span>{formatEndpoint(normalizedHost, targetPort)}</span>
                    )}
                    {isPublicTargetCurrent && <TargetStatus query={publicTarget}/>}
                </>
            )}
            {mode === "tailscale" && (
                <>
                    <div className={cls.Endpoint}>
                        <Input label="PostgreSQL Tailscale IP" value={tailscaleIp}
                               setValue={(value) => updateContext({tailscaleIp: value})}/>
                        <Input label="PostgreSQL port" value={tailscalePort}
                               setValue={(value) => updateContext({tailscalePort: value})}/>
                    </div>
                    <Input label="Auth key (optional)" type="password" autoComplete="off" value={context.authKey}
                           setValue={(value) => updateContext({authKey: value})}/>
                    <small>Leave empty to use cluster Headscale</small>
                    <Input label="Login server URL (optional)" value={context.loginServerUrl}
                           setValue={(value) => updateContext({loginServerUrl: value})}/>
                    <Input label="Hostname (optional)" value={context.hostname}
                           setValue={(value) => updateContext({hostname: value})}/>
                </>
            )}
        </div>
    )
}

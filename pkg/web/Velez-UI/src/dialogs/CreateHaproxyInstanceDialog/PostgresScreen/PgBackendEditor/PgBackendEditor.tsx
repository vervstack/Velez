import {Button, Dropdown, Input} from "@vervstack/chures"
import {useMemo, useState} from "react"

import type {PgBackend} from "@/app/api/velez/haproxy_api.pb"
import Choice from "@/components/base/Choice.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import {PgBackendRole} from "@/processes/api/haproxy.ts"
import {canAttachNetwork} from "@/processes/mappings/networks.ts"
import {useListPgInstancesQuery} from "@/processes/queries/pg_instances.ts"
import {useListNetworksQuery} from "@/processes/queries/networks.ts"
import {DEFAULT_PORT, parsePort} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"
import {usePgBackendValidation}
    from "@/dialogs/CreateHaproxyInstanceDialog/PostgresScreen/PgBackendEditor/usePgBackendValidation.ts"

interface Props {
    backends: PgBackend[]
    sidecarContainerId?: string
    onAdd(backend: PgBackend): void
    onCancel(): void
}

type NetworkMode = "tailscale" | "docker" | "public"

const MODES: {id: NetworkMode; name: string}[] = [
    {id: "tailscale", name: "Tailscale"},
    {id: "docker", name: "Docker network"},
    {id: "public", name: "Public internet"},
]

export default function PgBackendEditor({backends, sidecarContainerId, onAdd, onCancel}: Props) {
    const [source, setSource] = useState<"velez" | "own">()
    const [serviceName, setServiceName] = useState("")
    const [mode, setMode] = useState<NetworkMode>()
    const [networkId, setNetworkId] = useState("")
    const [host, setHost] = useState("")
    const [port, setPort] = useState(String(DEFAULT_PORT))
    const [user, setUser] = useState("")
    const [password, setPassword] = useState("")

    const pgQuery = useListPgInstancesQuery(source === "velez")
    const networks = useListNetworksQuery("", true)
    const instances = pgQuery.data?.instances
    const selected = instances?.find((pg) => pg.name === serviceName)
    const pgOptions = (instances ?? []).flatMap((pg) => {
        if (!pg.name || backends.some((backend) => backend.velezService === pg.name)) return []

        const name = pg.displayName || pg.name
        return [{id: pg.name, name: pg.environment ? `${name} (${pg.environment})` : name}]
    })
    const networkOptions = (networks.data?.networks ?? []).flatMap((network) =>
        canAttachNetwork(network) && network.isIccEnabled && network.name && network.id
            ? [{id: network.id, name: network.name}]
            : [])

    const backend = useMemo<PgBackend | undefined>(() => {
        const role = backends.length === 0 ? PgBackendRole.PRIMARY : PgBackendRole.BACKUP
        if (source === "velez" && selected?.name) {
            return {name: selected.name, velezService: selected.name, environment: selected.environment, role}
        }
        const hostValue = host.trim().replace(/^\[([^\]]+)\]$/, "$1")
        const portValue = parsePort(port)
        if (source !== "own" || !hostValue || !portValue || !user.trim() || !password || !mode) return
        if ((mode === "tailscale" && !sidecarContainerId) || (mode === "docker" && !networkId)) return

        const endpoint = {host: hostValue, port: portValue}
        const connectionTarget = mode === "docker" ? {dockerNetwork: networkId} :
            mode === "tailscale" ? {tailscaleEndpoint: endpoint} : {publicEndpoint: endpoint}
        let number = backends.length + 1
        while (backends.some((item) => item.name === `postgres-${number}`)) number++
        return {name: `postgres-${number}`, endpoint, user: user.trim(), password, role, connectionTarget}
    }, [backends, host, mode, networkId, password, port, selected, sidecarContainerId, source, user])

    const {isChecking, error, retest} = usePgBackendValidation(backend, sidecarContainerId, onAdd)

    return (
        <div className={cls.Backend}>
            <div className={cls.Options} role="group" aria-label="PostgreSQL instance source">
                <Choice title="Pick from Velez Postgres instances" active={source === "velez"}
                        onClick={() => setSource("velez")}/>
                <Choice title="Setup my own Postgres instance" active={source === "own"}
                        onClick={() => setSource("own")}/>
            </div>
            {source === "velez" && (
                <>
                    {pgQuery.isLoading && <SkeletonLoader shape="line"/>}
                    {pgQuery.isError && <QueryErrorState message="Failed to load PostgreSQL instances."
                                                         onRetry={pgQuery.refetch}/>}
                    {pgQuery.isSuccess && (
                        <Dropdown label="PostgreSQL instance" options={pgOptions}
                                  value={serviceName ? [serviceName] : []} portal
                                  onChange={(ids) => setServiceName(ids[0] ?? "")}/>
                    )}
                </>
            )}
            {source === "own" && (
                <>
                    <div className={cls.Options} role="group" aria-label="Network access">
                        {MODES.map(({id, name}) => {
                            const isDisabled = id === "tailscale" && !sidecarContainerId
                            return (
                                <div className={cls.Option} key={id}
                                     title={isDisabled ? "Set up a VCN sidecar to use Tailscale" : undefined}>
                                    <Choice title={name} active={mode === id} disabled={isDisabled}
                                            onClick={() => setMode(id)}/>
                                </div>
                            )
                        })}
                    </div>
                    {mode === "docker" && (
                        <>
                            {networks.isLoading && <SkeletonLoader shape="line"/>}
                            {networks.isError && <QueryErrorState message="Failed to load Docker networks."
                                                                  onRetry={networks.refetch}/>}
                            {networks.isSuccess && (
                                <Dropdown label="Docker network" options={networkOptions} portal
                                          value={networkId ? [networkId] : []}
                                          onChange={(ids) => setNetworkId(ids[0] ?? "")}/>
                            )}
                        </>
                    )}
                    <div className={cls.Endpoint}>
                        <Input label="Host" value={host} setValue={setHost}/>
                        <Input label="Port" value={port} setValue={setPort}/>
                    </div>
                    <div className={cls.Endpoint}>
                        <Input label="User" value={user} setValue={setUser}/>
                        <Input label="Password" type="password" autoComplete="off"
                               value={password} setValue={setPassword}/>
                    </div>
                </>
            )}
            <div className={cls.Actions}>
                <small role="status" className={error ? cls.Unreachable : undefined}>
                    {isChecking ? "Checking PostgreSQL…" : error}
                </small>
                <Button variant="secondary" type="button" disabled={!backend || isChecking}
                        onClick={retest}>Retest</Button>
                <Button variant="secondary" type="button" onClick={onCancel}>Cancel</Button>
            </div>
        </div>
    )
}

import {queryOptions, useQuery} from '@tanstack/react-query'
import {controlPlaneService} from '@/processes/api/control_plane'
import {FetchNodeHardware} from '@/processes/api/velez.ts'
import {GetInitReq} from '@/processes/api/api.ts'
import {VervPluginState, VervPluginType} from "@/app/api/velez";

export function ListNodesQuery() {
    return useQuery({
        queryKey: ["nodes"],
        queryFn: () => controlPlaneService.listNodes(),
    })
}

export function nodesSidebar() {
    return queryOptions({
        queryKey: ["nodes_main_layout"],
        queryFn: () => controlPlaneService.listNodes(),
    })
}

export function ListPluginsQuery() {
    return useQuery({
        queryKey: ["plugins"],
        queryFn: () => controlPlaneService.listPlugins(),
    })
}

export function ListEnvironmentsQuery() {
    return useQuery({
        queryKey: ["environments"],
        queryFn: () => controlPlaneService.listEnvironments(),
    })
}

export function hardwareQueryOptions() {
    return queryOptions({
        queryKey: ["hardware"] as const,
        queryFn: () => FetchNodeHardware(GetInitReq()),
        staleTime: 5 * 60 * 1000,
    })
}

export function NodeHardwareQuery() {
    return useQuery(hardwareQueryOptions())
}

export function IsStatefullModeEnabled(): boolean {
    return ListPluginsQuery().data
        ?.find(p =>
            p.type == VervPluginType.statefull_pg &&
            p.state == VervPluginState.running)
        !== undefined;
}

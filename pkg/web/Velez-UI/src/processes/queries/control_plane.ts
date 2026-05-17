import {queryOptions, useQuery} from '@tanstack/react-query'
import { controlPlaneService } from '@/processes/api/control_plane'

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

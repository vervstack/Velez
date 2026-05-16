import { queryOptions } from '@tanstack/react-query'
import { CacheKey } from '@/app/query/Cache'
import { controlPlaneService } from '@/processes/api/control_plane'

export function listNodesQuery() {
    return queryOptions({
        queryKey: [CacheKey.Nodes],
        queryFn: () => controlPlaneService.listNodes(),
    })
}

export function listPluginsQuery() {
    return queryOptions({
        queryKey: [CacheKey.Plugins],
        queryFn: () => controlPlaneService.listPlugins(),
    })
}

import {useQuery} from "@tanstack/react-query"

import {velezService} from "@/processes/api/velez.ts"

const VERSION_QUERY_KEY = ["version"]
const VERSION_STALE_TIME_MS = 60 * 60 * 1000

export function useVersionQuery() {
    return useQuery({
        queryKey: VERSION_QUERY_KEY,
        queryFn: () => velezService.getVersion(),
        staleTime: VERSION_STALE_TIME_MS,
        retry: false,
    })
}

import {useMutation, useQuery, useQueryClient} from "@tanstack/react-query"

import {runnersService} from "@/processes/api/runners"
import {CreateRunnerRequest} from "@/app/api/velez"

const RUNNERS_QUERY_KEY = ["runners"]
const LIST_REQ = {paging: {limit: "50", offset: "0"}}

export function useListRunnersQuery() {
    return useQuery({
        queryKey: RUNNERS_QUERY_KEY,
        queryFn: () => runnersService.listRunners(LIST_REQ),
    })
}

export function CreateRunnerMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (req: CreateRunnerRequest) => runnersService.createRunner(req),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: RUNNERS_QUERY_KEY})
        },
    })
}

export function DropRunnerMutation() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (name: string) => runnersService.dropRunner(name),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: RUNNERS_QUERY_KEY})
        },
    })
}

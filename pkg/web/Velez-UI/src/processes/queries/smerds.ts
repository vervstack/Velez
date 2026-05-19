import {useQuery} from '@tanstack/react-query'
import {FetchSmerds, FetchSmerd, FetchSmerdsByServiceId} from '@/processes/api/velez'
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";

export function useListSmerdsQuery() {
    return useQuery({
        queryKey: ['smerds'] as const,
        queryFn: FetchSmerds,
    })
}

export function useGetSmerdQuery(name: string, enabled = true) {
    return useQuery({
        queryKey: ['smerd', name] as const,
        queryFn: () => FetchSmerd(name),
        enabled,
    })
}

export function ListSmerdsByServiceIdQuery(serviceName: string) {
    const toaster = useToaster();

    return useQuery({
        queryKey: ['service-smerds', 'name', serviceName] as const,
        queryFn: () =>
            FetchSmerdsByServiceId(serviceName)
                .catch(toaster.catchGrpc),
    })
}

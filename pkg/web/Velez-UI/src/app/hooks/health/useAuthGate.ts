import {useQuery} from "@tanstack/react-query"

import {velezService} from "@/processes/api/velez.ts"

export type AuthGateStatus = "checking" | "authorized" | "unauthorized"

export const AUTH_GATE_QUERY_KEY = ["auth-gate"] as const

export function useAuthGate(): AuthGateStatus {
    const query = useQuery({
        queryKey: AUTH_GATE_QUERY_KEY,
        queryFn: () => velezService.ping(),
        retry: false,
        staleTime: Infinity,
        refetchOnWindowFocus: false,
    })

    if (query.isSuccess) {
        return "authorized"
    }

    if (query.isError) {
        return "unauthorized"
    }

    return "checking"
}

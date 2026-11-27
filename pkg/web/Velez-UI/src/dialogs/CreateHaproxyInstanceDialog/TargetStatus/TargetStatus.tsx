import type {useValidateHaproxyTargetQuery} from "@/processes/queries/haproxy.ts"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"

export default function TargetStatus({query}: {query: ReturnType<typeof useValidateHaproxyTargetQuery>}) {
    if (query.isFetching || query.isPending) {
        return <span role="status">Checking reachability…</span>
    }

    if (query.data?.ok) {
        return <span role="status" className={cls.Reachable}>Reachable</span>
    }

    const message = query.data?.errorMessage ?? query.error?.message
    return <span role="status" className={cls.Unreachable}>{message ? `Unreachable · ${message}` : "Unreachable"}</span>
}

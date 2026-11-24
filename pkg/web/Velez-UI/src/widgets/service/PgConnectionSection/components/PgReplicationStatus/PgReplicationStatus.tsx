import cls from "@/widgets/service/PgConnectionSection/components/PgReplicationStatus/PgReplicationStatus.module.css"

import type {GetPgReplicationStatusReplicaStatus} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import ChangePgMasterDialog from "@/dialogs/ChangePgMasterDialog/ChangePgMasterDialog.tsx"
import {usePgReplicationStatusQuery} from "@/processes/queries/pg_instances.ts"

interface Props {
    instance: string
}

export default function PgReplicationStatus({instance}: Props) {
    const statusQuery = usePgReplicationStatusQuery(instance)
    const OpenDialog = useDialog((state) => state.OpenDialog)

    function handlePromote() {
        const status = statusQuery.data
        if (status?.isReplica && status.inSync && status.upstream) {
            OpenDialog(<ChangePgMasterDialog newMaster={instance} upstream={status.upstream}/>)
        }
    }

    function renderPromoteButton() {
        const canPromote = statusQuery.data?.inSync && Boolean(statusQuery.data.upstream)
        const tooltip = canPromote ? undefined : "Wait until the replica is in sync."

        return <span
            className={cls.PromoteButtonWrapper}
            tabIndex={canPromote ? undefined : 0}
            aria-label={tooltip}
            data-tooltip-id={tooltip && "root-tooltip"}
            data-tooltip-content={tooltip}
        >
            <Button variant="secondary" sm disabled={!canPromote} onClick={handlePromote}>
                Promote to master
            </Button>
        </span>
    }

    function renderReplica(replica: GetPgReplicationStatusReplicaStatus, index: number) {
        return <li className={cls.StatusRow} key={`${replica.applicationName}-${index}`}>
            <strong>{replica.applicationName || "Unnamed replica"}</strong>
            <span className={cls.Detail}>{replica.state}</span>
            <span className={cls.Detail}>Lag {replica.lagBytes ?? "0"} B · {replica.replayLagSeconds ?? 0} s</span>
            {replica.inSync && <span className={cls.InSync}>In sync</span>}
        </li>
    }

    function renderStatus() {
        if (statusQuery.isLoading) {
            return <SkeletonLoader shape="line" height="1rem" width="10rem"/>
        }
        if (statusQuery.isError) {
            return <QueryErrorState message="Could not load replication status." onRetry={statusQuery.refetch}/>
        }
        const status = statusQuery.data
        if (status?.isReplica) {
            return <div className={cls.StatusRow}>
                <strong>Replica of {status.upstream || "unknown upstream"}</strong>
                {status.inSync && <span className={cls.InSync}>In sync</span>}
                {renderPromoteButton()}
            </div>
        }
        if (!status?.replicas?.length) {
            return <span className={cls.Detail}>No replicas connected</span>
        }
        return <ul className={cls.Replicas}>{status.replicas.map(renderReplica)}</ul>
    }

    return <section className={cls.PgReplicationStatusContainer} aria-label="Replication status">
        <h3 className={cls.Title}>Replication status</h3>
        {renderStatus()}
    </section>
}

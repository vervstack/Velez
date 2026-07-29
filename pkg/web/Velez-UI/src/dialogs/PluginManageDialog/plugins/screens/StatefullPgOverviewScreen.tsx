import PipelineFlowLoader from '@/dialogs/StepsDialog/PipelineFlowLoader.tsx';
import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

export default function StatefullPgOverviewScreen() {
    return (
        <div className={cls.ActionSection}>
            <PipelineFlowLoader/>

            <span className={cls.DisabledMessage}>
                statefull_pg is currently disabled. Enabling it provisions a managed Postgres cluster
                on this node that other services can bind to via connection-string environment
                variables. Depending on the host, it will run either as a container or a bound binary.
                Exposing a port, configured on the next step, is optional and only needed for external
                access.
            </span>
        </div>
    );
}

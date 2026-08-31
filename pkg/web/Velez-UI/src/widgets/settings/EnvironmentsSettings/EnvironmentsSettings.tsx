import cls from '@/widgets/settings/EnvironmentsSettings/EnvironmentsSettings.module.css';
import type {Environment} from '@/app/api/velez';
import {IsStatefullModeEnabled, ListEnvironmentsQuery} from "@/processes/queries/control_plane.ts";
import Button from "@/components/base/Button.tsx";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import EnvironmentCreateDialog from "@/dialogs/EnvironmentCreateDialog/EnvironmentCreateDialog.tsx";
import EnvironmentManageDialog from "@/dialogs/EnvironmentManageDialog/EnvironmentManageDialog.tsx";

export default function EnvironmentsSettings() {
    const envQuery = ListEnvironmentsQuery()
    const {OpenDialog, CloseDialog} = useDialog();

    function handleManage(env: Environment) {
        OpenDialog(
            <EnvironmentManageDialog
                environment={env}
                onCancel={CloseDialog}
                onSaved={CloseDialog}
                onDeleted={CloseDialog}
            />
        )
    }

    function handleAddEnvironment() {
        OpenDialog(
            <EnvironmentCreateDialog
                onCancel={CloseDialog}
                onCreated={CloseDialog}
            />
        )
    }

    const environments = envQuery.data?.environments || []

    const isSingleNodeMode = !IsStatefullModeEnabled()

    return (
        <div className={cls.EnvironmentsSettingsContainer}>
            <div className={cls.SectionTitle}>Environments</div>
            {!envQuery.isLoading && (
                <div className={cls.CardsWrapper}>
                    {environments.map((e) =>
                        <EnvRow key={e.id} env={e} handleManage={handleManage}/>)}
                    {environments.length === 0 && (
                        <div className={cls.EmptyState}>No environments configured</div>
                    )}
                </div>)}
            <div
                data-tooltip-id={isSingleNodeMode ? 'root-tooltip' : undefined}
                data-tooltip-content={isSingleNodeMode ? 'Not available in single node mode' : undefined}
            >
                <Button
                    disabled={isSingleNodeMode}
                    onClick={handleAddEnvironment}>
                    + Add environment
                </Button>
            </div>
        </div>
    );
}


function EnvRow({env, handleManage}: { env: Environment, handleManage: (e: Environment) => void }) {
    function onManage() {
        handleManage(env);
    }

    const isStateFullModeEnabled = IsStatefullModeEnabled()
    return (
        <div className={cls.EnvRowWrapper}>
            <div className={cls.EnvInfoWrapper}>
                <span className={cls.EnvName}>{env.name}</span>
                {env.suffix && <span className={cls.EnvSuffix}>{env.suffix}</span>}
            </div>
            <div
                data-tooltip-id={isStateFullModeEnabled ? undefined : 'root-tooltip'}
                data-tooltip-content={isStateFullModeEnabled ? undefined : 'Not available in single node mode'}
            >
                <Button
                    disabled={!isStateFullModeEnabled}
                    onClick={onManage}>
                    Manage
                </Button>
            </div>
        </div>
    );

}

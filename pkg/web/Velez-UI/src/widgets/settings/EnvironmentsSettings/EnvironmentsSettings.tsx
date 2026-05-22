import cls from '@/widgets/settings/EnvironmentsSettings/EnvironmentsSettings.module.css';
import EnvCard from '@/components/complex/EnvCard/EnvCard.tsx';
import type {ServiceEnvironment} from '@/model/service_page/ServicePageModel';
import {IsStatefullModeEnabled, ListEnvironmentsQuery} from "@/processes/queries/control_plane.ts";
import Button from "@/components/base/Button.tsx";

export default function EnvironmentsSettings() {
    const envQuery = ListEnvironmentsQuery()

    function handleManage(env: ServiceEnvironment) {
        // TODO
        console.log(env)
    }

    function handleAddEnvironment() {
        // TODO
    }

    const environments = envQuery.data?.environments?.map(e => {
        return {
            id: e,
            label: e,
            // TODO
            status: 'running',
            version: '',
            deployedAgo: '',
            health: 'healthy',
        } as ServiceEnvironment
    }) || []

    const isSingleNodeMode = !IsStatefullModeEnabled()

    return (
        <div className={cls.EnvironmentsSettingsContainer}>
            <div className={cls.SectionTitle}>Environments</div>
            {!envQuery.isLoading && (
                <div className={cls.CardsWrapper}>
                    {environments.map((e) =>
                        <EnvRow env={e} handleManage={handleManage}/>)}
                    {environments.length === 0 && (
                        <div className={cls.EmptyState}>No environments configured</div>
                    )}
                </div>)}
            <div
                data-tooltip-id={'root-tooltip'}
                data-tooltip-content={'Not available in single node mode'}
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


function EnvRow({env, handleManage}: { env: ServiceEnvironment, handleManage: (e: ServiceEnvironment) => void }) {
    function onManage() {
        handleManage(env);
    }

    const isStateFullModeEnabled = IsStatefullModeEnabled()
    return (
        <div key={env.id} className={cls.EnvRowWrapper}>
            <EnvCard env={env}/>
            <div
                data-tooltip-id={'root-tooltip'}
                data-tooltip-content={'Not available in single node mode'}
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

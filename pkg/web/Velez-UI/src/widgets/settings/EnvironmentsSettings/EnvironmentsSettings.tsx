import cls from '@/widgets/settings/EnvironmentsSettings/EnvironmentsSettings.module.css';
import EnvCard from '@/components/complex/EnvCard/EnvCard.tsx';
import type { ServiceEnvironment } from '@/model/service_page/ServicePageModel';

export default function EnvironmentsSettings() {
    const envs: ServiceEnvironment[] = []; // TODO: wire to real query

    function handleManage(_env: ServiceEnvironment) {
        // TODO
    }

    function handleAddEnvironment() {
        // TODO
    }

    return (
        <div className={cls.EnvironmentsSettingsContainer}>
            <div className={cls.SectionTitle}>Environments</div>
            <div className={cls.CardsWrapper}>
                {envs.map(function renderEnvRow(env: ServiceEnvironment) {
                    function onManage() {
                        handleManage(env);
                    }

                    return (
                        <div key={env.id} className={cls.EnvRowWrapper}>
                            <EnvCard env={env} />
                            <button className={cls.ManageButton} onClick={onManage}>Manage</button>
                        </div>
                    );
                })}
                {envs.length === 0 && (
                    <div className={cls.EmptyState}>No environments configured</div>
                )}
            </div>
            <button className={cls.AddButton} onClick={handleAddEnvironment}>+ Add environment</button>
        </div>
    );
}

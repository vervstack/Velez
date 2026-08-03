import {useEffect} from 'react';

import cls from '@/widgets/environment/EnvironmentSwitcher/EnvironmentSwitcher.module.css';

import {ListEnvironmentsQuery} from '@/processes/queries/control_plane.ts';
import {useEnvironmentStore} from '@/app/hooks/environment/Environment.ts';

export default function EnvironmentSwitcher() {
    const envQuery = ListEnvironmentsQuery();
    const environments = useEnvironmentStore((state) => state.environments);
    const selectedEnvironment = useEnvironmentStore((state) => state.selectedEnvironment);
    const setEnvironments = useEnvironmentStore((state) => state.setEnvironments);
    const selectEnvironment = useEnvironmentStore((state) => state.selectEnvironment);

    useEffect(() => {
        if (envQuery.data?.environments) {
            setEnvironments(envQuery.data.environments);
        }
        // setEnvironments is a stable zustand action reference; only re-sync when the fetched list changes.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [envQuery.data]);

    function handleChange(e: React.ChangeEvent<HTMLSelectElement>) {
        selectEnvironment(e.target.value);
    }

    if (environments.length === 0) {
        return null;
    }

    return (
        <div className={cls.EnvironmentSwitcherContainer}>
            <select
                className={cls.Select}
                value={selectedEnvironment}
                onChange={handleChange}
                title="Active environment"
            >
                {environments.map((e) =>
                    <option key={e.id} value={e.name}>{e.name}</option>)}
            </select>
        </div>
    );
}

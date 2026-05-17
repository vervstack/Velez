import {useState} from 'react';

import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

import {EnableStatefullCluster} from '@/app/api/velez';
import {controlPlaneService} from "@/processes/api/control_plane.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";

export default function StatefullPgPluginForm() {
    const toaster = useToaster();

    const [exposePort, setExposePort] = useState(false);
    const [portNumber, setPortNumber] = useState('5432');

    function handleExposeChange(e: React.ChangeEvent<HTMLInputElement>) {
        setExposePort(e.target.checked);
    }

    function handlePortChange(e: React.ChangeEvent<HTMLInputElement>) {
        setPortNumber(e.target.value);
    }

    function handleEnable() {
        const payload: EnableStatefullCluster = {
            isExposePort: exposePort,
            exposeToPort: exposePort ? portNumber : undefined,
        };

        controlPlaneService.enableStatefullPgCluster(payload)
            .catch(toaster.catchGrpc)

    }

    return (
        <div className={cls.ActionSection}>
            <label className={cls.CheckboxLabel}>
                <input
                    type="checkbox"
                    checked={exposePort}
                    onChange={handleExposeChange}
                />
                <span>Expose port</span>
            </label>
            {exposePort && (
                <div className={cls.InputGroup}>
                    <label className={cls.InputLabel}>Port number:</label>
                    <input
                        type="text"
                        className={cls.PortInput}
                        value={portNumber}
                        onChange={handlePortChange}
                    />
                </div>
            )}
            <button
                className={cls.EnableButton}
                onClick={handleEnable}
            >
                Enable
            </button>
        </div>
    );
}

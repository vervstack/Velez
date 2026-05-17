import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

import { PluginFormProps } from '@/dialogs/PluginManageDialog/PluginManageDialog';

export default function SimplePluginForm({ isLoading, onEnable }: PluginFormProps) {
    function handleClick() {
        onEnable();
    }

    return (
        <div className={cls.ActionSection}>
            <button
                className={cls.EnableButton}
                onClick={handleClick}
                disabled={isLoading}
            >
                {isLoading ? 'Enabling...' : 'Enable'}
            </button>
        </div>
    );
}

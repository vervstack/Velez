import Choice from '@/components/base/Choice.tsx';
import {StatefullPgContext} from '@/dialogs/PluginManageDialog/plugins/StatefullPgContext.ts';
import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

interface StatefullPgSettingsScreenProps
    extends Pick<StatefullPgContext, 'exposePort' | 'portNumber' | 'isRunningInContainer'> {
    updateContext(partial: Partial<StatefullPgContext>): void;
}

export default function StatefullPgSettingsScreen(
    {exposePort, portNumber, isRunningInContainer, updateContext}: StatefullPgSettingsScreenProps) {

    function handlePortChange(e: React.ChangeEvent<HTMLInputElement>) {
        updateContext({portNumber: e.target.value});
    }

    function handleToggleExposePort() {
        updateContext({exposePort: !exposePort});
    }

    const portTooltip = isRunningInContainer
        ? undefined
        : "Running as a binary always exposes the port";

    return (
        <div className={cls.ActionSection}>
            <label
                className={cls.CheckboxLabel}
                data-tooltip-id={portTooltip ? "root-tooltip" : undefined}
                data-tooltip-content={portTooltip}
                data-tooltip-place="top"
            >
                <Choice title={'Expose port'}
                        active={exposePort}
                        disabled={!isRunningInContainer}
                        onClick={handleToggleExposePort}/>
            </label>
            {!isRunningInContainer && (
                <span>Velez is running as a binary on this node, so the port is always exposed.</span>
            )}

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
        </div>
    );
}

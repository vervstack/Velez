import React, { useState } from 'react';
import cls from '@/components/complex/PluginManageDialog/PluginManageDialog.module.css';
import IconButton from '@/components/base/IconButton';
import StatusDot from '@/components/base/StatusDot';
import { VervPluginType, VervPluginState, EnableStatefullCluster, EnableHeadscaleServer } from '@/app/api/velez';
import { useToaster } from '@/app/hooks/toaster/Toaster';

interface PluginManageDialogProps {
    isOpen: boolean;
    pluginName: string;
    pluginType: VervPluginType;
    pluginState: VervPluginState;
    onClose: () => void;
    onEnable: (type: VervPluginType, payload?: any) => Promise<void>;
}

function formatState(state?: VervPluginState): string {
    switch (state) {
        case VervPluginState.running: return 'Running';
        case VervPluginState.disabled: return 'Disabled';
        case VervPluginState.warning: return 'Warning';
        case VervPluginState.dead: return 'Dead';
        default: return 'Not installed';
    }
}

export default function PluginManageDialog({
    isOpen,
    pluginName,
    pluginType,
    pluginState,
    onClose,
    onEnable,
}: PluginManageDialogProps) {
    const [isLoading, setIsLoading] = useState(false);
    const [exposePort, setExposePort] = useState(false);
    const [portNumber, setPortNumber] = useState('5432');
    const [headscaleMode, setHeadscaleMode] = useState<'deploy' | 'external'>('deploy');
    const [customPort, setCustomPort] = useState('');
    const [customImage, setCustomImage] = useState('');
    const [externalUrl, setExternalUrl] = useState('');
    const [externalToken, setExternalToken] = useState('');

    const toaster = useToaster();

    if (!isOpen) return null;

    function handleOverlayClick() {
        onClose();
    }

    function handleContainerClick(e: React.MouseEvent) {
        e.stopPropagation();
    }

    function mapStateToStatus(state?: VervPluginState): 'running' | 'degraded' | 'stopped' | 'online' | 'offline' | 'enabled' | 'disabled' {
        switch (state) {
            case VervPluginState.running:
                return 'enabled';
            case VervPluginState.disabled:
                return 'disabled';
            case VervPluginState.warning:
            case VervPluginState.dead:
                return 'degraded';
            default:
                return 'disabled';
        }
    }

    async function handleSimpleEnable() {
        setIsLoading(true);
        try {
            await onEnable(pluginType);
            toaster.bake({ title: 'Success', description: `${pluginName} enabled`, level: 'Info' });
            onClose();
        } catch (error) {
            if (error instanceof Error) {
                toaster.catchGrpc(error);
            }
        } finally {
            setIsLoading(false);
        }
    }

    async function handleStatefullEnable() {
        setIsLoading(true);
        try {
            const payload: EnableStatefullCluster = {
                isExposePort: exposePort,
                exposeToPort: exposePort ? portNumber : undefined,
            };
            await onEnable(pluginType, payload);
            toaster.bake({ title: 'Success', description: `${pluginName} enabled`, level: 'Info' });
            onClose();
        } catch (error) {
            if (error instanceof Error) {
                toaster.catchGrpc(error);
            }
        } finally {
            setIsLoading(false);
        }
    }

    async function handleHeadscaleEnable() {
        setIsLoading(true);
        try {
            let payload: EnableHeadscaleServer;

            if (headscaleMode === 'deploy') {
                payload = {
                    deployConfig: {
                        customPort: customPort ? parseInt(customPort, 10) : undefined,
                        customImage: customImage || undefined,
                    },
                };
            } else {
                payload = {
                    externalConnect: {
                        url: externalUrl,
                        token: externalToken,
                    },
                };
            }

            await onEnable(pluginType, payload);
            toaster.bake({ title: 'Success', description: `${pluginName} enabled`, level: 'Info' });
            onClose();
        } catch (error) {
            if (error instanceof Error) {
                toaster.catchGrpc(error);
            }
        } finally {
            setIsLoading(false);
        }
    }

    const isRunning = pluginState === VervPluginState.running;
    const isSimplePlugin = [
        VervPluginType.matreshka,
        VervPluginType.makosh,
        VervPluginType.webserver,
        VervPluginType.portainer,
    ].includes(pluginType);

    return (
        <div className={cls.ModalOverlay} onClick={handleOverlayClick}>
            <div className={cls.ModalContainer} onClick={handleContainerClick}>
                <div className={cls.ModalHeader}>
                    <h2 className={cls.ModalTitle}>Manage {pluginName}</h2>
                    <IconButton label="×" onClick={onClose} title="Close" />
                </div>
                <div className={cls.ModalContent}>
                    <div className={cls.StatusSection}>
                        <span className={cls.StatusLabel}>Status:</span>
                        <div className={cls.StatusRow}>
                            <StatusDot status={mapStateToStatus(pluginState)} />
                            <span className={cls.StatusText}>{formatState(pluginState)}</span>
                        </div>
                    </div>

                    {!isRunning ? (
                        <>
                            {isSimplePlugin && (
                                <div className={cls.ActionSection}>
                                    <button
                                        className={cls.EnableButton}
                                        onClick={handleSimpleEnable}
                                        disabled={isLoading}
                                    >
                                        {isLoading ? 'Enabling...' : 'Enable'}
                                    </button>
                                </div>
                            )}

                            {pluginType === VervPluginType.statefull_pg && (
                                <div className={cls.ActionSection}>
                                    <label className={cls.CheckboxLabel}>
                                        <input
                                            type="checkbox"
                                            checked={exposePort}
                                            onChange={(e) => setExposePort(e.target.checked)}
                                            disabled={isLoading}
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
                                                onChange={(e) => setPortNumber(e.target.value)}
                                                disabled={isLoading}
                                            />
                                        </div>
                                    )}
                                    <button
                                        className={cls.EnableButton}
                                        onClick={handleStatefullEnable}
                                        disabled={isLoading}
                                    >
                                        {isLoading ? 'Enabling...' : 'Enable'}
                                    </button>
                                </div>
                            )}

                            {pluginType === VervPluginType.headscale && (
                                <div className={cls.ActionSection}>
                                    <div className={cls.RadioGroup}>
                                        <label className={cls.RadioLabel}>
                                            <input
                                                type="radio"
                                                value="deploy"
                                                checked={headscaleMode === 'deploy'}
                                                onChange={() => setHeadscaleMode('deploy')}
                                                disabled={isLoading}
                                            />
                                            <span>Deploy new</span>
                                        </label>
                                        <label className={cls.RadioLabel}>
                                            <input
                                                type="radio"
                                                value="external"
                                                checked={headscaleMode === 'external'}
                                                onChange={() => setHeadscaleMode('external')}
                                                disabled={isLoading}
                                            />
                                            <span>Connect to external</span>
                                        </label>
                                    </div>

                                    {headscaleMode === 'deploy' && (
                                        <div className={cls.ConfigSection}>
                                            <div className={cls.InputGroup}>
                                                <label className={cls.InputLabel}>Custom port (optional):</label>
                                                <input
                                                    type="text"
                                                    className={cls.ConfigInput}
                                                    value={customPort}
                                                    onChange={(e) => setCustomPort(e.target.value)}
                                                    placeholder="e.g., 8443"
                                                    disabled={isLoading}
                                                />
                                            </div>
                                            <div className={cls.InputGroup}>
                                                <label className={cls.InputLabel}>Custom image (optional):</label>
                                                <input
                                                    type="text"
                                                    className={cls.ConfigInput}
                                                    value={customImage}
                                                    onChange={(e) => setCustomImage(e.target.value)}
                                                    placeholder="e.g., juanfont/headscale:v0.22.0"
                                                    disabled={isLoading}
                                                />
                                            </div>
                                        </div>
                                    )}

                                    {headscaleMode === 'external' && (
                                        <div className={cls.ConfigSection}>
                                            <div className={cls.InputGroup}>
                                                <label className={cls.InputLabel}>Server URL:</label>
                                                <input
                                                    type="text"
                                                    className={cls.ConfigInput}
                                                    value={externalUrl}
                                                    onChange={(e) => setExternalUrl(e.target.value)}
                                                    placeholder="https://headscale.example.com"
                                                    disabled={isLoading}
                                                />
                                            </div>
                                            <div className={cls.InputGroup}>
                                                <label className={cls.InputLabel}>API Token:</label>
                                                <input
                                                    type="text"
                                                    className={cls.ConfigInput}
                                                    value={externalToken}
                                                    onChange={(e) => setExternalToken(e.target.value)}
                                                    placeholder="Your API token"
                                                    disabled={isLoading}
                                                />
                                            </div>
                                        </div>
                                    )}

                                    <button
                                        className={cls.EnableButton}
                                        onClick={handleHeadscaleEnable}
                                        disabled={isLoading}
                                    >
                                        {isLoading ? 'Enabling...' : 'Enable'}
                                    </button>
                                </div>
                            )}
                        </>
                    ) : (
                        <div className={cls.DisabledMessage}>Plugin is already running.</div>
                    )}
                </div>
            </div>
        </div>
    );
}

import { useState } from 'react';

import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

import { EnableHeadscaleServer } from '@/app/api/velez';
import { PluginFormProps } from '@/dialogs/PluginManageDialog/PluginManageDialog';

export default function HeadscalePluginForm({ isLoading, onEnable }: PluginFormProps) {
    const [headscaleMode, setHeadscaleMode] = useState<'deploy' | 'external'>('deploy');
    const [customPort, setCustomPort] = useState('');
    const [customImage, setCustomImage] = useState('');
    const [externalUrl, setExternalUrl] = useState('');
    const [externalToken, setExternalToken] = useState('');

    function handleSetDeploy() {
        setHeadscaleMode('deploy');
    }

    function handleSetExternal() {
        setHeadscaleMode('external');
    }

    function handleCustomPortChange(e: React.ChangeEvent<HTMLInputElement>) {
        setCustomPort(e.target.value);
    }

    function handleCustomImageChange(e: React.ChangeEvent<HTMLInputElement>) {
        setCustomImage(e.target.value);
    }

    function handleExternalUrlChange(e: React.ChangeEvent<HTMLInputElement>) {
        setExternalUrl(e.target.value);
    }

    function handleExternalTokenChange(e: React.ChangeEvent<HTMLInputElement>) {
        setExternalToken(e.target.value);
    }

    function handleEnable() {
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

        onEnable(payload);
    }

    return (
        <div className={cls.ActionSection}>
            <div className={cls.RadioGroup}>
                <label className={cls.RadioLabel}>
                    <input
                        type="radio"
                        value="deploy"
                        checked={headscaleMode === 'deploy'}
                        onChange={handleSetDeploy}
                        disabled={isLoading}
                    />
                    <span>Deploy new</span>
                </label>
                <label className={cls.RadioLabel}>
                    <input
                        type="radio"
                        value="external"
                        checked={headscaleMode === 'external'}
                        onChange={handleSetExternal}
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
                            onChange={handleCustomPortChange}
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
                            onChange={handleCustomImageChange}
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
                            onChange={handleExternalUrlChange}
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
                            onChange={handleExternalTokenChange}
                            placeholder="Your API token"
                            disabled={isLoading}
                        />
                    </div>
                </div>
            )}

            <button
                className={cls.EnableButton}
                onClick={handleEnable}
                disabled={isLoading}
            >
                {isLoading ? 'Enabling...' : 'Enable'}
            </button>
        </div>
    );
}

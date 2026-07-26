import { useState, useEffect } from 'react';

import cls from '@/widgets/settings/ApiSettings/ApiSettings.module.css';
import Input from '@/components/base/Input.tsx';
import StatusDot from '@/components/base/StatusDot.tsx';
import { useCredentialsStore } from '@/app/settings/creds.ts';
import { useToaster } from '@/app/hooks/toaster/Toaster.ts';
import { useConnectionHealth } from '@/app/hooks/health/useConnectionHealth.ts';

const DEFAULT_URL = import.meta.env.VITE_VELEZ_BACKEND_URL || 'http://0.0.0.0:53891';
const DEFAULT_AUTH = import.meta.env.VITE_VELEZ_AUTH_HEADER || '';

function isValidUrl(v: string): boolean {
    try {
        new URL(v);
        return true;
    } catch {
        return false;
    }
}

function mapHealthToDotStatus(status: 'checking' | 'healthy' | 'unhealthy'): 'pending' | 'healthy' | 'error' {
    if (status === 'healthy') return 'healthy';
    if (status === 'unhealthy') return 'error';
    return 'pending';
}

export default function ApiSettings() {
    const toaster = useToaster();
    const credStore = useCredentialsStore();
    const health = useConnectionHealth();

    const [url, setUrl] = useState(credStore.url || DEFAULT_URL);
    const [token, setToken] = useState(credStore.token || DEFAULT_AUTH);
    const [tokenVisible, setTokenVisible] = useState(false);

    useEffect(() => {
        setUrl(credStore.url || DEFAULT_URL);
        setToken(credStore.token || DEFAULT_AUTH);
    }, [credStore]);

    const urlValid = isValidUrl(url);
    const urlWarning = urlValid && url.startsWith('http://') ? 'Using HTTP — data is unencrypted' : null;
    const authWarning = !token ? 'Auth header is empty' : null;
    const canSave = urlValid;

    function handleSave() {
        if (!canSave) return;
        credStore.setUrl(url);
        credStore.setToken(token);
        toaster.bake({
            title: 'Settings saved',
            description: 'Your backend URL and auth token have been updated.',
            level: 'Info',
        });
        health.recheck();
    }

    function healthTooltip(): string {
        const checkedAt = health.lastCheckedAt ? health.lastCheckedAt.toLocaleTimeString() : 'never';
        if (health.status === 'checking') return 'Checking connection…';
        if (health.status === 'unhealthy') return `Unreachable as of ${checkedAt}${health.error ? ` — ${health.error}` : ''}`;
        return `Healthy as of ${checkedAt}`;
    }

    function handleReset() {
        setUrl(DEFAULT_URL);
        setToken(DEFAULT_AUTH);
    }

    function toggleTokenVisible() {
        setTokenVisible(v => !v);
    }

    return (
        <div className={cls.ApiSettingsContainer}>
            <div className={cls.SettingsSection}>
                <label className={cls.SettingsLabel}>Backend URL</label>
                <div className={cls.UrlRow}>
                    <Input onChange={setUrl} inputValue={url} />
                    <span
                        className={cls.HealthDotWrapper}
                        data-tooltip-id="root-tooltip"
                        data-tooltip-content={healthTooltip()}
                    >
                        <StatusDot status={mapHealthToDotStatus(health.status)} />
                    </span>
                </div>
                {!urlValid && url && <div className={cls.ValidationError}>Invalid URL</div>}
                {urlWarning && <div className={cls.ValidationWarn}>{urlWarning}</div>}
            </div>

            <div className={cls.SettingsSection}>
                <label className={cls.SettingsLabel}>Auth header</label>
                <div className={cls.TokenRow}>
                    <Input
                        onChange={setToken}
                        inputValue={tokenVisible ? token : (token ? '••••••••' : '')}
                        disabled={!tokenVisible}
                    />
                    <button className={cls.RevealButton} onClick={toggleTokenVisible}>
                        {tokenVisible ? 'Hide' : 'Show'}
                    </button>
                </div>
                {authWarning && <div className={cls.ValidationWarn}>{authWarning}</div>}
            </div>

            <div className={cls.EnvSection}>
                <div className={cls.EnvTitle}>Build-time env</div>
                <div className={cls.EnvRow}>
                    <span className={cls.EnvKey}>VITE_VELEZ_BACKEND_URL</span>
                    <span className={cls.EnvVal}>{import.meta.env.VITE_VELEZ_BACKEND_URL || '(not set)'}</span>
                </div>
                <div className={cls.EnvRow}>
                    <span className={cls.EnvKey}>VITE_VELEZ_AUTH_HEADER</span>
                    <span className={cls.EnvVal}>{import.meta.env.VITE_VELEZ_AUTH_HEADER ? '••••••' : '(not set)'}</span>
                </div>
            </div>

            <div className={cls.SettingsActions}>
                <button className={cls.ResetButton} onClick={handleReset}>Reset to defaults</button>
                <button className={cls.SaveButton} onClick={handleSave} disabled={!canSave}>Save</button>
            </div>
        </div>
    );
}

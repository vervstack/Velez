import React from 'react';

import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

import StatusDot from '@/components/base/StatusDot';

import {VervPluginState, VervPluginType} from '@/app/api/velez';

import SimplePluginForm from '@/dialogs/PluginManageDialog/plugins/SimplePluginForm';
import HeadscalePluginForm from '@/dialogs/PluginManageDialog/plugins/HeadscalePluginForm';
import {ListPluginsQuery} from "@/processes/queries/control_plane.ts";
import UnknownPlugin from "@/dialogs/PluginManageDialog/plugins/UnknownPlugin.tsx";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";

export default function PluginManageDialog({pluginType}: PluginManageDialogProps) {

    const PluginForm = pluginForms[pluginType] || UnknownPlugin;

    const pluginsQuery = ListPluginsQuery();

    const plugin = pluginsQuery.data?.find(p => p.type == pluginType) || new VervPlugin(VervPluginType.unknown_service_type, "")

    return (
        <div
            className={cls.ModalContainer}
        >

            <div className={cls.ModalHeader}>
                <h2 className={cls.ModalTitle}>Manage {pluginType}</h2>
            </div>

            <div className={cls.ModalContent}>
                <div className={cls.StatusSection}>
                    <span className={cls.StatusLabel}>Status:</span>
                    <div className={cls.StatusRow}>
                        <StatusDot status={mapStateToStatus(plugin?.state)}/>
                        <span className={cls.StatusText}>{formatState(plugin?.state)}</span>
                    </div>
                </div>

                {!pluginsQuery.isLoading && <PluginForm {...plugin}/>}
            </div>
        </div>
    );
}

const pluginForms: Partial<Record<VervPluginType, React.ComponentType<VervPlugin>>> = {
    [VervPluginType.headscale]: HeadscalePluginForm,

    [VervPluginType.matreshka]: SimplePluginForm,
    [VervPluginType.makosh]: SimplePluginForm,
    [VervPluginType.webserver]: SimplePluginForm,
    [VervPluginType.portainer]: SimplePluginForm,
};

interface PluginManageDialogProps {
    pluginType: VervPluginType;
}

function formatState(state?: VervPluginState): string {
    switch (state) {
        case VervPluginState.running:
            return 'Running';
        case VervPluginState.disabled:
            return 'Disabled';
        case VervPluginState.warning:
            return 'Warning';
        case VervPluginState.dead:
            return 'Dead';
        default:
            return 'Not installed';
    }
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


import {describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';

import PluginMatrix from '@/widgets/controlplane/PluginMatrix';
import {NodeBaseInfo, VervPluginState, VervPluginType} from '@/app/api/velez';
import {VervPlugin} from '@/model/services/VervPlugins.tsx';

vi.mock('react-router-dom', () => ({
    useNavigate: () => vi.fn(),
}));

function buildPlugin(type: VervPluginType, serviceName: string, state: VervPluginState): VervPlugin {
    const plugin = new VervPlugin(type, serviceName);
    plugin.state = state;
    return plugin;
}

const nodes: NodeBaseInfo[] = [{id: 'node-1', name: 'node-1'}];

describe('PluginMatrix', () => {
    it('renders rows sorted by status rank: running, then warning/dead, then disabled/unknown', () => {
        const disabledPlugin = buildPlugin(VervPluginType.matreshka, 'matreshka', VervPluginState.disabled);
        const runningPlugin = buildPlugin(VervPluginType.makosh, 'makosh', VervPluginState.running);
        const warningPlugin = buildPlugin(VervPluginType.headscale, 'headscale', VervPluginState.warning);

        render(
            <PluginMatrix
                nodes={nodes}
                plugins={[disabledPlugin, runningPlugin, warningPlugin]}
            />,
        );

        const titles = screen.getAllByText(/^(Matreshka|Makosh|Headscale)$/).map(el => el.textContent);

        expect(titles).toEqual(['Makosh', 'Headscale', 'Matreshka']);
    });

    it('does not throw when mounting with the FLIP ref/layout-effect code in jsdom', () => {
        const plugin = buildPlugin(VervPluginType.matreshka, 'matreshka', VervPluginState.running);

        expect(() => render(<PluginMatrix nodes={nodes} plugins={[plugin]}/>)).not.toThrow();
    });
});

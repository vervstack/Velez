import {describe, expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';

import NodeHealthList from '@/widgets/controlplane/NodeHealthList';
import {NodeBaseInfo, VervPluginType} from '@/app/api/velez';
import {VervPlugin} from '@/model/services/VervPlugins.tsx';

function buildPlugin(type: VervPluginType, serviceName: string, nodeIds: string[]): VervPlugin {
    const plugin = new VervPlugin(type, serviceName);
    plugin.nodeIds = nodeIds;
    return plugin;
}

const nodes: NodeBaseInfo[] = [
    {id: 'node-1', name: 'node-1'},
    {id: 'node-2', name: 'node-2'},
];

describe('NodeHealthList', () => {
    it('filters plugins per node by nodeIds membership', () => {
        const makosh = buildPlugin(VervPluginType.makosh, 'makosh', ['node-1']);
        const matreshka = buildPlugin(VervPluginType.matreshka, 'matreshka', ['node-2']);
        const headscale = buildPlugin(VervPluginType.headscale, 'headscale', ['node-1', 'node-2']);

        render(
            <NodeHealthList
                nodes={nodes}
                plugins={[makosh, matreshka, headscale]}
            />,
        );

        const node1Icons = screen.getAllByTitle('Makosh');
        expect(node1Icons).toHaveLength(1);

        const node2Icons = screen.getAllByTitle('Matreshka');
        expect(node2Icons).toHaveLength(1);

        const sharedIcons = screen.getAllByTitle('Headscale');
        expect(sharedIcons).toHaveLength(2);
    });

    it('renders no plugin icons for a node when no plugin lists it in nodeIds', () => {
        const orphanPlugin = buildPlugin(VervPluginType.matreshka, 'matreshka', ['node-99']);

        render(
            <NodeHealthList
                nodes={nodes}
                plugins={[orphanPlugin]}
            />,
        );

        expect(screen.queryByTitle('Matreshka')).not.toBeInTheDocument();
    });

    it('does not throw when node.id is undefined', () => {
        const nodesWithoutId: NodeBaseInfo[] = [{name: 'node-x'}];
        const plugin = buildPlugin(VervPluginType.makosh, 'makosh', []);

        expect(() =>
            render(<NodeHealthList nodes={nodesWithoutId} plugins={[plugin]}/>),
        ).not.toThrow();
    });
});

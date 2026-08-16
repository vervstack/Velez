import {describe, expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';

import NodeTopologyGraph from '@/widgets/controlplane/NodeTopologyGraph';
import {NodeBaseInfo, NodeStatus, VervPluginType} from '@/app/api/velez';
import {VervPlugin} from '@/model/services/VervPlugins';

function buildNode(id: string, status: NodeStatus = NodeStatus.NodeStatus_Online): NodeBaseInfo {
    return {id, name: id, status};
}

function buildStatefullPlugin(masterNodeId?: string): VervPlugin {
    const plugin = new VervPlugin(VervPluginType.statefull_pg, 'statefull_pg');
    plugin.masterNodeId = masterNodeId;
    return plugin;
}

const nodes: NodeBaseInfo[] = [
    buildNode('node-1'),
    buildNode('node-2'),
    buildNode('node-3'),
    buildNode('node-4'),
];

describe('NodeTopologyGraph', () => {
    it('renders one node group per node (center + all outer nodes)', () => {
        const plugins = [buildStatefullPlugin('node-1')];

        render(<NodeTopologyGraph nodes={nodes} plugins={plugins}/>);

        const nodeGroups = screen.getAllByTestId('topology-node');
        expect(nodeGroups).toHaveLength(nodes.length);
    });

    it('picks the master node as the center when masterNodeId is set', () => {
        const plugins = [buildStatefullPlugin('node-3')];

        render(<NodeTopologyGraph nodes={nodes} plugins={plugins}/>);

        const centerGroups = screen.getAllByTestId('topology-node')
            .filter(g => g.getAttribute('data-center') === 'true');
        expect(centerGroups).toHaveLength(1);
        expect(centerGroups[0].textContent).toContain('node-3');
    });

    it('falls back to the first node as the center when no master node is found', () => {
        const plugins: VervPlugin[] = [];

        render(<NodeTopologyGraph nodes={nodes} plugins={plugins}/>);

        const centerGroups = screen.getAllByTestId('topology-node')
            .filter(g => g.getAttribute('data-center') === 'true');
        expect(centerGroups).toHaveLength(1);
        expect(centerGroups[0].textContent).toContain('node-1');

        // all nodes should still be rendered, none lost/duplicated
        expect(screen.getAllByTestId('topology-node')).toHaveLength(nodes.length);
    });

    it('does not throw when masterNodeId points to a node not present in nodes', () => {
        const plugins = [buildStatefullPlugin('missing-node')];

        expect(() => render(<NodeTopologyGraph nodes={nodes} plugins={plugins}/>)).not.toThrow();
    });
});

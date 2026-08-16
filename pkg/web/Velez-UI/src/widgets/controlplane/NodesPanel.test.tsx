import {describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';

import NodesPanel from '@/widgets/controlplane/NodesPanel';
import {NodeBaseInfo, NodeStatus} from '@/app/api/velez';
import {VervPlugin} from '@/model/services/VervPlugins';

vi.mock('react-router-dom', () => ({
    useNavigate: () => vi.fn(),
}));

function buildNode(id: string): NodeBaseInfo {
    return {id, name: id, status: NodeStatus.NodeStatus_Online};
}

const plugins: VervPlugin[] = [];

describe('NodesPanel', () => {
    it('renders the card grid (NodeHealthList) for exactly 3 nodes', () => {
        const nodes = [buildNode('node-1'), buildNode('node-2'), buildNode('node-3')];

        render(<NodesPanel nodes={nodes} plugins={plugins}/>);

        expect(screen.getByText('Node Health')).toBeInTheDocument();
        expect(screen.queryByText('Node Topology')).not.toBeInTheDocument();
    });

    it('renders the card grid for fewer than 3 nodes', () => {
        const nodes = [buildNode('node-1')];

        render(<NodesPanel nodes={nodes} plugins={plugins}/>);

        expect(screen.getByText('Node Health')).toBeInTheDocument();
    });

    it('renders the topology graph for exactly 4 nodes', () => {
        const nodes = [buildNode('node-1'), buildNode('node-2'), buildNode('node-3'), buildNode('node-4')];

        render(<NodesPanel nodes={nodes} plugins={plugins}/>);

        expect(screen.getByText('Node Topology')).toBeInTheDocument();
        expect(screen.queryByText('Node Health')).not.toBeInTheDocument();
    });

    it('renders the topology graph for more than 4 nodes', () => {
        const nodes = [
            buildNode('node-1'), buildNode('node-2'), buildNode('node-3'),
            buildNode('node-4'), buildNode('node-5'),
        ];

        render(<NodesPanel nodes={nodes} plugins={plugins}/>);

        expect(screen.getByText('Node Topology')).toBeInTheDocument();
    });
});

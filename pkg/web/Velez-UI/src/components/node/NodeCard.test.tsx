import {describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';

import NodeCard from '@/components/node/NodeCard';
import {NodeBaseInfo, VervPluginType} from '@/app/api/velez';
import {VervPlugin} from '@/model/services/VervPlugins.tsx';

const node: NodeBaseInfo = {id: 'node-1', name: 'node-1'};

function buildPlugin(type: VervPluginType, serviceName: string): VervPlugin {
    return new VervPlugin(type, serviceName);
}

describe('NodeCard', () => {
    it('renders plugin icons when plugins is non-empty', () => {
        const plugins = [
            buildPlugin(VervPluginType.makosh, 'makosh'),
            buildPlugin(VervPluginType.matreshka, 'matreshka'),
        ];

        render(<NodeCard node={node} plugins={plugins}/>);

        expect(screen.getByTitle('Makosh')).toBeInTheDocument();
        expect(screen.getByTitle('Matreshka')).toBeInTheDocument();
    });

    it('renders nothing extra when plugins is empty', () => {
        const {container} = render(<NodeCard node={node} plugins={[]}/>);

        expect(container.querySelectorAll('img[title]').length).toBe(0);
    });

    it('does not throw when mounting with plugins omitted-equivalent empty array', () => {
        expect(() => render(<NodeCard node={node} plugins={[]} onShell={vi.fn()} onDrain={vi.fn()}/>)).not.toThrow();
    });
});

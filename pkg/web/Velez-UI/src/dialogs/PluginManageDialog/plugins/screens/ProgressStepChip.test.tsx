import {describe, expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';

import {TaskStatusStatus} from '@/app/api/velez';
import ProgressStepChip from '@/dialogs/PluginManageDialog/plugins/screens/ProgressStepChip.tsx';

describe('ProgressStepChip', () => {
    it('renders a done step with the done modifier and no shimmer', () => {
        const {container} = render(
            <ProgressStepChip index={1} name="generate_credentials" status={TaskStatusStatus.DONE}/>
        );

        expect(screen.getByText('01')).toBeInTheDocument();
        expect(screen.getByText('Generating credentials')).toBeInTheDocument();
        expect(container.querySelector('[class*="ProgressStepChipContainer"]')?.className).toMatch(/done/);
        expect(container.querySelector('[class*="Shimmer"]')).toBeNull();
    });

    it('renders a running step with the running modifier and a shimmer overlay', () => {
        const {container} = render(
            <ProgressStepChip index={3} name="start_container" status={TaskStatusStatus.RUNNING}/>
        );

        expect(screen.getByText('03')).toBeInTheDocument();
        expect(screen.getByText('Starting container')).toBeInTheDocument();
        expect(container.querySelector('[class*="ProgressStepChipContainer"]')?.className).toMatch(/running/);
        expect(container.querySelector('[class*="Shimmer"]')).not.toBeNull();
    });

    it('renders a pending step as dim/future with no shimmer', () => {
        const {container} = render(
            <ProgressStepChip index={4} name="wait_for_postgres_ready" status={TaskStatusStatus.PENDING}/>
        );

        expect(screen.getByText('Waiting for Postgres')).toBeInTheDocument();
        const chip = container.querySelector('[class*="ProgressStepChipContainer"]');
        expect(chip?.className).not.toMatch(/done|running|failed/);
        expect(container.querySelector('[class*="Shimmer"]')).toBeNull();
    });

    it('renders a failed step with the failed modifier and a red-tinted border class', () => {
        const {container} = render(
            <ProgressStepChip index={2} name="create_container" status={TaskStatusStatus.FAILED}/>
        );

        expect(screen.getByText('Creating container')).toBeInTheDocument();
        expect(container.querySelector('[class*="ProgressStepChipContainer"]')?.className).toMatch(/failed/);
        expect(container.querySelector('[class*="Shimmer"]')).toBeNull();
    });

    it('falls back to a humanized label for unmapped job names', () => {
        render(<ProgressStepChip index={5} name="some_new_future_step" status={TaskStatusStatus.PENDING}/>);

        expect(screen.getByText('Some New Future Step')).toBeInTheDocument();
    });
});

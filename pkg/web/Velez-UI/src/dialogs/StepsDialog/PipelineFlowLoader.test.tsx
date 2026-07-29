import {describe, expect, it} from 'vitest';
import {render} from '@testing-library/react';

import PipelineFlowLoader from '@/dialogs/StepsDialog/PipelineFlowLoader.tsx';

describe('PipelineFlowLoader', () => {
    it('renders without throwing', () => {
        expect(() => render(<PipelineFlowLoader/>)).not.toThrow();
    });

    it('renders an svg', () => {
        const {container} = render(<PipelineFlowLoader/>);

        expect(container.querySelector('svg')).toBeInTheDocument();
    });

    it.each([
        'pipeline-server-bottom',
        'pipeline-server-top-left',
        'pipeline-server-top-right',
        'pipeline-database',
        'pipeline-connector-s1-db',
        'pipeline-connector-db-s2',
        'pipeline-connector-db-s3',
    ])('renders element with data-testid="%s"', (testId) => {
        const {getByTestId} = render(<PipelineFlowLoader/>);

        expect(getByTestId(testId)).toBeInTheDocument();
    });
});

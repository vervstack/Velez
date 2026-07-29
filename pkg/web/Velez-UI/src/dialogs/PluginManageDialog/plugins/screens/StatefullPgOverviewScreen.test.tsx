import {describe, expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';

import StatefullPgOverviewScreen from '@/dialogs/PluginManageDialog/plugins/screens/StatefullPgOverviewScreen.tsx';

describe('StatefullPgOverviewScreen', () => {
    it('renders the pipeline animation alongside the explanatory message', () => {
        render(<StatefullPgOverviewScreen/>);

        expect(screen.getByTestId('pipeline-database')).toBeInTheDocument();
        expect(screen.getByText(/statefull_pg is currently disabled/)).toBeInTheDocument();
    });
});

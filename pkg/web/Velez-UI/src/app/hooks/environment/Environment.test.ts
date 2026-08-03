import {beforeEach, describe, expect, it} from 'vitest';

import {useEnvironmentStore} from '@/app/hooks/environment/Environment.ts';
import type {Environment} from '@/app/api/velez';

function env(name: string, id = name): Environment {
    return {id, name} as Environment;
}

beforeEach(() => {
    useEnvironmentStore.setState({environments: [], selectedEnvironment: ''});
});

describe('useEnvironmentStore', () => {
    it('defaults the selection to PROD when present in the fetched list', () => {
        useEnvironmentStore.getState().setEnvironments([env('STAGING'), env('PROD'), env('DEV')]);

        expect(useEnvironmentStore.getState().selectedEnvironment).toBe('PROD');
    });

    it('falls back to the first environment when PROD is absent', () => {
        useEnvironmentStore.getState().setEnvironments([env('STAGING'), env('DEV')]);

        expect(useEnvironmentStore.getState().selectedEnvironment).toBe('STAGING');
    });

    it('leaves selection empty when the fetched list is empty', () => {
        useEnvironmentStore.getState().setEnvironments([]);

        expect(useEnvironmentStore.getState().selectedEnvironment).toBe('');
    });

    it('keeps the current selection if it is still present on refetch', () => {
        useEnvironmentStore.getState().setEnvironments([env('STAGING'), env('PROD')]);
        useEnvironmentStore.getState().selectEnvironment('STAGING');

        useEnvironmentStore.getState().setEnvironments([env('STAGING'), env('PROD'), env('DEV')]);

        expect(useEnvironmentStore.getState().selectedEnvironment).toBe('STAGING');
    });

    it('re-defaults selection if the previously selected environment disappears', () => {
        useEnvironmentStore.getState().setEnvironments([env('STAGING'), env('PROD')]);
        useEnvironmentStore.getState().selectEnvironment('STAGING');

        useEnvironmentStore.getState().setEnvironments([env('PROD'), env('DEV')]);

        expect(useEnvironmentStore.getState().selectedEnvironment).toBe('PROD');
    });

    it('selectEnvironment sets the selection explicitly', () => {
        useEnvironmentStore.getState().setEnvironments([env('STAGING'), env('PROD')]);
        useEnvironmentStore.getState().selectEnvironment('STAGING');

        expect(useEnvironmentStore.getState().selectedEnvironment).toBe('STAGING');
    });
});

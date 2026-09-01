import {describe, expect, it} from 'vitest';

import {deriveAppName} from '@/dialogs/CreateAppDialog/processes/deriveAppName.ts';

describe('deriveAppName', () => {
    it('derives the name from an https URL with a .git suffix', () => {
        expect(deriveAppName('https://github.com/org/repo.git')).toBe('repo');
    });

    it('derives the name from an https URL without a .git suffix', () => {
        expect(deriveAppName('https://github.com/org/repo')).toBe('repo');
    });

    it('derives the name from an SCP-style git@ URL', () => {
        expect(deriveAppName('git@github.com:org/repo.git')).toBe('repo');
    });

    it('ignores a trailing slash', () => {
        expect(deriveAppName('https://github.com/org/repo/')).toBe('repo');
    });

    it('returns an empty string for an empty input', () => {
        expect(deriveAppName('')).toBe('');
    });

    it('returns an empty string for a garbage string with no slash', () => {
        expect(deriveAppName('notaurl')).toBe('');
    });
});

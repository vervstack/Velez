export function deriveAppName(gitUrl: string): string {
    const trimmed = gitUrl.trim();
    if (!trimmed) return '';

    const withoutTrailingSlashes = trimmed.replace(/\/+$/, '');
    if (!withoutTrailingSlashes) return '';

    const hasSlash = withoutTrailingSlashes.includes('/');
    const hasColon = withoutTrailingSlashes.includes(':');
    if (!hasSlash && !hasColon) return '';

    const afterSlash = withoutTrailingSlashes.split('/').pop() || '';
    const lastSegment = afterSlash.includes(':') ? (afterSlash.split(':').pop() || '') : afterSlash;

    return lastSegment.replace(/\.git$/, '');
}

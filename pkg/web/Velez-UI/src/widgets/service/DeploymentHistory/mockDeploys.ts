export interface MockDeploy {
    id: number
    version: string
    date: string
    who: string
    git: string
    branch: string
    tag: string | null
    env: string
    status: 'success' | 'failed' | 'rolled'
    current?: boolean
}

export const MOCK_DEPLOYS: MockDeploy[] = [
    { id: 147, version: '0.4.9-rc.2', date: '2026-05-08 09:14', who: 'r.popov',     git: 'a3f8c12', branch: 'feat/runtime-push', tag: null,      env: 'dev',     status: 'success', current: true  },
    { id: 146, version: '0.4.9-rc.1', date: '2026-05-07 17:02', who: 'r.popov',     git: 'b1d2e44', branch: 'feat/runtime-push', tag: null,      env: 'dev',     status: 'success' },
    { id: 145, version: '0.4.8',      date: '2026-05-06 11:48', who: 'a.kuznetsov', git: 'c8e0991', branch: 'main',               tag: 'v0.4.8',  env: 'test',    status: 'success' },
    { id: 144, version: '0.4.8-rc.3', date: '2026-05-05 15:21', who: 'a.kuznetsov', git: 'd72b330', branch: 'main',               tag: null,      env: 'test',    status: 'success' },
    { id: 143, version: '0.4.7',      date: '2026-04-24 10:05', who: 'r.popov',     git: 'e09a1c5', branch: 'main',               tag: 'v0.4.7',  env: 'prod',    status: 'success', current: true  },
    { id: 142, version: '0.4.7-rc.4', date: '2026-04-22 18:33', who: 'r.popov',     git: 'f4b8217', branch: 'main',               tag: null,      env: 'staging', status: 'success' },
    { id: 141, version: '0.4.7-rc.3', date: '2026-04-22 14:12', who: 'r.popov',     git: 'a991f02', branch: 'main',               tag: null,      env: 'staging', status: 'failed'  },
    { id: 140, version: '0.4.6',      date: '2026-04-12 08:55', who: 'a.kuznetsov', git: '12fa804', branch: 'main',               tag: 'v0.4.6',  env: 'prod',    status: 'success' },
    { id: 139, version: '0.4.5',      date: '2026-04-04 13:40', who: 'm.lebedev',   git: '77ac01a', branch: 'main',               tag: 'v0.4.5',  env: 'prod',    status: 'success' },
    { id: 138, version: '0.4.4',      date: '2026-03-29 09:18', who: 'r.popov',     git: 'c0fde11', branch: 'main',               tag: 'v0.4.4',  env: 'prod',    status: 'success' },
    { id: 137, version: '0.4.3',      date: '2026-03-22 17:00', who: 'a.kuznetsov', git: '59bd8e2', branch: 'main',               tag: 'v0.4.3',  env: 'prod',    status: 'rolled'  },
    { id: 136, version: '0.4.2',      date: '2026-03-15 10:44', who: 'm.lebedev',   git: '4e22fa8', branch: 'main',               tag: 'v0.4.2',  env: 'prod',    status: 'success' },
]

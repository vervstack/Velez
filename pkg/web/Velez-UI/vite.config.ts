import {fileURLToPath, URL} from 'node:url'

import {defineConfig, loadEnv} from 'vite'
import react from '@vitejs/plugin-react'

export default ({mode}: { mode: string }) => {
    process.env = {...process.env, ...loadEnv(mode, process.cwd())};

    return defineConfig({
        base: '/',
        plugins: [react()],

        build: {
            rollupOptions: {
                output: {
                    manualChunks(id) {
                        if (id.includes('node_modules')) {
                            if (id.includes('react-dom') || id.includes('react-router')) {
                                return 'vendor-react';
                            }
                            if (id.includes('@tanstack')) {
                                return 'vendor-query';
                            }
                            return 'vendor';
                        }
                    },
                },
            },
        },

        resolve: {
            alias: {
                '@': fileURLToPath(new URL('./src', import.meta.url)),
            },
        },
    });
}

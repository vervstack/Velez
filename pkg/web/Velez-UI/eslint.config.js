import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import reactPlugin from 'eslint-plugin-react'
import importX from 'eslint-plugin-import-x'
import {createTypeScriptImportResolver} from 'eslint-import-resolver-typescript'
import tseslint from 'typescript-eslint'
import {globalIgnores} from 'eslint/config'

const localPlugin = {
    rules: {
        'no-relative-imports': {
            meta: {type: 'suggestion'},
            create(context) {
                return {
                    ImportDeclaration(node) {
                        const src = node.source.value
                        if (src.startsWith('./') || src.startsWith('../')) {
                            context.report({
                                node,
                                message: "Use '@/' path alias instead of relative imports.",
                            })
                        }
                    },
                }
            },
        },
    },
}

export default tseslint.config([
    globalIgnores(['dist', '**/*.pb.ts']),
    {
        files: ['**/*.{ts,tsx}'],
        extends: [
            js.configs.recommended,
            tseslint.configs.recommended,
            reactHooks.configs['recommended-latest'],
            reactRefresh.configs.vite,
        ],
        plugins: {
            react: reactPlugin,
            'import-x': importX,
            local: localPlugin,
        },
        languageOptions: {
            ecmaVersion: 2020,
            globals: globals.browser,
        },
        settings: {
            'import-x/resolver-next': [
                createTypeScriptImportResolver(),
            ],
            react: {
                version: 'detect',
            },
        },
        rules: {
            '@typescript-eslint/no-empty-object-type': 'off',
            '@typescript-eslint/ban-ts-comment': 'off',
            'react-refresh/only-export-components': 'off',
            '@typescript-eslint/no-unused-vars': ['error', {argsIgnorePattern: '^_'}],

            'react-hooks/exhaustive-deps': 'off',
            'import-x/no-cycle': 'error',
            'no-empty': ['error', {allowEmptyCatch: true}],

            // 'func-style' baseline: 5 pre-existing occurrences (const-assigned arrow functions used as
            // handlers/helpers in components/base/Checkbox.tsx, components/base/Input.tsx,
            // widgets/deploy/DeployWidget.tsx) as of this rule's introduction — fix the ones you touch rather
            // than sweeping the whole codebase.
            'func-style': ['warn', 'declaration', {allowArrowFunctions: false}],

            // 'react/no-multi-comp' baseline: 44 pre-existing occurrences (several widgets/pages colocate small
            // helper sub-components in the same file) as of this rule's introduction — fix the ones you touch
            // rather than sweeping the whole codebase.
            'react/no-multi-comp': ['warn', {ignoreStateless: false}],

            // 'react/jsx-max-depth' baseline: 56 pre-existing occurrences as of this rule's introduction — fix
            // the ones you touch rather than sweeping the whole codebase.
            'react/jsx-max-depth': ['warn', {max: 3}],

            // 'react/forbid-component-props' baseline: 2 pre-existing occurrences (components/base/Search.tsx
            // passes style to a child component) as of this rule's introduction — fix the ones you touch rather
            // than sweeping the whole codebase.
            'react/forbid-component-props': ['warn', {forbid: ['style']}],

            // 'max-len' baseline: 27 pre-existing occurrences across 10 files as of this rule's introduction —
            // fix the ones you touch rather than sweeping the whole codebase.
            'max-len': ['warn', {code: 120, ignoreUrls: true, ignoreStrings: false, ignoreTemplateLiterals: false}],

            // 'local/no-relative-imports' baseline: 4 pre-existing occurrences (widgets/service/ResourcesSection/*,
            // widgets/service/ServiceGraph/ServiceGraph.tsx) as of this rule's introduction — fix the ones you
            // touch rather than sweeping the whole codebase.
            'local/no-relative-imports': 'warn',

            // 'no-console' baseline: 5 pre-existing console.log calls as of this rule's introduction — fix the
            // ones you touch rather than sweeping the whole codebase.
            'no-console': ['warn', {allow: ['warn', 'error']}],

            // 'max-lines' baseline: 4 pre-existing files over 250 lines (processes/api/service.ts,
            // pages/home/HomePage.tsx, pages/service/ServiceInfoPage.tsx, widgets/sidebar/Sidebar.tsx) as of this
            // rule's introduction — fix the ones you touch rather than sweeping the whole codebase.
            'max-lines': ['warn', {max: 250, skipBlankLines: true, skipComments: true}],

            // 'max-lines-per-function' baseline: 6 pre-existing occurrences as of this rule's introduction — fix
            // the ones you touch rather than sweeping the whole codebase.
            'max-lines-per-function': ['warn', {max: 100, skipBlankLines: true, skipComments: true, IIFEs: true}],

            // This single rule name can only carry one severity across the whole config, but its selector list
            // spans both clean-today selectors (div-needs-className, no zIndex Property) and selectors with a
            // real pre-existing baseline (template-literal className, raw button/input, window.alert/confirm).
            // ESLint flat config replaces (doesn't merge) an array-valued rule across matching config objects for
            // the same file, so the clean selectors can't be split into a separate 'error' block without either
            // duplicating the whole array or losing coverage. Setting the whole rule to 'warn' is a deliberate
            // trade-off: the clean selectors just never fire, so nothing is lost by their being nominally 'warn'.
            // Baseline as of this rule's introduction (real linter counts): raw <button> 32, div-without-className
            // 23, raw <input> 15, template-literal className 13, ObjectPattern>6 5, window.alert/confirm 1
            // (src/pages/service/parts/DeployMenu.tsx), zIndex 0 — fix the ones you touch rather than sweeping the
            // whole codebase.
            'no-restricted-syntax': [
                'warn',
                {
                    selector: 'JSXOpeningElement[name.name="div"]:not(:has(JSXAttribute[name.name="className"]))',
                    message: '<div> must have a className (use a CSS module class).',
                },
                {
                    selector: 'Property[key.name="zIndex"]',
                    message: 'Never use z-index — rely on DOM order instead (see pkg/web/Velez-UI/CLAUDE.md).',
                },
                {
                    selector: 'JSXAttribute[name.name="className"] > JSXExpressionContainer > TemplateLiteral',
                    message: 'Do not build className with a template literal — use cn() from \'classnames\' instead.',
                },
                {
                    selector: 'JSXOpeningElement[name.name="button"]',
                    message: 'Never use a raw <button> — use Button/ActionButton/IconButton from components/base.',
                },
                {
                    selector: 'JSXOpeningElement[name.name="input"]',
                    message: 'Never use a raw <input> — use Input/TextInput/Search/Checkbox/Choice from components/base.',
                },
                {
                    selector: 'CallExpression[callee.object.name="window"][callee.property.name=/^(alert|confirm)$/]',
                    message: 'Never use window.alert/window.confirm — use useToaster() (see pkg/web/Velez-UI/CLAUDE.md).',
                },
                {
                    selector: 'ObjectPattern[properties.length>6]',
                    message: 'Destructuring more than 6 properties — keep the whole object as one variable (e.g. `const context = useX(...)`) instead of exploding it into separate bindings/props.',
                },
            ],

            // 'import-x/order' baseline: 148 pre-existing occurrences (grouping/ordering was never enforced
            // before, so most files need at least one blank-line or reorder fix) as of this rule's introduction —
            // fix the ones you touch rather than sweeping the whole codebase.
            'import-x/order': ['warn', {
                groups: [
                    ['builtin', 'external'],
                    ['internal'],
                    ['parent', 'sibling', 'index'],
                    ['unknown'],
                ],
                pathGroups: [
                    {
                        pattern: 'react|react-dom|react-router-dom',
                        group: 'builtin',
                        position: 'before',
                    },
                    {
                        pattern: '@/**',
                        group: 'internal',
                        position: 'before',
                    },
                    {
                        pattern: '*.{css}',
                        patternOptions: {matchBase: true},
                        group: 'index',
                        position: 'after',
                    },
                ],
                'newlines-between': 'always',
            }],

            // Enforce the layer order from CLAUDE.md: pages > widgets > segments > components > processes/api >
            // model > app. dialogs/ must not import from pages (pages open dialogs, never the other way around).
            'import-x/no-restricted-paths': ['error', {
                basePath: './src',
                zones: [
                    {
                        target: './components',
                        from: ['./widgets', './segments', './pages'],
                        message: 'src/components/** is below widgets/segments/pages — it may not import from them.',
                    },
                    {
                        target: ['./widgets', './segments'],
                        from: ['./pages'],
                        message: 'src/widgets/** and src/segments/** are below pages — they may not import from them.',
                    },
                    {
                        target: './dialogs',
                        from: ['./pages'],
                        message: 'Dialogs must not import from pages — pages open dialogs, never the other way around.',
                    },
                ],
            }],

            // gRPC clients must only be used inside src/processes/ (type-only imports via `import type` are
            // allowed anywhere — this codebase mostly imports proto types as plain named imports, so this also
            // catches type-only usages of generated types outside processes/).
            // Baseline: 26 pre-existing files import from '@/app/api/velez' directly (mostly type usages of
            // generated enums/interfaces, not just live gRPC calls) as of this rule's introduction — fix the
            // ones you touch rather than sweeping the whole codebase.
            '@typescript-eslint/no-restricted-imports': ['warn', {
                patterns: [{
                    group: ['@/app/api/**'],
                    message: 'gRPC clients must only be called from src/processes/. Use a process function instead.',
                    allowTypeImports: true,
                }],
            }],
        },
    },
    {
        files: ['src/processes/**/*.{ts,tsx}'],
        rules: {
            '@typescript-eslint/no-restricted-imports': 'off',
        },
    },
])

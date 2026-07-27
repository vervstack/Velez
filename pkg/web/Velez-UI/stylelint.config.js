export default {
    ignoreFiles: ['dist/**'],
    extends: [
        'stylelint-config-standard',
        'stylelint-config-css-modules',
    ],
    rules: {
        // No named colors — always use CSS variables. Warning only: 12 pre-existing occurrences as of this
        // rule's introduction — fix the ones you touch rather than sweeping the whole codebase.
        'color-named': ['never', {severity: 'warning'}],

        // No !important
        'declaration-no-important': true,

        // Allow blank lines between property groups (common grouping pattern)
        'declaration-empty-line-before': null,

        // Don't require shorthand when longhand is clearer
        'declaration-block-no-redundant-longhand-properties': null,

        // Allow single-line keyframe blocks (common shorthand)
        'declaration-block-single-line-max-declarations': null,

        // Allow camelCase keyframe names (e.g. slideIn, fadeOut)
        'keyframes-name-pattern': null,

        // Vendor prefixes still needed for backdrop-filter etc.
        'property-no-vendor-prefix': null,

        // Allow more than 4 decimal places in calculated values
        'number-max-precision': null,

        // Sizes must be expressed in rem — no px/em/vh/vw/etc. Warning only: 355 pre-existing occurrences as of
        // this rule's introduction — fix the ones you touch rather than sweeping the whole codebase.
        'unit-disallowed-list': [
            ['px', 'em', 'ex', 'ch', 'vh', 'vw', 'vmin', 'vmax', 'cm', 'mm', 'in', 'pt', 'pc', 'q'],
            {severity: 'warning'},
        ],

        // The rules below come from stylelint-config-standard/stylelint-config-recommended (via `extends`
        // above), not something added specifically for this project. Enabling those base configs surfaced real
        // pre-existing violations that were never linted before, so — consistent with the warn/error framework
        // used throughout this config — each is downgraded to 'warning' with its real baseline count rather than
        // silently fixed; the rule's default primary/secondary options are preserved, only severity changes.
        // Fix the ones you touch rather than sweeping the whole codebase.
        'color-function-notation': ['modern', {severity: 'warning'}], // 63
        'color-function-alias-notation': ['without-alpha', {severity: 'warning'}], // 63
        'alpha-value-notation': [
            'percentage',
            {
                exceptProperties: ['opacity', 'fill-opacity', 'flood-opacity', 'stop-opacity', 'stroke-opacity'],
                severity: 'warning',
            },
        ], // 63
        'block-no-empty': [true, {severity: 'warning'}], // 9
        'rule-empty-line-before': [
            'always-multi-line',
            {except: ['first-nested'], ignore: ['after-comment'], severity: 'warning'},
        ], // 8
        'value-keyword-case': ['lower', {severity: 'warning'}], // 6
        'shorthand-property-no-redundant-values': [true, {severity: 'warning'}], // 3
        'no-descending-specificity': [true, {severity: 'warning'}], // 3
        'media-feature-range-notation': ['context', {severity: 'warning'}], // 3
        'font-family-name-quotes': ['always-where-recommended', {severity: 'warning'}], // 3
        'custom-property-empty-line-before': [
            'always',
            {
                except: ['after-custom-property', 'first-nested'],
                ignore: ['after-comment', 'inside-single-line-block'],
                severity: 'warning'
            },
        ], // 2
        'selector-type-no-unknown': [true, {ignore: ['custom-elements'], severity: 'warning'}], // 1
        'declaration-block-no-shorthand-property-overrides': [true, {severity: 'warning'}], // 1
        'color-hex-length': ['short', {severity: 'warning'}], // 1
    },
    overrides: [
        // Component CSS Modules: enforce naming and unit conventions
        {
            files: ['src/**/*.module.css'],
            rules: {
                // No px/em for font-size — use rem. Warning only: 9 pre-existing occurrences as of this rule's
                // introduction — fix the ones you touch rather than sweeping the whole codebase.
                'declaration-property-unit-disallowed-list': [
                    {'font-size': ['px', 'em']},
                    {severity: 'warning'},
                ],

                // Never use z-index — rely on DOM order instead (see pkg/web/Velez-UI/CLAUDE.md)
                'declaration-property-value-disallowed-list': {
                    'z-index': [/.*/],
                },

                // PascalCase or camelCase class names (matches CSS Modules usage). Warning only: 5 pre-existing
                // occurrences (EnvCard.module.css uses snake_case status_* modifier classes) as of this rule's
                // introduction — fix the ones you touch rather than sweeping the whole codebase.
                'selector-class-pattern': ['^[A-Za-z][a-zA-Z0-9]*$', {severity: 'warning'}],
            },
        },
        // Global token file: hex color definitions and base sizing are the source of truth here
        {
            files: ['src/index.css'],
            rules: {
                'color-named': null,
                'declaration-property-unit-disallowed-list': null,
                'selector-class-pattern': null,
            },
        },
    ],
}

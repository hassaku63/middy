// @ts-check
import js from '@eslint/js';
import globals from 'globals';
import tseslint from 'typescript-eslint';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import prettier from 'eslint-config-prettier';
import { readdirSync } from 'node:fs';
import { join } from 'node:path';

// Architecture boundaries as lint rules (ADR-0007). Features are vertical
// slices that must not import each other; anything shared lives in
// src/shared. Components (*.tsx) stay free of side effects: no direct API
// calls, no browser globals; hooks and services own those.
const featureDirs = readdirSync(join(import.meta.dirname, 'src/features'), {
  withFileTypes: true,
})
  .filter((d) => d.isDirectory())
  .map((d) => d.name);

const featureBoundaryRules = featureDirs.map((name) => ({
  files: [`src/features/${name}/**/*.{ts,tsx}`],
  rules: {
    '@typescript-eslint/no-restricted-imports': [
      'error',
      {
        patterns: [
          {
            group: ['@/features/*', `!@/features/${name}`, `!@/features/${name}/**`],
            message: `features must not import each other; move shared code to src/shared (from ${name})`,
            allowTypeImports: true,
          },
          {
            group: ['../*', '!../*.css'],
            message:
              'use ./ inside a feature and @/shared for shared code; never reach into sibling features',
            allowTypeImports: true,
          },
        ],
      },
    ],
  },
}));

const componentPurityRules = {
  files: ['src/**/*.tsx'],
  ignores: ['src/**/*.test.tsx', 'src/main.tsx'],
  rules: {
    '@typescript-eslint/no-restricted-imports': [
      'error',
      {
        patterns: [
          {
            group: ['@/shared/api', '@/shared/api/*'],
            message:
              'components do not call the API; put the call in a hook (use*.ts) and pass results as props',
            allowTypeImports: true,
          },
        ],
      },
    ],
    'no-restricted-globals': [
      'error',
      ...[
        'fetch',
        'AudioContext',
        'localStorage',
        'sessionStorage',
        'indexedDB',
        'setTimeout',
        'setInterval',
      ].map((name) => ({
        name,
        message: `${name} is a side effect; keep it in a hook or service, not in a component`,
      })),
    ],
  },
};

const testHygieneRules = {
  files: ['src/**/*.test.{ts,tsx}'],
  rules: {
    'no-restricted-syntax': [
      'error',
      {
        selector: 'CallExpression[callee.property.name=/^toMatch(Inline)?Snapshot$/]',
        message:
          'snapshot tests are banned (docs/dev/testing.md); assert on roles, text and data instead',
      },
    ],
  },
};

export default tseslint.config(
  { ignores: ['dist', 'dev-dist', 'coverage', 'node_modules'] },
  js.configs.recommended,
  ...tseslint.configs.recommendedTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        projectService: { allowDefaultProject: ['eslint.config.js'] },
        tsconfigRootDir: import.meta.dirname,
      },
      globals: { ...globals.browser },
    },
  },
  {
    files: ['**/*.{ts,tsx}'],
    plugins: { 'react-hooks': reactHooks, 'react-refresh': reactRefresh },
    rules: {
      ...reactHooks.configs['recommended-latest'].rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      '@typescript-eslint/consistent-type-imports': 'error',
      '@typescript-eslint/no-floating-promises': 'error',
      // Size and complexity budgets (ADR-0007). Generous; a hit means "split".
      'max-lines': ['error', { max: 400, skipBlankLines: true, skipComments: true }],
      'max-lines-per-function': ['error', { max: 120, skipBlankLines: true, skipComments: true }],
      complexity: ['error', 15],
    },
  },
  {
    files: ['**/*.test.{ts,tsx}', 'src/test/**'],
    rules: { 'max-lines-per-function': 'off' },
  },
  {
    files: ['**/*.test.{ts,tsx}', 'src/test/**'],
    rules: { '@typescript-eslint/no-non-null-assertion': 'off' },
  },
  ...featureBoundaryRules,
  componentPurityRules,
  testHygieneRules,
  {
    files: ['*.config.{js,ts}'],
    languageOptions: { globals: { ...globals.node } },
  },
  {
    // Plain JS config files are not part of the TS project: no type-aware rules.
    files: ['**/*.js'],
    ...tseslint.configs.disableTypeChecked,
  },
  prettier,
);

import svelte from 'eslint-plugin-svelte';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsParser from '@typescript-eslint/parser';
import prettier from 'eslint-plugin-prettier';

/** @type {import('eslint').Linter.FlatConfig[]} */
export default [
  {
    files: ['**/*.svelte'],
    languageOptions: {
      parser: 'svelte-eslint-parser',
      parserOptions: {
        parser: { ts: '@typescript-eslint/parser' },
        extraFileExtensions: ['.svelte'],
        project: './tsconfig.json',
        tsconfigRootDir: import.meta.dirname,
      },
    },
    plugins: { svelte },
    rules: {
      ...svelte.configs.recommended.rules,
    },
  },
  {
    files: ['**/*.{js,ts}'],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        project: './tsconfig.json',
        tsconfigRootDir: import.meta.dirname,
      },
    },
    plugins: { '@typescript-eslint': tseslint },
    rules: {
      ...tseslint.configs.recommended.rules,
    },
  },
  {
    plugins: { prettier },
    rules: {
      'prettier/prettier': 'warn',
    },
  },
]; 
import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import ts from 'typescript-eslint';
import prettier from 'eslint-config-prettier';
import svelteParser from 'svelte-eslint-parser';

export default ts.config(
	js.configs.recommended,
	...ts.configs.recommended,
	...svelte.configs.recommended,
	prettier,
	{
		languageOptions: {
			globals: {
				...globals.browser,
				...globals.node,
				...globals.serviceworker
			}
		},
		rules: {
			'@typescript-eslint/no-explicit-any': 'warn',
			'@typescript-eslint/ban-ts-comment': 'warn',
			'svelte/prefer-svelte-reactivity': 'off',
			'svelte/no-navigation-without-resolve': 'warn',
			'svelte/no-at-html-tags': 'warn',
			'svelte/no-unused-svelte-ignore': 'warn',
			'no-useless-assignment': 'warn'
		}
	},
	{
		files: ['**/*.svelte'],
		languageOptions: {
			parserOptions: {
				parser: ts.parser
			}
		}
	},
	{
		files: ['**/*.svelte.ts'],
		languageOptions: {
			parser: svelteParser,
			parserOptions: {
				parser: ts.parser
			}
		}
	},
	{
		ignores: ['build/', '.svelte-kit/', 'node_modules/', 'static/hls.min.js']
	}
);

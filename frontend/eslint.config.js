import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import tseslint from 'typescript-eslint';
import prettier from 'eslint-config-prettier';

export default tseslint.config(
	js.configs.recommended,
	...tseslint.configs.recommended,
	...svelte.configs['flat/recommended'],
	prettier,
	{
		// TypeScript already reports undefined names (including DOM types such
		// as KeyboardEvent), and no-undef cannot see type-only globals.
		rules: { 'no-undef': 'off' },
		languageOptions: {
			globals: {
				console: 'readonly',
				// Plain global browser `fetch` in a .svelte <script> block needs
				// to be listed explicitly.
				fetch: 'readonly'
			}
		}
	},
	{
		// svelte-eslint-parser must hand <script lang="ts"> blocks to
		// typescript-eslint's parser, or `import type` and `$props()` type
		// annotations fail with a bare "Unexpected token" parse error.
		files: ['**/*.svelte'],
		languageOptions: {
			parserOptions: {
				parser: tseslint.parser
			}
		}
	},
	{
		ignores: ['build/', '.svelte-kit/', 'dist/']
	}
);

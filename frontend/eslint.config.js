import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import prettier from 'eslint-config-prettier'
import globals from 'globals'

// ESLint 9 "flat config". El orden importa: prettier va al final para
// desactivar las reglas de ESLint que chocan con el formateo de Prettier
// (Prettier manda en formato; ESLint manda en calidad de codigo).
export default [
  { ignores: ['dist/**', 'node_modules/**', 'coverage/**'] },

  js.configs.recommended,
  ...pluginVue.configs['flat/recommended'],

  {
    files: ['**/*.{js,vue}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },

  {
    rules: {
      // `_` como nombre = descarte intencional; no se reporta.
      // caughtErrors: 'none' = un `catch (e)` que no usa `e` no es error.
      'no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_', caughtErrors: 'none' },
      ],
      // Un `catch {}` vacio (tragar error a proposito) es valido.
      'no-empty': ['error', { allowEmptyCatch: true }],
    },
  },

  // vitest.config.js usa `globals: true`, asi que en los tests describe/it/expect/vi
  // son globales y no se importan.
  {
    files: ['**/*.{test,spec}.js'],
    languageOptions: {
      globals: {
        describe: 'readonly',
        it: 'readonly',
        test: 'readonly',
        expect: 'readonly',
        vi: 'readonly',
        beforeEach: 'readonly',
        afterEach: 'readonly',
        beforeAll: 'readonly',
        afterAll: 'readonly',
      },
    },
  },

  prettier,
]

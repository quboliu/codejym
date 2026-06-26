import js from '@eslint/js'
import globals from 'globals'
import vue from 'eslint-plugin-vue'
import vueParser from 'vue-eslint-parser'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.vue'],
    plugins: {
      vue: vue,
    },
    extends: [
      ...vue.configs['flat/recommended'],
    ],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tseslint.parser,
        ecmaVersion: 2020,
        sourceType: 'module',
      },
      ecmaVersion: 2020,
      globals: globals.browser,
    },
  },
  {
    files: ['**/*.{ts,js}'],
    extends: [
      js.configs.recommended,
      ...tseslint.configs.recommended,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
  },
])

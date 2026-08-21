# ESLint root
root: true

env:
  browser: true
  es2021: true

extends:
  - eslint:recommended
  - plugin:@typescript-eslint/recommended
  - plugin:react-hooks/recommended
  - prettier

parser: '@typescript-eslint/parser'

parserOptions:
  ecmaVersion: latest
  sourceType: module
  ecmaFeatures:
    jsx: true

plugins:
  - react-refresh
  - '@typescript-eslint'

rules:
  react-refresh/only-export-components: warn
  '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }]
  '@typescript-eslint/no-explicit-any': warn
  '@typescript-eslint/explicit-function-return-type': off
  '@typescript-eslint/explicit-module-boundary-types': off
  '@typescript-eslint/no-non-null-assertion': warn
  react-hooks/rules-of-hooks: error
  react-hooks/exhaustive-deps: warn

ignorePatterns:
  - dist
  - node_modules
  - '*.cjs

const process = require('node:process')

process.env.ESLINT_TSCONFIG = 'tsconfig.json'

module.exports = {
  extends: '@antfu',

  rules: {
    'vue/component-name-in-template-casing': [
      'warn',
      'PascalCase',
      { registeredComponentsOnly: false },
    ],
    // Require typescript in vue components
    'vue/block-lang': [
      'error',
      {
        script: {
          lang: 'ts',
        },
      },
    ],
    'curly': ['warn', 'all'],
    '@typescript-eslint/brace-style': ['warn', '1tbs', { allowSingleLine: true }],
  },
}

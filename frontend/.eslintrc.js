process.env.ESLINT_TSCONFIG = 'tsconfig.json'

module.exports = {
  extends: '@antfu',

  rules: {
    'vue/component-name-in-template-casing': [
      'warn',
      'PascalCase',
      { registeredComponentsOnly: false },
    ],
    'curly': ['warn', 'all'],
    '@typescript-eslint/brace-style': ['warn', '1tbs', { allowSingleLine: true }],
    'antfu/top-level-function': 'off',
  },
}

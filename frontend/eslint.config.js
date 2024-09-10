import antfu from '@antfu/eslint-config'

export default await antfu({
  typescript: true,
  vue: true,
  stylistic: true,
  jsonc: true,
  yaml: true,

  overrides: {
    typescript: {
      'ts/no-explicit-any': 'warn',
    },
  },
}, {
  rules: {
    curly: ['warn', 'all'],

    'style/brace-style': ['warn', '1tbs', { allowSingleLine: true }],
    'style/quote-props': ['warn', 'as-needed'],

    // Enable additional vue rules
    // https://eslint.vuejs.org/rules/
    'vue/component-name-in-template-casing': [
      'warn',
      'PascalCase',
      { registeredComponentsOnly: false },
    ],

    'jsonc/sort-keys': 'off',
    'regexp/prefer-d': 'off',
    'regexp/prefer-w': 'off',
  },
})

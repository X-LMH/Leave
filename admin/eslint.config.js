import { defineConfig } from '@soybeanjs/eslint-config-vue';

export default [
  {
    ignores: ['**/node_modules/**', '**/dist/**', '**/dist-ssr/**', '**/coverage/**']
  },
  ...(await defineConfig({
    'vue/component-name-in-template-casing': [
      'warn',
      'PascalCase',
      {
        registeredComponentsOnly: false,
        ignores: ['/^icon-/']
      }
    ]
  }))
];

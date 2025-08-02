// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,
  spaLoadingTemplate: 'spa-loading-template.html',

  modules: [
    '@nuxt/eslint',
    '@nuxt/ui',
    '@nuxt/icon',
    '@vueuse/nuxt',
    '@pinia/nuxt',
    'pinia-plugin-persistedstate/nuxt',
    '@nuxtjs/i18n',
  ],

  css: [
    '@/assets/styles/markdown.scss',
  ],

  piniaPluginPersistedstate: {
    storage: 'localStorage',
  },

  i18n: {
    defaultLocale: 'en',
    langDir: 'locales',
    strategy: 'no_prefix',
    locales: [
      { code: 'de', file: 'de.yml' },
      { code: 'en', file: 'en.yml' },
      { code: 'es', file: 'es.yml' },
    ],
  },

  icon: {
    serverBundle: {
      remote: false,
    },
    clientBundle: {
      scan: {
        globInclude: ['**/*.{vue,jsx,tsx,md,mdc,mdx,yml,yaml}', 'app.config.ts'],
      },
    },
  },

  eslint: {
    config: {
      standalone: false,
    },
  },

  typescript: {
    typeCheck: true,
    strict: true,
  },

  nitro: {
    devProxy: {
      '/_api': {
        target: 'http://localhost:8080/_api',
      },
    },
  },

  compatibilityDate: '2025-01-07',
})

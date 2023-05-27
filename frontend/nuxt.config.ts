// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,

  modules: [
    '@element-plus/nuxt',
    '@unocss/nuxt',
    '@vueuse/nuxt',
    'nuxt-icon',
    '@pinia/nuxt',
    '@pinia-plugin-persistedstate/nuxt',
    '@nuxtjs/i18n',
  ],

  elementPlus: {
    icon: false,
    importStyle: 'scss',
  },

  unocss: {
    // presets
    uno: true, // enable `@unocss/preset-uno`
    icons: true, // enable `@unocss/preset-icons`
    attributify: true, // enable `@unocss/preset-attributify`,

    // core options
    shortcuts: [],
    rules: [],
  },

  css: [
    '@/assets/styles/markdown.scss',
  ],

  piniaPersistedstate: {
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

  typescript: {
    // typeCheck: true,
    strict: true,
  },

  vite: {
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: '@use "@/assets/styles/elementplus.scss" as element;',
        },
      },
    },
  },

  nitro: {
    devProxy: {
      '/_api': {
        target: 'http://localhost:8080/_api',
      },
    },
  },
})

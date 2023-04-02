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
  ],

  elementPlus: {
    icon: false,
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

  piniaPersistedstate: {
    storage: 'localStorage',
  },

  typescript: {
    // typeCheck: true,
    strict: true,
  },

  nitro: {
    devProxy: {
      '/_api': {
        target: 'http://localhost:8080/_api',
      },
    },
  },
})

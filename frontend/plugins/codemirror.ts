import VueCodemirror from 'vue-codemirror'

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.use(VueCodemirror,
    {
      // Don't load basicSetup
      extensions: [],
    })
})

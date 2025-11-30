import { createApp } from 'vue'
import App from './App.vue'
import router from './router'

// importa TUTTE le funzioni di api.js
import * as api from './services/api.js'

import ErrorMsg from './components/ErrorMsg.vue'
import LoadingSpinner from './components/LoadingSpinner.vue'

import './assets/dashboard.css'
import './assets/main.css'

const app = createApp(App)

// rendo l'API disponibile ovunque come this.$api
app.config.globalProperties.$api = api

app.component('ErrorMsg', ErrorMsg)
app.component('LoadingSpinner', LoadingSpinner)

app.use(router)
app.mount('#app')

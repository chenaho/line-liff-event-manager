import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'
import App from './App.vue'
import router from './router'
import axios from 'axios'

// Set global API base URL
axios.defaults.baseURL = import.meta.env.VITE_API_BASE_URL || '/api' // Fallback to relative path for proxy

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')

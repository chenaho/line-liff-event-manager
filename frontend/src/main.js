import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './style.css'
import App from './App.vue'
import router from './router'
import axios from 'axios'
// import VConsole from 'vconsole'

// Initialize vConsole for debugging (disabled in production)
// const vConsole = new VConsole()

// Log environment variables for debugging
console.log('=== Environment Variables ===')
console.log('VITE_LIFF_ID:', import.meta.env.VITE_LIFF_ID)
console.log('VITE_API_BASE_URL:', import.meta.env.VITE_API_BASE_URL)
console.log('MODE:', import.meta.env.MODE)
console.log('============================')

// Set global API base URL
axios.defaults.baseURL = import.meta.env.VITE_API_BASE_URL || '/api' // Fallback to relative path for proxy

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')

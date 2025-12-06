import { defineStore } from 'pinia'
import liff from '@line/liff'
import axios from 'axios'

export const useAuthStore = defineStore('auth', {
    state: () => ({
        user: null,
        token: null,
        liffId: import.meta.env.VITE_LIFF_ID,
        isLiffInitialized: false,
        error: null
    }),
    actions: {
        async initLiff() {
            try {
                await liff.init({ liffId: this.liffId })
                this.isLiffInitialized = true

                if (liff.isLoggedIn()) {
                    const idToken = liff.getIDToken()
                    await this.loginBackend(idToken)
                } else {
                    liff.login()
                }
            } catch (err) {
                this.error = 'LIFF Init Failed: ' + err.message
                console.error(err)
            }
        },
        async loginBackend(idToken) {
            try {
                console.log('Login attempt with idToken:', idToken ? `${idToken.substring(0, 20)}...` : 'EMPTY!')
                console.log('idToken length:', idToken ? idToken.length : 0)
                const response = await axios.post('/api/auth/login', { idToken })
                this.token = response.data.token
                this.user = response.data.user

                // Set default auth header
                axios.defaults.headers.common['Authorization'] = `Bearer ${this.token}`
            } catch (err) {
                this.error = 'Backend Login Failed: ' + err.message
                console.error(err)
            }
        }
    }
})

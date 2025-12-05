import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import AdminView from '../views/AdminView.vue'
import EventView from '../views/EventView.vue'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            name: 'home',
            component: HomeView
        },
        {
            path: '/admin',
            name: 'admin',
            component: AdminView,
            // meta: { requiresAuth: true, requiresAdmin: true } // Add guard later
        },
        {
            path: '/event/:id',
            name: 'event',
            component: EventView
        }
    ]
})

export default router

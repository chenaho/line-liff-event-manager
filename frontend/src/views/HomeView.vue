<script setup>
import { onMounted, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

onMounted(() => {
  console.log('HomeView mounted')
  console.log('Query params:', route.query)
  console.log('EventId from query:', route.query.eventId)
  
  if (!authStore.isLiffInitialized) {
    authStore.initLiff()
  }
})

watch(() => authStore.user, (user) => {
  if (user) {
    // Check if there's an eventId in the query parameters
    const eventId = route.query.eventId
    
    console.log('User logged in:', user.lineDisplayName)
    console.log('User role:', user.role)
    console.log('EventId from query:', eventId)
    
    if (eventId) {
      console.log('Redirecting to event:', eventId)
      router.push(`/event/${eventId}`)
    } else if (user.role === 'admin') {
      console.log('Redirecting to admin dashboard')
      router.push('/admin')
    } else {
      console.log('Regular user without specific event - staying on home')
      // Regular user without specific event - stay on home
      // Could show a list of events or welcome message
    }
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100">
    <div class="text-center">
      <h1 class="text-4xl font-bold text-gray-800 mb-4">LINE Event Manager</h1>
      <p v-if="authStore.error" class="text-red-500">{{ authStore.error }}</p>
      <p v-else-if="!authStore.user" class="text-gray-600">Logging in...</p>
      <div v-else>
        <p class="text-xl">Welcome, {{ authStore.user.lineDisplayName }}</p>
        <p class="text-sm text-gray-500">Role: {{ authStore.user.role }}</p>
      </div>
    </div>
  </div>
</template>

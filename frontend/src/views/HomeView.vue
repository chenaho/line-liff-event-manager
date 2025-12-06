<script setup>
import { onMounted, watch, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter, useRoute } from 'vue-router'
import Toast from '../components/Toast.vue'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

// Generate avatar URL using dicebear or LINE profile picture
const avatarUrl = computed(() => {
  if (authStore.user?.pictureUrl) {
    return authStore.user.pictureUrl
  }
  const seed = authStore.user?.lineDisplayName || 'User'
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(seed)}`
})

const userIdShort = computed(() => {
  if (!authStore.user?.lineUserID) return 'U...'
  return authStore.user.lineUserID.substring(0, 8) + '...'
})

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
    }
  }
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header v-if="authStore.user" class="bg-gradient-to-r from-green-500 to-green-600 text-white p-4 shadow-md">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <div class="relative">
            <img :src="avatarUrl" alt="Profile" class="w-10 h-10 rounded-full border-2 border-white bg-gray-200">
            <div class="absolute bottom-0 right-0 w-3 h-3 bg-green-300 border-2 border-white rounded-full"></div>
          </div>
          <div>
            <div class="text-sm opacity-90">Hello,</div>
            <div class="font-bold text-lg">{{ authStore.user.lineDisplayName }}</div>
          </div>
        </div>
        <div class="text-xs bg-green-600 px-2 py-1 rounded">{{ userIdShort }}</div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex items-center justify-center p-6" :class="authStore.user ? 'min-h-[calc(100vh-80px)]' : 'min-h-screen'">
      <div class="text-center fade-in">
        <h1 class="text-4xl font-bold text-gray-800 mb-4">LINE Event Manager</h1>
        <p v-if="authStore.error" class="text-red-500 bg-red-50 px-4 py-2 rounded-lg">{{ authStore.error }}</p>
        <p v-else-if="!authStore.user" class="text-gray-600">
          <i class="fas fa-spinner fa-spin mr-2"></i>Logging in...
        </p>
        <div v-else class="space-y-2">
          <p class="text-xl text-gray-700">Welcome to LIFF Event Manager</p>
          <p class="text-sm text-gray-500">Role: <span class="font-semibold">{{ authStore.user.role }}</span></p>
        </div>
      </div>
    </main>

    <Toast />
  </div>
</template>

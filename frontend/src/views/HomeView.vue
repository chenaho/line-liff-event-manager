<script setup>
import { onMounted, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

onMounted(() => {
  if (!authStore.isLiffInitialized) {
    authStore.initLiff()
  }
})

watch(() => authStore.user, (user) => {
  if (user) {
    if (user.role === 'admin') {
      router.push('/admin')
    } else {
      // If there is an event ID in query, redirect there?
      // Or just stay home / list events?
      // For now, let's just show a welcome message or list.
      // Spec says "User View". Maybe list events?
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

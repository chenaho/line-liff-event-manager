<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useEventStore } from '../stores/event'
import EventVote from '../components/EventVote.vue'
import EventLineUp from '../components/EventLineUp.vue'
import EventMemo from '../components/EventMemo.vue'
import Toast from '../components/Toast.vue'

const route = useRoute()
const eventStore = useEventStore()
const eventId = route.params.id

const event = ref(null)
const status = ref(null)
let pollInterval = null

onMounted(async () => {
  event.value = await eventStore.fetchEvent(eventId)
  if (event.value) {
    status.value = await eventStore.fetchEventStatus(eventId)
    
    // Poll for status updates
    pollInterval = setInterval(async () => {
      status.value = await eventStore.fetchEventStatus(eventId)
    }, 5000)
  }
})

onUnmounted(() => {
  if (pollInterval) clearInterval(pollInterval)
})

const getTypeBadgeClass = (type) => {
  const classes = {
    'VOTE': 'bg-blue-100 text-blue-800',
    'LINEUP': 'bg-green-100 text-green-800',
    'MEMO': 'bg-purple-100 text-purple-800'
  }
  return classes[type] || 'bg-gray-100 text-gray-800'
}
</script>

<template>
  <div v-if="event" class="min-h-screen bg-gray-50 pb-20">
    <!-- Header -->
    <div class="bg-white p-4 shadow-md sticky top-0 z-10">
      <h1 class="text-xl font-bold text-gray-800">{{ event.title }}</h1>
      <div class="flex items-center gap-2 mt-1">
        <span 
          class="text-xs font-bold px-2 py-0.5 rounded"
          :class="getTypeBadgeClass(event.type)"
        >
          {{ event.type }}
        </span>
        <span 
          v-if="!event.isActive" 
          class="text-xs font-bold px-2 py-0.5 rounded bg-red-100 text-red-800"
        >
          CLOSED
        </span>
      </div>
    </div>

    <!-- Content -->
    <div class="p-4">
      <EventVote v-if="event.type === 'VOTE'" :event="event" :status="status" />
      <EventLineUp v-else-if="event.type === 'LINEUP'" :event="event" :status="status" />
      <EventMemo v-else-if="event.type === 'MEMO'" :event="event" :status="status" />
    </div>

    <Toast />
  </div>
  <div v-else class="p-4 text-center text-gray-500 min-h-screen flex items-center justify-center">
    <div>
      <i class="fas fa-spinner fa-spin text-4xl mb-4"></i>
      <p>Loading event...</p>
    </div>
  </div>
</template>

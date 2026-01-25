<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useEventStore } from '../stores/event'
import EventVote from '../components/EventVote.vue'
import EventLineUp from '../components/EventLineUp.vue'
import EventMemo from '../components/EventMemo.vue'
import Toast from '../components/Toast.vue'

const route = useRoute()
const router = useRouter()
const eventStore = useEventStore()

const event = ref(null)
const status = ref(null)
const notFound = ref(false)
let pollInterval = null

onMounted(async () => {
  const eventId = route.params.id
  const tagId = route.query.tagId

  // If tagId is provided, fetch event by tag
  if (tagId) {
    event.value = await eventStore.fetchEventByTag(tagId)
    if (!event.value) {
      notFound.value = true
      return
    }
    // Redirect to the actual event URL for consistency
    router.replace({ name: 'event', params: { id: event.value.eventId } })
  } else if (eventId) {
    event.value = await eventStore.fetchEvent(eventId)
  }

  if (event.value) {
    status.value = await eventStore.fetchEventStatus(event.value.eventId)
    
    // Poll for status updates
    pollInterval = setInterval(async () => {
      status.value = await eventStore.fetchEventStatus(event.value.eventId)
    }, 5000)
  } else {
    notFound.value = true
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

const formatDateTime = (dateTimeString) => {
  if (!dateTimeString) return ''
  const date = new Date(dateTimeString)
  return date.toLocaleString('zh-TW', { 
    month: '2-digit', 
    day: '2-digit', 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}
</script>

<template>
  <div v-if="event" class="min-h-screen bg-gray-50 pb-20">
    <!-- Header -->
    <div class="bg-white p-4 shadow-md sticky top-0 z-50">
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
      
      <!-- Time Range Display -->
      <div v-if="event.config.startTime || event.config.endTime" class="mt-2 text-sm text-gray-600">
        <i class="fas fa-clock mr-1"></i>
        <span v-if="event.config.startTime">{{ formatDateTime(event.config.startTime) }}</span>
        <span v-if="event.config.startTime && event.config.endTime"> - </span>
        <span v-if="event.config.endTime">{{ formatDateTime(event.config.endTime) }}</span>
      </div>
    </div>

    <!-- Content -->
    <div class="p-4">
      <EventVote v-if="event.type === 'VOTE'" :event="event" v-model:status="status" />
      <EventLineUp v-else-if="event.type === 'LINEUP'" :event="event" :status="status" />
      <EventMemo v-else-if="event.type === 'MEMO'" :event="event" :status="status" />
    </div>

    <Toast />
  </div>
  <div v-else-if="notFound" class="p-4 text-center text-gray-500 min-h-screen flex items-center justify-center">
    <div>
      <i class="fas fa-exclamation-circle text-4xl mb-4 text-red-400"></i>
      <p class="text-lg font-medium">找不到活動</p>
      <p class="text-sm mt-2">請確認網址是否正確</p>
    </div>
  </div>
  <div v-else class="p-4 text-center text-gray-500 min-h-screen flex items-center justify-center">
    <div>
      <i class="fas fa-spinner fa-spin text-4xl mb-4"></i>
      <p>Loading event...</p>
    </div>
  </div>
</template>

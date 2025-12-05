<script setup>
import { ref, onMounted } from 'vue'
import { useEventStore } from '../stores/event'
import { useAuthStore } from '../stores/auth'

const eventStore = useEventStore()
const authStore = useAuthStore()

const showCreateModal = ref(false)
const newEvent = ref({
  title: '',
  type: 'VOTE',
  config: {
    // VOTE defaults
    allowMultiSelect: false,
    maxVotes: 1,
    options: [],
    optionsText: '', // For textarea input
    
    // LINEUP defaults
    maxParticipants: 10,
    waitlistLimit: 5,
    maxCountPerUser: 1,
    
    // MEMO defaults
    maxCommentsPerUser: 3,
    allowReaction: true
  }
})

onMounted(() => {
  eventStore.fetchEvents()
})

const createEvent = async () => {
  // Process options for VOTE
  if (newEvent.value.type === 'VOTE') {
    newEvent.value.config.options = newEvent.value.config.optionsText.split('\n').filter(o => o.trim())
  }
  
  try {
    await eventStore.createEvent(newEvent.value)
    showCreateModal.value = false
    // Reset form (simplified)
    newEvent.value.title = ''
  } catch (e) {
    alert('Failed to create event')
  }
}

const toggleStatus = (event) => {
  eventStore.updateEventStatus(event.eventId, !event.isActive)
}

const copyLink = (eventId) => {
  const liffId = authStore.liffId
  const url = `https://liff.line.me/${liffId}?eventId=${eventId}` // Or path based
  // Actually spec says: https://liff.line.me/{liffId}?eventId={id}
  // But our router uses /event/:id.
  // We need to handle query param in HomeView to redirect.
  // Or use path in LIFF URL if configured: https://liff.line.me/{liffId}/event/{id}
  // Let's assume query param for now as it's standard LIFF.
  navigator.clipboard.writeText(url)
  alert('Copied: ' + url)
}
</script>

<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-gray-800">Admin Dashboard</h1>
      <button @click="showCreateModal = true" class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
        + Create Event
      </button>
    </div>

    <div v-if="eventStore.loading" class="text-center py-4">Loading...</div>
    
    <div v-else class="grid gap-4">
      <div v-for="event in eventStore.events" :key="event.eventId" class="bg-white p-4 rounded shadow border border-gray-200 flex justify-between items-center">
        <div>
          <div class="flex items-center gap-2">
            <span class="px-2 py-1 text-xs font-bold rounded" 
              :class="{
                'bg-blue-100 text-blue-800': event.type === 'VOTE',
                'bg-green-100 text-green-800': event.type === 'LINEUP',
                'bg-yellow-100 text-yellow-800': event.type === 'MEMO'
              }">
              {{ event.type }}
            </span>
            <h2 class="text-xl font-semibold">{{ event.title }}</h2>
          </div>
          <p class="text-sm text-gray-500 mt-1">ID: {{ event.eventId }}</p>
        </div>
        
        <div class="flex items-center gap-3">
          <button @click="copyLink(event.eventId)" class="text-gray-600 hover:text-blue-600">
            Link
          </button>
          <button @click="toggleStatus(event)" 
            class="px-3 py-1 rounded text-sm font-medium"
            :class="event.isActive ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">
            {{ event.isActive ? 'Active' : 'Inactive' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
      <div class="bg-white rounded-lg p-6 w-full max-w-lg">
        <h2 class="text-2xl font-bold mb-4">New Event</h2>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">Title</label>
            <input v-model="newEvent.title" type="text" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2">
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700">Type</label>
            <select v-model="newEvent.type" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2">
              <option value="VOTE">Vote</option>
              <option value="LINEUP">Line Up</option>
              <option value="MEMO">Memo</option>
            </select>
          </div>

          <!-- Type Specific Config -->
          <div v-if="newEvent.type === 'VOTE'" class="space-y-2 border-t pt-2">
            <label class="block text-sm font-medium text-gray-700">Options (One per line)</label>
            <textarea v-model="newEvent.config.optionsText" rows="4" class="block w-full border border-gray-300 rounded-md shadow-sm p-2"></textarea>
            <div class="flex items-center">
              <input v-model="newEvent.config.allowMultiSelect" type="checkbox" class="mr-2">
              <span class="text-sm">Allow Multi-select</span>
            </div>
          </div>

          <div v-if="newEvent.type === 'LINEUP'" class="space-y-2 border-t pt-2">
            <div>
              <label class="block text-sm font-medium text-gray-700">Max Participants</label>
              <input v-model.number="newEvent.config.maxParticipants" type="number" class="block w-full border border-gray-300 rounded-md shadow-sm p-2">
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">Waitlist Limit</label>
              <input v-model.number="newEvent.config.waitlistLimit" type="number" class="block w-full border border-gray-300 rounded-md shadow-sm p-2">
            </div>
          </div>
          
          <!-- MEMO config skipped for brevity -->

        </div>

        <div class="mt-6 flex justify-end gap-3">
          <button @click="showCreateModal = false" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded">Cancel</button>
          <button @click="createEvent" class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">Create</button>
        </div>
      </div>
    </div>
  </div>
</template>

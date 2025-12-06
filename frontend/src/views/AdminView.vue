<script setup>
import { ref, onMounted } from 'vue'
import { useEventStore } from '../stores/event'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'
import Toast from '../components/Toast.vue'

const eventStore = useEventStore()
const authStore = useAuthStore()
const { showToast } = useToast()

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
    // Reset form
    newEvent.value.title = ''
    newEvent.value.config.optionsText = ''
    showToast('Event created successfully!')
  } catch (e) {
    showToast('Failed to create event')
  }
}

const toggleStatus = (event) => {
  eventStore.updateEventStatus(event.eventId, !event.isActive)
  showToast(`Event ${!event.isActive ? 'activated' : 'deactivated'}`)
}

const copyLink = (eventId) => {
  const liffId = authStore.liffId
  const url = `https://liff.line.me/${liffId}?eventId=${eventId}`
  navigator.clipboard.writeText(url)
  showToast('Link copied to clipboard!')
}

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
  <div class="p-6 max-w-4xl mx-auto min-h-screen bg-gray-50">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-gray-800">
        <i class="fas fa-chart-line mr-2 text-blue-600"></i>
        Admin Dashboard
      </h1>
      <button 
        @click="showCreateModal = true" 
        class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors shadow-md flex items-center gap-2"
      >
        <i class="fas fa-plus"></i>
        Create Event
      </button>
    </div>

    <div v-if="eventStore.loading" class="text-center py-8">
      <i class="fas fa-spinner fa-spin text-4xl text-gray-400"></i>
      <p class="text-gray-500 mt-2">Loading...</p>
    </div>
    
    <div v-else-if="eventStore.events.length === 0" class="text-center py-12">
      <i class="fas fa-inbox text-6xl text-gray-300 mb-4"></i>
      <p class="text-gray-500">No events yet. Create your first event!</p>
    </div>
    
    <div v-else class="grid gap-4">
      <div 
        v-for="event in eventStore.events" 
        :key="event.eventId" 
        class="bg-white p-4 rounded-xl shadow-md border border-gray-200 flex justify-between items-center hover:shadow-lg transition-shadow"
      >
        <div>
          <div class="flex items-center gap-2">
            <span 
              class="px-2 py-1 text-xs font-bold rounded" 
              :class="getTypeBadgeClass(event.type)"
            >
              {{ event.type }}
            </span>
            <h2 class="text-xl font-semibold text-gray-800">{{ event.title }}</h2>
          </div>
          <p class="text-sm text-gray-500 mt-1">ID: {{ event.eventId }}</p>
        </div>
        
        <div class="flex items-center gap-3">
          <button 
            @click="copyLink(event.eventId)" 
            class="text-gray-600 hover:text-blue-600 transition-colors px-3 py-1 rounded hover:bg-blue-50"
          >
            <i class="fas fa-link mr-1"></i>
            Link
          </button>
          <button 
            @click="toggleStatus(event)" 
            class="px-3 py-1 rounded text-sm font-medium transition-colors"
            :class="event.isActive 
              ? 'bg-green-100 text-green-700 hover:bg-green-200' 
              : 'bg-red-100 text-red-700 hover:bg-red-200'"
          >
            <i :class="event.isActive ? 'fas fa-check-circle' : 'fas fa-times-circle'" class="mr-1"></i>
            {{ event.isActive ? 'Active' : 'Inactive' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <div 
      v-if="showCreateModal" 
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      @click.self="showCreateModal = false"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-lg shadow-2xl fade-in">
        <h2 class="text-2xl font-bold mb-4 text-gray-800">
          <i class="fas fa-plus-circle mr-2 text-blue-600"></i>
          New Event
        </h2>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Title</label>
            <input 
              v-model="newEvent.title" 
              type="text" 
              class="w-full border border-gray-300 rounded-lg shadow-sm p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter event title"
            >
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <select 
              v-model="newEvent.type" 
              class="w-full border border-gray-300 rounded-lg shadow-sm p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="VOTE">Vote</option>
              <option value="LINEUP">Line Up</option>
              <option value="MEMO">Memo</option>
            </select>
          </div>

          <!-- Type Specific Config -->
          <div v-if="newEvent.type === 'VOTE'" class="space-y-2 border-t pt-4">
            <label class="block text-sm font-medium text-gray-700">Options (One per line)</label>
            <textarea 
              v-model="newEvent.config.optionsText" 
              rows="4" 
              class="w-full border border-gray-300 rounded-lg shadow-sm p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Option 1&#10;Option 2&#10;Option 3"
            ></textarea>
            <div class="flex items-center">
              <input 
                v-model="newEvent.config.allowMultiSelect" 
                type="checkbox" 
                class="mr-2 w-4 h-4 text-blue-600"
              >
              <span class="text-sm text-gray-700">Allow Multi-select</span>
            </div>
          </div>

          <div v-if="newEvent.type === 'LINEUP'" class="space-y-3 border-t pt-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">最大參與人數</label>
              <input 
                type="number" 
                v-model.number="newEvent.config.maxParticipants" 
                placeholder="最大參與人數"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
                required
              >
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">每人報名上限 (0=無限制)</label>
              <input 
                type="number" 
                v-model.number="newEvent.config.maxCountPerUser" 
                placeholder="0"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
                min="0"
              >
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">候補名額上限</label>
              <input 
                v-model.number="newEvent.config.waitlistLimit" 
                type="number" 
                class="w-full border border-gray-300 rounded-lg shadow-sm p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
            </div>
          </div>
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <button 
            @click="showCreateModal = false" 
            class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
          >
            Cancel
          </button>
          <button 
            @click="createEvent" 
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors shadow-md"
          >
            Create
          </button>
        </div>
      </div>
    </div>

    <Toast />
  </div>
</template>

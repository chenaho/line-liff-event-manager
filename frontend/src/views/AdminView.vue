<script setup>
import { ref, onMounted, computed } from 'vue'
import { useEventStore } from '../stores/event'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'
import Toast from '../components/Toast.vue'

const eventStore = useEventStore()
const authStore = useAuthStore()
const { showToast } = useToast()

const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingEvent = ref(null)
const showArchived = ref(false)

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
    allowReaction: true,
    
    // Time range (optional)
    startTime: '',
    endTime: ''
  }
})

onMounted(() => {
  eventStore.fetchEvents()
})

const createEvent = async () => {
  // Create a copy to avoid mutating the original
  const eventData = JSON.parse(JSON.stringify(newEvent.value))
  
  // Process options for VOTE
  if (eventData.type === 'VOTE') {
    eventData.config.options = eventData.config.optionsText.split('\n').filter(o => o.trim())
    delete eventData.config.optionsText
  }
  
  // Convert datetime-local to ISO format and remove if empty
  if (eventData.config.startTime) {
    // datetime-local format: "2025-12-07T10:00"
    // Convert to ISO 8601: "2025-12-07T10:00:00Z"
    eventData.config.startTime = new Date(eventData.config.startTime).toISOString()
  } else {
    delete eventData.config.startTime
  }
  
  if (eventData.config.endTime) {
    eventData.config.endTime = new Date(eventData.config.endTime).toISOString()
  } else {
    delete eventData.config.endTime
  }
  
  // Remove optionsText for non-VOTE types
  if (eventData.type !== 'VOTE' && eventData.config.optionsText !== undefined) {
    delete eventData.config.optionsText
  }
  
  try {
    await eventStore.createEvent(eventData)
    showCreateModal.value = false
    // Reset form
    newEvent.value.title = ''
    newEvent.value.config.optionsText = ''
    newEvent.value.config.startTime = ''
    newEvent.value.config.endTime = ''
    showToast('Event created successfully!')
  } catch (e) {
    console.error('Create event error:', e)
    showToast('Failed to create event: ' + (e.response?.data?.error || e.message))
  }
}


const openEditModal = (event) => {
  // Convert ISO datetime to datetime-local format (YYYY-MM-DDTHH:mm)
  const formatForInput = (isoString) => {
    if (!isoString) return ''
    const date = new Date(isoString)
    // Get local time in YYYY-MM-DDTHH:mm format
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${year}-${month}-${day}T${hours}:${minutes}`
  }

  editingEvent.value = {
    ...event,
    config: {
      ...event.config,
      optionsText: event.config.options?.join('\n') || '',
      startTime: formatForInput(event.config.startTime),
      endTime: formatForInput(event.config.endTime)
    }
  }
  showEditModal.value = true
}


const updateEvent = async () => {
  // Create a copy to avoid mutating the original
  const eventData = JSON.parse(JSON.stringify(editingEvent.value))
  
  // Process options for VOTE
  if (eventData.type === 'VOTE') {
    eventData.config.options = eventData.config.optionsText.split('\n').filter(o => o.trim())
    delete eventData.config.optionsText
  }
  
  // Convert datetime-local to ISO format and remove if empty
  if (eventData.config.startTime) {
    eventData.config.startTime = new Date(eventData.config.startTime).toISOString()
  } else {
    delete eventData.config.startTime
  }
  
  if (eventData.config.endTime) {
    eventData.config.endTime = new Date(eventData.config.endTime).toISOString()
  } else {
    delete eventData.config.endTime
  }
  
  // Remove optionsText for non-VOTE types
  if (eventData.type !== 'VOTE' && eventData.config.optionsText !== undefined) {
    delete eventData.config.optionsText
  }
  
  try {
    await eventStore.updateEvent(eventData.eventId, eventData)
    showEditModal.value = false
    editingEvent.value = null
    showToast('Event updated successfully!')
  } catch (e) {
    console.error('Update event error:', e)
    showToast('Failed to update event: ' + (e.response?.data?.error || e.message))
  }
}

const toggleStatus = async (event) => {
  await eventStore.updateEventStatus(event.eventId, !event.isActive)
  showToast(`Event ${event.isActive ? 'deactivated' : 'activated'}`)
}

const archiveEvent = async (event) => {
  try {
    await eventStore.archiveEvent(event.eventId, !event.isArchived)
    showToast(event.isArchived ? '已取消封存' : '已封存活動')
  } catch (e) {
    showToast('操作失敗: ' + e.message)
  }
}

const filteredEvents = computed(() => {
  if (!eventStore.events) return []
  if (showArchived.value) {
    return eventStore.events
  }
  return eventStore.events.filter(e => !e.isArchived)
})

const copyLink = (eventId) => {
  const liffId = authStore.liffId
  const url = `https://liff.line.me/${liffId}?eventId=${eventId}`
  navigator.clipboard.writeText(url)
  showToast('Link copied!')
  // Open the link in a new window
  window.open(url, '_blank')
}


const getTypeBadgeClass = (type) => {
  const classes = {
    'VOTE': 'bg-blue-100 text-blue-800',
    'LINEUP': 'bg-green-100 text-green-800',
    'MEMO': 'bg-purple-100 text-purple-800'
  }
  return classes[type] || 'bg-gray-100 text-gray-800'
}

const isEventActive = (event) => {
  if (!event.config.startTime && !event.config.endTime) {
    return event.isActive // No time restrictions
  }
  
  const now = new Date()
  const start = event.config.startTime ? new Date(event.config.startTime) : null
  const end = event.config.endTime ? new Date(event.config.endTime) : null
  
  if (start && now < start) return false
  if (end && now > end) return false
  
  return event.isActive
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
    
    <!-- Events List -->
    <template v-else>
      <!-- Show Archived Toggle -->
      <div class="mb-4 flex items-center gap-2">
        <label class="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" v-model="showArchived" class="w-4 h-4 rounded" />
          <span class="text-sm text-gray-600">顯示已封存的活動</span>
        </label>
      </div>

      <div class="grid gap-4">
        <div 
          v-for="event in filteredEvents" 
          :key="event.eventId" 
          class="bg-white p-4 rounded-xl shadow-md border border-gray-200 hover:shadow-lg transition-shadow"
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
          
          <!-- Time Range Display -->
          <div v-if="event.config.startTime || event.config.endTime" class="mt-2 text-sm text-gray-600">
            <i class="fas fa-clock mr-1"></i>
            <span v-if="event.config.startTime">{{ formatDateTime(event.config.startTime) }}</span>
            <span v-if="event.config.startTime && event.config.endTime"> - </span>
            <span v-if="event.config.endTime">{{ formatDateTime(event.config.endTime) }}</span>
          </div>
          
          <!-- Status Indicator -->
          <div class="mt-2 flex items-center gap-2">
            <span 
              class="inline-flex items-center px-2 py-1 text-xs font-medium rounded"
              :class="isEventActive(event) 
                ? 'bg-green-100 text-green-700' 
                : 'bg-gray-100 text-gray-700'"
            >
              <span 
                class="w-2 h-2 rounded-full mr-1"
                :class="isEventActive(event) ? 'bg-green-500' : 'bg-gray-400'"
              ></span>
              {{ isEventActive(event) ? '進行中' : '未開始/已結束' }}
            </span>
          </div>
          
          <p class="text-xs text-gray-400 mt-2">ID: {{ event.eventId }}</p>
        </div>
        
        <!-- Action Buttons - Moved Below -->
        <div class="mt-4 pt-3 border-t border-gray-100 flex items-center gap-2">
          <button 
            @click="openEditModal(event)" 
            class="text-gray-600 hover:text-blue-600 transition-colors px-3 py-1 rounded hover:bg-blue-50"
            title="編輯活動"
          >
            <i class="fas fa-edit mr-1"></i>
            Edit
          </button>
          <button 
            @click="copyLink(event.eventId)" 
            class="text-gray-600 hover:text-blue-600 transition-colors px-3 py-1 rounded hover:bg-blue-50"
            title="複製連結"
          >
            <i class="fas fa-link mr-1"></i>
            Link
          </button>
          <button 
            @click="toggleStatus(event)" 
            :class="event.isActive 
              ? 'text-green-600 hover:text-green-700 hover:bg-green-50' 
              : 'text-red-600 hover:text-red-700 hover:bg-red-50'"
            class="transition-colors px-3 py-1 rounded"
            :title="event.isActive ? '停用活動' : '啟用活動'"
          >
            <i :class="event.isActive ? 'fas fa-toggle-on' : 'fas fa-toggle-off'" class="mr-1"></i>
            {{ event.isActive ? 'Active' : 'Inactive' }}
          </button>
          <button 
            @click="archiveEvent(event)" 
            :class="event.isArchived 
              ? 'text-yellow-600 hover:text-yellow-700 hover:bg-yellow-50' 
              : 'text-gray-600 hover:text-gray-700 hover:bg-gray-50'"
            class="transition-colors px-3 py-1 rounded"
            :title="event.isArchived ? '取消封存' : '封存活動'"
          >
            <i :class="event.isArchived ? 'fas fa-box-open' : 'fas fa-archive'" class="mr-1"></i>
            {{ event.isArchived ? 'Unarchive' : 'Archive' }}
          </button>
        </div>
      </div>
    </div>
    </template>

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

          <!-- Time Range (Optional) -->
          <div class="border-t pt-4 space-y-3">
            <label class="block text-sm font-medium text-gray-700">
              <i class="fas fa-clock mr-1"></i>
              活動時間範圍 (選填)
            </label>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs text-gray-600 mb-1">開始時間</label>
                <input 
                  v-model="newEvent.config.startTime" 
                  type="datetime-local" 
                  class="w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
              </div>
              <div>
                <label class="block text-xs text-gray-600 mb-1">結束時間</label>
                <input 
                  v-model="newEvent.config.endTime" 
                  type="datetime-local" 
                  class="w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
              </div>
            </div>
            <p class="text-xs text-gray-500">
              <i class="fas fa-info-circle mr-1"></i>
              如果不設定時間，活動將沒有時間限制
            </p>
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

          <div v-if="newEvent.type === 'MEMO'" class="space-y-3 border-t pt-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">每人留言上限</label>
              <input 
                type="number" 
                v-model.number="newEvent.config.maxCommentsPerUser" 
                placeholder="例如：3"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
                min="1"
              >
              <p class="text-xs text-gray-500 mt-1">設定每位用戶最多可發表幾則留言</p>
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

    <!-- Edit Modal -->
    <div 
      v-if="showEditModal && editingEvent" 
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      @click.self="showEditModal = false"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto shadow-2xl fade-in">
        <h2 class="text-2xl font-bold mb-4 text-gray-800">
          <i class="fas fa-edit mr-2 text-blue-600"></i>
          Edit Event
        </h2>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Title</label>
            <input 
              v-model="editingEvent.title" 
              type="text" 
              class="w-full border border-gray-300 rounded-lg shadow-sm p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <select 
              v-model="editingEvent.type" 
              class="w-full border border-gray-300 rounded-lg shadow-sm p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled
            >
              <option value="VOTE">Vote</option>
              <option value="LINEUP">Line Up</option>
              <option value="MEMO">Memo</option>
            </select>
            <p class="text-xs text-gray-500 mt-1">類型無法修改</p>
          </div>

          <!-- Time Range -->
          <div class="border-t pt-4 space-y-3">
            <label class="block text-sm font-medium text-gray-700">
              <i class="fas fa-clock mr-1"></i>
              活動時間範圍 (選填)
            </label>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs text-gray-600 mb-1">開始時間</label>
                <input 
                  v-model="editingEvent.config.startTime" 
                  type="datetime-local" 
                  class="w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
              </div>
              <div>
                <label class="block text-xs text-gray-600 mb-1">結束時間</label>
                <input 
                  v-model="editingEvent.config.endTime" 
                  type="datetime-local" 
                  class="w-full border border-gray-300 rounded-lg shadow-sm p-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
              </div>
            </div>
          </div>

          <!-- VOTE Config -->
          <div v-if="editingEvent.type === 'VOTE'" class="space-y-2 border-t pt-4">
            <label class="block text-sm font-medium text-gray-700">Options (One per line)</label>
            <textarea 
              v-model="editingEvent.config.optionsText" 
              rows="4" 
              class="w-full border border-gray-300 rounded-lg shadow-sm p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
            ></textarea>
          </div>

          <!-- LINEUP Config -->
          <div v-if="editingEvent.type === 'LINEUP'" class="space-y-3 border-t pt-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">最大參與人數</label>
              <input 
                type="number" 
                v-model.number="editingEvent.config.maxParticipants" 
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
              >
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">每人報名上限 (0=無限制)</label>
              <input 
                type="number" 
                v-model.number="editingEvent.config.maxCountPerUser" 
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
                min="0"
              >
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">候補名額上限</label>
              <input 
                type="number" 
                v-model.number="editingEvent.config.waitlistLimit" 
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
                min="0"
              >
            </div>
          </div>

          <div v-if="editingEvent.type === 'MEMO'" class="space-y-3 border-t pt-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">每人留言上限</label>
              <input 
                type="number" 
                v-model.number="editingEvent.config.maxCommentsPerUser" 
                placeholder="例如：3"
                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
                min="1"
              >
              <p class="text-xs text-gray-500 mt-1">設定每位用戶最多可發表幾則留言</p>
            </div>
          </div>
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <button 
            @click="showEditModal = false" 
            class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
          >
            Cancel
          </button>
          <button 
            @click="updateEvent" 
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors shadow-md"
          >
            Update
          </button>
        </div>
      </div>
    </div>

    <Toast />
  </div>
</template>

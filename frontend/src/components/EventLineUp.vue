<script setup>
import { computed } from 'vue'
import { useEventStore } from '../stores/event'
import { useAuthStore } from '../stores/auth'

const props = defineProps(['event', 'status'])
const eventStore = useEventStore()
const authStore = useAuthStore()

const myRecord = computed(() => {
  if (!props.status || !props.status.records) return null
  return props.status.records.find(r => r.userId === authStore.user?.lineUserId && r.type === 'LINEUP')
})

const successCount = computed(() => {
  if (!props.status || !props.status.records) return 0
  return props.status.records.filter(r => r.type === 'LINEUP' && r.status === 'SUCCESS').length
})

const isFull = computed(() => successCount.value >= props.event.config.maxParticipants)

const handleAction = async (count) => {
  try {
    await eventStore.submitAction(props.event.eventId, 'LINEUP', {
      count: count,
      userDisplayName: authStore.user?.lineDisplayName // Send name for display
    })
  } catch (e) {
    alert('Action failed: ' + e.message)
  }
}
</script>

<template>
  <div>
    <div class="text-center mb-6">
      <div class="text-5xl font-bold text-blue-600 mb-2">
        {{ successCount }} <span class="text-2xl text-gray-400">/ {{ event.config.maxParticipants }}</span>
      </div>
      <p class="text-gray-500">Participants</p>
    </div>

    <div class="mb-8">
      <button 
        v-if="!myRecord"
        @click="handleAction(1)"
        class="w-full py-4 rounded-xl font-bold text-xl shadow-lg transition transform active:scale-95"
        :class="isFull ? 'bg-yellow-500 text-white hover:bg-yellow-600' : 'bg-green-500 text-white hover:bg-green-600'"
      >
        {{ isFull ? 'Join Waitlist' : '+1 Join Now' }}
      </button>

      <button 
        v-else
        @click="handleAction(-1)"
        class="w-full bg-red-100 text-red-600 py-4 rounded-xl font-bold text-xl hover:bg-red-200 transition"
      >
        -1 Cancel ({{ myRecord.status }})
      </button>
    </div>

    <!-- List -->
    <div class="space-y-2">
      <h3 class="font-bold text-gray-700">Participants</h3>
      <div v-if="status && status.records">
        <div v-for="rec in status.records.filter(r => r.type === 'LINEUP' && r.status === 'SUCCESS')" :key="rec.userId" class="flex items-center gap-3 p-2 bg-white rounded shadow-sm">
          <div class="w-8 h-8 bg-gray-200 rounded-full flex items-center justify-center text-xs">
            {{ rec.userDisplayName ? rec.userDisplayName[0] : '?' }}
          </div>
          <span>{{ rec.userDisplayName || 'Unknown' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

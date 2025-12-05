<script setup>
import { ref, computed } from 'vue'
import { useEventStore } from '../stores/event'

const props = defineProps(['event', 'status'])
const eventStore = useEventStore()

const selected = ref([])

const submitVote = async () => {
  if (selected.value.length === 0) return
  
  try {
    await eventStore.submitAction(props.event.eventId, 'VOTE', {
      selectedOptions: selected.value
    })
    alert('Voted!')
  } catch (e) {
    alert('Failed to vote')
  }
}

// Calculate percentages
const results = computed(() => {
  if (!props.status || !props.status.records) return {}
  const counts = {}
  let total = 0
  props.status.records.forEach(r => {
    if (r.type === 'VOTE' && r.selectedOptions) {
      r.selectedOptions.forEach(opt => {
        counts[opt] = (counts[opt] || 0) + 1
        total++
      })
    }
  })
  return { counts, total }
})

const getPercent = (option) => {
  if (!results.value.total) return 0
  return Math.round(((results.value.counts[option] || 0) / results.value.total) * 100)
}
</script>

<template>
  <div class="space-y-4">
    <div v-for="option in event.config.options" :key="option" class="border p-3 rounded hover:bg-gray-50 cursor-pointer">
      <label class="flex items-center w-full cursor-pointer">
        <input 
          v-if="event.config.allowMultiSelect" 
          type="checkbox" 
          :value="option" 
          v-model="selected"
          class="mr-3 h-5 w-5 text-blue-600"
        >
        <input 
          v-else 
          type="radio" 
          :value="option" 
          v-model="selected" 
          name="vote-option"
          class="mr-3 h-5 w-5 text-blue-600"
        > <!-- Note: v-model with array for radio is tricky, assume single value for radio logic but here using array for consistency -->
        
        <div class="flex-1">
          <div class="flex justify-between mb-1">
            <span class="font-medium">{{ option }}</span>
            <span class="text-sm text-gray-500">{{ results.counts[option] || 0 }} votes ({{ getPercent(option) }}%)</span>
          </div>
          <div class="w-full bg-gray-200 rounded-full h-2.5">
            <div class="bg-blue-600 h-2.5 rounded-full" :style="{ width: getPercent(option) + '%' }"></div>
          </div>
        </div>
      </label>
    </div>

    <button @click="submitVote" class="w-full bg-blue-600 text-white py-3 rounded-lg font-bold text-lg hover:bg-blue-700 transition">
      Submit Vote
    </button>
  </div>
</template>

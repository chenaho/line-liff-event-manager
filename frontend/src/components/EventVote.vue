<script setup>
import { ref, computed } from 'vue'
import { useEventStore } from '../stores/event'
import { useToast } from '../composables/useToast'

const props = defineProps(['event', 'status'])
const eventStore = useEventStore()
const { showToast } = useToast()

const selected = ref([])

const submitVote = async () => {
  if (selected.value.length === 0) {
    showToast('Please select at least one option')
    return
  }
  
  try {
    await eventStore.submitAction(props.event.eventId, 'VOTE', {
      selectedOptions: selected.value
    })
    showToast('Vote submitted successfully!')
  } catch (e) {
    showToast('Failed to submit vote')
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

const getVoteCount = (option) => {
  return results.value.counts[option] || 0
}

const totalVotes = computed(() => {
  return results.value.total || 0
})

const isSelected = (option) => {
  return selected.value.includes(option)
}

const toggleOption = (option) => {
  if (props.event.config.allowMultiSelect) {
    const index = selected.value.indexOf(option)
    if (index > -1) {
      selected.value.splice(index, 1)
    } else {
      selected.value.push(option)
    }
  } else {
    selected.value = [option]
  }
}
</script>

<template>
  <div class="fade-in space-y-6">
    <!-- Event Info Card -->
    <div class="bg-white p-4 rounded-xl shadow-sm border border-gray-100">
      <span class="text-xs font-bold text-blue-500 bg-blue-50 px-2 py-1 rounded">VOTE</span>
      <h2 class="text-2xl font-bold mt-2 text-gray-800">{{ event.title }}</h2>
      <div class="mt-2 text-xs text-gray-400">
        {{ event.config.allowMultiSelect ? `複選 (最多 ${event.config.maxVotes || '無限制'} 項)` : '單選' }} • 共 {{ totalVotes }} 票
      </div>
    </div>

    <!-- Options -->
    <div class="space-y-3">
      <div 
        v-for="option in event.config.options" 
        :key="option" 
        @click="toggleOption(option)"
        class="relative bg-white rounded-lg p-3 shadow-sm border cursor-pointer transition-all duration-200 hover:shadow-md"
        :class="isSelected(option) ? 'border-blue-500 ring-1 ring-blue-500' : 'border-gray-200'"
      >
        <div class="flex justify-between items-center relative z-10">
          <div class="flex items-center gap-3">
            <!-- Custom Checkbox/Radio -->
            <div 
              class="w-5 h-5 rounded border flex items-center justify-center transition-all duration-200"
              :class="isSelected(option) ? 'bg-blue-500 border-blue-500' : 'border-gray-300'"
            >
              <i v-if="isSelected(option)" class="fas fa-check text-white text-xs"></i>
            </div>
            <span class="font-medium text-gray-700">{{ option }}</span>
          </div>
          <span class="text-sm font-bold text-gray-600">{{ getVoteCount(option) }} 票</span>
        </div>
        
        <!-- Progress Bar -->
        <div class="absolute bottom-0 left-0 h-1 bg-blue-100 rounded-bl-lg rounded-br-lg w-full overflow-hidden">
          <div 
            class="h-full bg-blue-500 transition-all duration-500" 
            :style="{ width: getPercent(option) + '%' }"
          ></div>
        </div>
      </div>
    </div>

    <!-- Submit Button -->
    <button 
      @click="submitVote" 
      class="w-full bg-blue-600 text-white py-3 rounded-xl font-bold shadow-lg hover:bg-blue-700 active:scale-95 transition-transform"
    >
      提交投票
    </button>
  </div>
</template>

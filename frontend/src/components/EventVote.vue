<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useEventStore } from '../stores/event'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'

const props = defineProps(['event', 'status'])
const eventStore = useEventStore()
const authStore = useAuthStore()
const { showToast } = useToast()

const selected = ref([])
const expandedOptions = ref({}) // Track which options have expanded voter lists

// Find current user's vote
const myVote = computed(() => {
  if (!props.status?.records) return null
  return props.status.records.find(r => 
    r.type === 'VOTE' && r.userId === authStore.user?.lineUserId
  )
})

// Initialize selected options with user's previous vote
onMounted(() => {
  if (myVote.value?.selectedOptions) {
    selected.value = [...myVote.value.selectedOptions]
  }
})

// Watch for status changes to update selected options
watch(() => props.status, () => {
  if (myVote.value?.selectedOptions && selected.value.length === 0) {
    selected.value = [...myVote.value.selectedOptions]
  }
}, { deep: true })

const submitVote = async () => {
  if (selected.value.length === 0) {
    showToast('請至少選擇一個選項')
    return
  }
  
  try {
    await eventStore.submitAction(props.event.eventId, 'VOTE', {
      selectedOptions: selected.value
    })
    showToast('投票成功！')
  } catch (e) {
    showToast('投票失敗: ' + e.message)
  }
}

// Calculate percentages
const results = computed(() => {
  if (!props.status || !props.status.records) {
    return { counts: {}, total: 0 }
  }
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
  if (!results.value.counts) return 0
  return results.value.counts[option] || 0
}

// Get voters for specific option
const getVoters = (option) => {
  if (!props.status?.records) return []
  return props.status.records
    .filter(r => r.type === 'VOTE' && r.selectedOptions?.includes(option))
    .map(r => ({
      userId: r.userId,
      displayName: r.userDisplayName || 'Unknown',
      isMe: r.userId === authStore.user?.lineUserId
    }))
}

const getAvatarUrl = (displayName) => {
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(displayName)}`
}

const toggleVoterList = (option) => {
  expandedOptions.value[option] = !expandedOptions.value[option]
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
      <div v-if="myVote" class="mt-2 text-xs text-green-600 bg-green-50 px-2 py-1 rounded inline-block">
        <i class="fas fa-check-circle mr-1"></i>
        您已投票
      </div>
    </div>

    <!-- Options -->
    <div class="space-y-3">
      <div 
        v-for="option in event.config.options" 
        :key="option" 
        class="bg-white rounded-lg shadow-sm border overflow-hidden"
        :class="isSelected(option) ? 'border-blue-500 ring-1 ring-blue-500' : 'border-gray-200'"
      >
        <!-- Option Header (clickable) -->
        <div 
          @click="toggleOption(option)"
          class="relative p-3 cursor-pointer transition-all duration-200 hover:bg-gray-50"
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
          <div class="absolute bottom-0 left-0 h-1 bg-blue-100 w-full overflow-hidden">
            <div 
              class="h-full bg-blue-500 transition-all duration-500" 
              :style="{ width: getPercent(option) + '%' }"
            ></div>
          </div>
        </div>

        <!-- Voter List Toggle Button -->
        <div 
          v-if="getVoteCount(option) > 0"
          @click="toggleVoterList(option)"
          class="px-3 py-2 bg-gray-50 border-t border-gray-100 cursor-pointer hover:bg-gray-100 transition-colors flex items-center justify-between text-xs text-gray-600"
        >
          <span>
            <i class="fas fa-users mr-1"></i>
            {{ getVoteCount(option) }} 位投票者
          </span>
          <i class="fas" :class="expandedOptions[option] ? 'fa-chevron-up' : 'fa-chevron-down'"></i>
        </div>

        <!-- Voter List (Expandable) -->
        <div 
          v-if="expandedOptions[option]" 
          class="px-3 py-2 bg-gray-50 border-t border-gray-100 space-y-2"
        >
          <div 
            v-for="voter in getVoters(option)" 
            :key="voter.userId"
            class="flex items-center gap-2 p-2 rounded"
            :class="voter.isMe ? 'bg-blue-50' : 'bg-white'"
          >
            <img 
              :src="getAvatarUrl(voter.displayName)" 
              class="w-6 h-6 rounded-full bg-gray-200"
              :alt="voter.displayName"
            >
            <span class="text-sm text-gray-700">
              {{ voter.displayName }}
              <span v-if="voter.isMe" class="text-blue-600 font-semibold">(您)</span>
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Submit Button -->
    <button 
      @click="submitVote" 
      class="w-full bg-blue-600 text-white py-3 rounded-xl font-bold shadow-lg hover:bg-blue-700 active:scale-95 transition-transform"
    >
      {{ myVote ? '更新投票' : '提交投票' }}
    </button>
  </div>
</template>

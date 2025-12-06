<script setup>
import { computed } from 'vue'
import { useEventStore } from '../stores/event'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'

const props = defineProps(['event', 'status'])
const eventStore = useEventStore()
const authStore = useAuthStore()
const { showToast } = useToast()

const myRecord = computed(() => {
  if (!props.status || !props.status.records) return null
  return props.status.records.find(r => r.userId === authStore.user?.lineUserId && r.type === 'LINEUP' && r.status !== 'CANCELLED')
})

const successCount = computed(() => {
  if (!props.status || !props.status.records) return 0
  return props.status.records.filter(r => r.type === 'LINEUP' && r.status === 'SUCCESS').length
})

const waitlistCount = computed(() => {
  if (!props.status || !props.status.records) return 0
  return props.status.records.filter(r => r.type === 'LINEUP' && r.status === 'WAITLIST').length
})

const isFull = computed(() => successCount.value >= props.event.config.maxParticipants)
const remaining = computed(() => props.event.config.maxParticipants - successCount.value)

const participants = computed(() => {
  if (!props.status || !props.status.records) return []
  return props.status.records.filter(r => r.type === 'LINEUP' && r.status === 'SUCCESS')
})

const waitlist = computed(() => {
  if (!props.status || !props.status.records) return []
  return props.status.records.filter(r => r.type === 'LINEUP' && r.status === 'WAITLIST')
})

const getAvatarUrl = (displayName) => {
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(displayName || 'User')}`
}

const handleAction = async (count) => {
  try {
    await eventStore.submitAction(props.event.eventId, 'LINEUP', {
      count: count,
      userDisplayName: authStore.user?.lineDisplayName
    })
    showToast(count > 0 ? '報名成功！' : '已取消報名')
  } catch (e) {
    showToast('操作失敗: ' + e.message)
  }
}
</script>

<template>
  <div class="fade-in space-y-4">
    <!-- Event Info Card -->
    <div class="bg-white p-4 rounded-2xl shadow-sm border-l-4 border-green-500">
      <div class="flex items-start gap-3">
        <!-- Icon -->
        <div class="text-3xl">🏸</div>
        
        <!-- Content -->
        <div class="flex-1">
          <h2 class="text-xl font-bold text-gray-800">{{ event.title }}</h2>
          <p v-if="event.description" class="text-sm text-gray-500 mt-1">{{ event.description }}</p>
          
          <!-- Stats Row -->
          <div class="flex items-center gap-4 mt-3">
            <div>
              <div class="text-xs text-gray-400">目前人數</div>
              <div class="text-2xl font-black" :class="isFull ? 'text-red-500' : 'text-green-500'">
                {{ successCount }} <span class="text-sm text-gray-400 font-normal">/ {{ event.config.maxParticipants }}</span>
              </div>
            </div>
            <div class="text-right ml-auto">
              <div class="text-xs text-gray-500">候補名額: {{ waitlistCount }}</div>
            </div>
          </div>
        </div>
        
        <!-- Avatar Placeholders -->
        <div class="flex -space-x-2">
          <div v-for="i in Math.min(3, successCount)" :key="i" 
               class="w-10 h-10 rounded-full bg-gray-200 border-2 border-white"></div>
        </div>
      </div>
    </div>

    <!-- Cancel Button (only shown when user has registered) -->
    <button 
      v-if="myRecord"
      @click="handleAction(-1)"
      class="w-full bg-pink-50 text-pink-600 py-3 rounded-xl font-bold border border-pink-200 hover:bg-pink-100 transition-colors flex items-center justify-center"
    >
      <i class="fas fa-times mr-2"></i>
      取消報名 (-1)
    </button>
    
    <!-- Join Button (only shown when user hasn't registered) -->
    <button 
      v-else
      @click="handleAction(1)"
      class="w-full py-3 rounded-xl font-bold shadow-lg transition-all active:scale-95 flex items-center justify-center"
      :class="isFull ? 'bg-yellow-500 text-white hover:bg-yellow-600' : 'bg-green-500 text-white hover:bg-green-600'"
    >
      <i class="fas fa-plus mr-2"></i>
      {{ isFull ? '加入候補' : '我要報名 (+1)' }}
    </button>

    <!-- Participants List -->
    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div class="bg-gray-50 px-4 py-2.5 text-sm font-medium text-gray-600 border-b">
        已報名名單
      </div>
      
      <ul class="divide-y divide-gray-100">
        <li 
          v-for="(p, index) in participants" 
          :key="p.userId" 
          class="p-4 flex items-center justify-between hover:bg-gray-50 transition-colors"
        >
          <div class="flex items-center gap-3">
            <!-- Number -->
            <div class="font-mono text-gray-400 w-6 text-center font-medium">{{ index + 1 }}</div>
            
            <!-- Avatar -->
            <img :src="getAvatarUrl(p.userDisplayName)" class="w-10 h-10 rounded-full bg-gray-200 border border-gray-200">
            
            <!-- Name and Note -->
            <div>
              <div class="font-medium text-gray-800">{{ p.userDisplayName || 'Unknown' }}</div>
              <div v-if="p.note" class="text-xs text-gray-500 mt-0.5">
                {{ p.note }}
              </div>
            </div>
          </div>
          
          <!-- Status Badge -->
          <div class="text-xs text-green-600 font-bold bg-green-50 px-3 py-1 rounded-full">
            正取
          </div>
        </li>
        
        <!-- Empty State -->
        <li v-if="participants.length === 0" class="p-8 text-center text-gray-400">
          <i class="fas fa-users text-3xl mb-2"></i>
          <p class="text-sm">還沒有人報名</p>
        </li>
      </ul>
      
      <!-- Waitlist Section -->
      <div v-if="waitlist.length > 0">
        <div class="bg-yellow-50 px-4 py-2.5 text-sm font-medium text-yellow-700 border-t border-b border-yellow-100">
          候補名單
        </div>
        <ul class="divide-y divide-gray-100">
          <li 
            v-for="(p, index) in waitlist" 
            :key="p.userId" 
            class="p-4 flex items-center justify-between hover:bg-gray-50 transition-colors"
          >
            <div class="flex items-center gap-3">
              <div class="font-mono text-gray-400 w-6 text-center font-medium">{{ index + 1 }}</div>
              <img :src="getAvatarUrl(p.userDisplayName)" class="w-10 h-10 rounded-full bg-gray-200 border border-gray-200">
              <div>
                <div class="font-medium text-gray-800">{{ p.userDisplayName || 'Unknown' }}</div>
              </div>
            </div>
            <div class="text-xs text-yellow-600 font-bold bg-yellow-50 px-3 py-1 rounded-full border border-yellow-200">
              候補
            </div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

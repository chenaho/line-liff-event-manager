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
  return props.status.records.find(r => r.userId === authStore.user?.lineUserID && r.type === 'LINEUP')
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
    <!-- Header Card -->
    <div class="bg-white p-5 rounded-xl shadow-sm border-l-4 border-green-500 relative overflow-hidden">
      <div class="absolute right-[-10px] top-[-10px] opacity-10">
        <i class="fas fa-users text-9xl"></i>
      </div>
      
      <h2 class="text-2xl font-bold text-gray-800 relative z-10">{{ event.title }}</h2>
      
      <div class="flex justify-between items-end mt-4 relative z-10">
        <div>
          <div class="text-sm text-gray-400">目前人數</div>
          <div class="text-3xl font-black" :class="isFull ? 'text-red-500' : 'text-green-500'">
            {{ successCount }} <span class="text-lg text-gray-400 font-normal">/ {{ event.config.maxParticipants }}</span>
          </div>
        </div>
        <div class="text-right">
          <div class="text-xs bg-gray-100 px-2 py-1 rounded text-gray-500 mb-1">
            候補名額: {{ event.config.waitlistLimit }}
          </div>
          <div v-if="!isFull" class="text-xs text-green-600 font-semibold">
            還有 {{ remaining }} 個名額
          </div>
          <div v-else class="text-xs text-red-600 font-semibold">
            已額滿
          </div>
        </div>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="flex gap-2">
      <button 
        v-if="!myRecord"
        @click="handleAction(1)"
        class="flex-1 py-3 rounded-xl font-bold shadow-lg transition-all active:scale-95 flex items-center justify-center"
        :class="isFull ? 'bg-yellow-500 text-white hover:bg-yellow-600' : 'bg-green-500 text-white hover:bg-green-600'"
      >
        <i class="fas fa-plus mr-2"></i>
        {{ isFull ? '加入候補' : '我要報名 (+1)' }}
      </button>
      
      <button 
        v-else
        @click="handleAction(-1)"
        class="flex-1 bg-red-100 text-red-600 py-3 rounded-xl font-bold border border-red-200 hover:bg-red-200 transition-colors flex items-center justify-center"
      >
        <i class="fas fa-times mr-2"></i>
        取消報名 (-1)
      </button>
    </div>

    <!-- Participants List -->
    <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
      <div class="bg-gray-50 px-4 py-2 text-xs font-bold text-gray-500 border-b">已報名名單</div>
      <ul class="divide-y divide-gray-100">
        <li 
          v-for="(p, index) in participants" 
          :key="p.userId" 
          class="p-3 flex items-center justify-between"
        >
          <div class="flex items-center gap-3">
            <div class="font-mono text-gray-300 w-4 text-center text-sm">{{ index + 1 }}</div>
            <img :src="getAvatarUrl(p.userDisplayName)" class="w-8 h-8 rounded-full bg-gray-200">
            <div>
              <div class="font-medium text-sm text-gray-800">{{ p.userDisplayName || 'Unknown' }}</div>
              <div v-if="p.note" class="text-xs text-gray-500 bg-yellow-50 inline-block px-1 rounded mt-0.5">
                {{ p.note }}
              </div>
            </div>
          </div>
          <div class="text-xs text-green-600 font-bold bg-green-50 px-2 py-1 rounded-full">正取</div>
        </li>
      </ul>
      
      <!-- Waitlist Section -->
      <div v-if="waitlist.length > 0">
        <div class="bg-yellow-50 px-4 py-2 text-xs font-bold text-yellow-700 border-t border-b">候補名單</div>
        <ul class="divide-y divide-gray-100">
          <li 
            v-for="(p, index) in waitlist" 
            :key="p.userId" 
            class="p-3 flex items-center justify-between"
          >
            <div class="flex items-center gap-3">
              <div class="font-mono text-gray-300 w-4 text-center text-sm">{{ index + 1 }}</div>
              <img :src="getAvatarUrl(p.userDisplayName)" class="w-8 h-8 rounded-full bg-gray-200">
              <div>
                <div class="font-medium text-sm text-gray-800">{{ p.userDisplayName || 'Unknown' }}</div>
              </div>
            </div>
            <div class="text-xs text-yellow-600 font-bold bg-yellow-50 px-2 py-1 rounded-full">候補</div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

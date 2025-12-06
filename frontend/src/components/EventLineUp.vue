<script setup>
import { ref, computed } from 'vue'
import { useEventStore } from '../stores/event'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'

const props = defineProps(['event', 'status'])
const eventStore = useEventStore()
const authStore = useAuthStore()
const { showToast } = useToast()

// Count user's active registrations
const myRegistrationCount = computed(() => {
  if (!props.status?.records) return 0
  return props.status.records.filter(r => 
    r.userId === authStore.user?.lineUserId && 
    r.type === 'LINEUP' && 
    r.status !== 'CANCELLED'
  ).length
})

// Check if user can register more
const canRegisterMore = computed(() => {
  if (authStore.user?.role === 'admin') return true
  const limit = props.event.config.maxCountPerUser || 0
  return limit === 0 || myRegistrationCount.value < limit
})

const successCount = computed(() => {
  if (!props.status?.records) return 0
  return props.status.records.filter(r => r.type === 'LINEUP' && r.status === 'SUCCESS').length
})

const waitlistCount = computed(() => {
  if (!props.status?.records) return 0
  return props.status.records.filter(r => r.type === 'LINEUP' && r.status === 'WAITLIST').length
})

const isFull = computed(() => successCount.value >= props.event.config.maxParticipants)
const remaining = computed(() => props.event.config.maxParticipants - successCount.value)

// Group participants by user and count registrations
const participants = computed(() => {
  if (!props.status?.records) return []
  
  const userMap = new Map()
  props.status.records.forEach(r => {
    if (r.type === 'LINEUP' && r.status === 'SUCCESS') {
      if (!userMap.has(r.userId)) {
        userMap.set(r.userId, {
          userId: r.userId,
          displayName: r.userDisplayName || 'Unknown',
          count: 0
        })
      }
      userMap.get(r.userId).count++
    }
  })
  
  return Array.from(userMap.values())
})

const waitlist = computed(() => {
  if (!props.status?.records) return []
  
  const userMap = new Map()
  props.status.records.forEach(r => {
    if (r.type === 'LINEUP' && r.status === 'WAITLIST') {
      if (!userMap.has(r.userId)) {
        userMap.set(r.userId, {
          userId: r.userId,
          displayName: r.userDisplayName || 'Unknown',
          count: 0
        })
      }
      userMap.get(r.userId).count++
    }
  })
  
  return Array.from(userMap.values())
})

const getAvatarUrl = (displayName) => {
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(displayName)}`
}

const handleRegister = async () => {
  if (!canRegisterMore.value) {
    const limit = props.event.config.maxCountPerUser
    showToast(`已達報名上限 (${limit})`)
    return
  }
  
  try {
    await eventStore.submitAction(props.event.eventId, 'LINEUP', {
      count: 1,
      userDisplayName: authStore.user?.lineDisplayName
    })
    showToast('報名成功！')
  } catch (e) {
    showToast('報名失敗: ' + e.message)
  }
}

const handleCancel = async () => {
  if (myRegistrationCount.value === 0) {
    showToast('沒有可取消的報名')
    return
  }
  
  try {
    await eventStore.submitAction(props.event.eventId, 'LINEUP', {
      count: -1,
      userDisplayName: authStore.user?.lineDisplayName
    })
    showToast('已取消一筆報名')
  } catch (e) {
    showToast('取消失敗: ' + e.message)
  }
}
</script>

<template>
  <div class="fade-in space-y-6">
    <!-- Event Info Card -->
    <div class="bg-gradient-to-r from-blue-500 to-purple-500 p-6 rounded-xl shadow-lg text-white relative overflow-hidden">
      <div class="absolute top-0 right-0 opacity-10">
        <i class="fas fa-users text-9xl"></i>
      </div>
      <div class="relative z-10">
        <span class="text-xs font-bold bg-white/20 px-2 py-1 rounded">LINEUP</span>
        <h2 class="text-3xl font-bold mt-2">{{ event.title }}</h2>
        <div class="mt-4 flex items-center gap-4">
          <div class="text-center">
            <div class="text-4xl font-bold">{{ successCount }}</div>
            <div class="text-xs opacity-80">/ {{ event.config.maxParticipants }}</div>
            <div class="text-xs opacity-80">目前人數</div>
          </div>
          <div v-if="waitlistCount > 0" class="text-center">
            <div class="text-2xl font-bold">{{ waitlistCount }}</div>
            <div class="text-xs opacity-80">候補名額</div>
          </div>
        </div>
        
        <!-- User's Registration Info -->
        <div v-if="myRegistrationCount > 0" class="mt-4 bg-white/20 px-3 py-2 rounded-lg">
          <div class="text-sm">
            您的報名數: <span class="font-bold text-lg">{{ myRegistrationCount }}</span>
            <span v-if="event.config.maxCountPerUser > 0 && authStore.user?.role !== 'admin'">
              / {{ event.config.maxCountPerUser }}
            </span>
            <span v-if="authStore.user?.role === 'admin'" class="text-xs ml-2">(管理員無限制)</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="grid grid-cols-2 gap-3">
      <button 
        @click="handleCancel"
        :disabled="myRegistrationCount === 0"
        class="bg-pink-50 text-pink-600 py-3 rounded-xl font-bold shadow-sm border border-pink-200 hover:bg-pink-100 active:scale-95 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <i class="fas fa-minus-circle mr-2"></i>
        取消一筆
      </button>
      <button 
        @click="handleRegister"
        :disabled="!canRegisterMore || isFull"
        class="bg-blue-600 text-white py-3 rounded-xl font-bold shadow-lg hover:bg-blue-700 active:scale-95 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <i class="fas fa-plus-circle mr-2"></i>
        再報名一次
      </button>
    </div>

    <!-- Participants List -->
    <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
      <div class="px-4 py-3 bg-gray-50 border-b border-gray-100">
        <h3 class="font-bold text-gray-700">已報名名單 ({{ participants.length }})</h3>
      </div>
      
      <div v-if="participants.length === 0" class="px-4 py-8 text-center text-gray-400">
        <i class="fas fa-inbox text-4xl mb-2"></i>
        <p class="text-sm">尚無人報名</p>
      </div>
      
      <div v-else class="divide-y divide-gray-100">
        <div 
          v-for="(participant, index) in participants" 
          :key="participant.userId"
          class="px-4 py-3 hover:bg-gray-50 transition-colors flex items-center gap-3"
        >
          <span class="text-sm font-bold text-gray-400 w-6">{{ index + 1 }}</span>
          <img 
            :src="getAvatarUrl(participant.displayName)" 
            class="w-10 h-10 rounded-full bg-gray-200 border-2 border-blue-500"
            :alt="participant.displayName"
          >
          <div class="flex-1">
            <div class="font-medium text-gray-800">
              {{ participant.displayName }}
              <span v-if="participant.count > 1" class="text-blue-600 font-bold ml-1">(x{{ participant.count }})</span>
            </div>
          </div>
          <span class="text-xs font-bold text-green-600 bg-green-50 px-2 py-1 rounded">正取</span>
        </div>
      </div>
    </div>

    <!-- Waitlist -->
    <div v-if="waitlist.length > 0" class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
      <div class="px-4 py-3 bg-orange-50 border-b border-orange-100">
        <h3 class="font-bold text-orange-700">候補名單 ({{ waitlist.length }})</h3>
      </div>
      
      <div class="divide-y divide-gray-100">
        <div 
          v-for="(person, index) in waitlist" 
          :key="person.userId"
          class="px-4 py-3 hover:bg-gray-50 transition-colors flex items-center gap-3"
        >
          <span class="text-sm font-bold text-gray-400 w-6">{{ index + 1 }}</span>
          <img 
            :src="getAvatarUrl(person.displayName)" 
            class="w-10 h-10 rounded-full bg-gray-200 border-2 border-orange-500"
            :alt="person.displayName"
          >
          <div class="flex-1">
            <div class="font-medium text-gray-800">
              {{ person.displayName }}
              <span v-if="person.count > 1" class="text-orange-600 font-bold ml-1">(x{{ person.count }})</span>
            </div>
          </div>
          <span class="text-xs font-bold text-orange-600 bg-orange-50 px-2 py-1 rounded">候補</span>
        </div>
      </div>
    </div>
  </div>
</template>

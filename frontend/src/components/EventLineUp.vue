<script setup>
import { ref, computed } from 'vue'
import { useEventStore } from '../stores/event'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'

const props = defineProps(['event', 'status'])
const eventStore = useEventStore()
const authStore = useAuthStore()
const { showToast } = useToast()

const registrationNote = ref('')

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

// Check if registration is completely closed (event full AND waitlist full)
const isRegistrationClosed = computed(() => {
  if (!isFull.value) return false // Event not full, can register
  
  // Event is full, check if waitlist has space
  const waitlistLimit = props.event.config.waitlistLimit || 0
  if (waitlistLimit === 0) return false // No waitlist limit, can always join waitlist
  
  return waitlistCount.value >= waitlistLimit // Waitlist is also full
})

const remaining = computed(() => props.event.config.maxParticipants - successCount.value)

// Show individual registrations instead of grouping
const participants = computed(() => {
  if (!props.status?.records) return []
  
  return props.status.records
    .filter(r => r.type === 'LINEUP' && r.status === 'SUCCESS')
    .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp)) // Sort by timestamp
    .map(r => ({
      id: r.id, // Add record ID
      userId: r.userId,
      displayName: r.userDisplayName || 'Unknown',
      pictureUrl: r.userPictureUrl || null,
      note: r.note || '',
      timestamp: r.timestamp,
      isMe: r.userId === authStore.user?.lineUserId
    }))
})

const waitlist = computed(() => {
  if (!props.status?.records) return []
  
  return props.status.records
    .filter(r => r.type === 'LINEUP' && r.status === 'WAITLIST')
    .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
    .map(r => ({
      id: r.id, // Add record ID
      userId: r.userId,
      displayName: r.userDisplayName || 'Unknown',
      pictureUrl: r.userPictureUrl || null,
      note: r.note || '',
      timestamp: r.timestamp,
      isMe: r.userId === authStore.user?.lineUserId
    }))
})

// Privacy mode settings
const isPrivacyMode = computed(() => props.event.config.privacyMode === true)

const getDisplayAvatar = (person) => {
  // In privacy mode, always use generated avatar
  if (isPrivacyMode.value && !person.isMe) {
    return `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(person.userId)}`
  }
  // Normal mode: use actual picture or fallback
  if (person.pictureUrl) {
    return person.pictureUrl + '/small'
  }
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(person.displayName)}`
}

const getMaskedName = (person) => {
  // Show full name for self
  if (person.isMe) {
    return person.displayName
  }
  // In privacy mode, mask name
  if (isPrivacyMode.value) {
    const name = person.displayName || 'Unknown'
    if (name.length <= 2) {
      return name.charAt(0) + '*'
    }
    return name.substring(0, 4) + '...'
  }
  // Normal mode: show full name, truncate if too long
  const name = person.displayName || 'Unknown'
  if (name.length > 10) {
    return name.substring(0, 10) + '...'
  }
  return name
}

const getAvatarUrl = (person) => {
  // Prioritize LINE pictureUrl with /small suffix
  if (person.pictureUrl) {
    return person.pictureUrl + '/small'
  }
  // Fallback to generated avatar
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(person.displayName)}`
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
      userDisplayName: authStore.user?.lineDisplayName,
      userPictureUrl: authStore.user?.pictureUrl,
      note: registrationNote.value
    })
    showToast('報名成功！')
    registrationNote.value = '' // Clear note after registration
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
      userDisplayName: authStore.user?.lineDisplayName,
      userPictureUrl: authStore.user?.pictureUrl
    })
    showToast('取消成功！')
  } catch (e) {
    showToast('取消失敗: ' + e.message)
  }
}

// Note editing
const editingNote = ref(false)
const editingNoteValue = ref('')
const editingRecordId = ref(null)

const openNoteEditor = (record) => {
  if (!record.isMe) return // Only allow editing own notes
  
  if (record.id) {
    editingRecordId.value = record.id
    editingNoteValue.value = record.note || ''
    editingNote.value = true
  }
}

const saveNote = async () => {
  try {
    // Update note via API
    await eventStore.updateRegistrationNote(
      props.event.eventId, 
      editingRecordId.value, 
      editingNoteValue.value
    )
    showToast('備註已更新！')
    editingNote.value = false
  } catch (e) {
    showToast('更新失敗: ' + e.message)
  }
}

const cancelNoteEdit = () => {
  editingNote.value = false
  editingNoteValue.value = ''
  editingRecordId.value = null
}

// Export to clipboard
const exportToClipboard = async () => {
  const maxParticipants = props.event.config.maxParticipants
  let text = `[${props.event.title}] - 人數上限:${maxParticipants}\n`
  
  // Main participants list
  for (let i = 0; i < maxParticipants; i++) {
    const num = String(i + 1).padStart(2, '0')
    if (i < participants.value.length) {
      const p = participants.value[i]
      const noteText = p.note ? `-${p.note}` : ''
      text += `${num}. ${p.displayName}${noteText}\n`
    } else {
      text += `${num}.\n`
    }
  }
  
  // Waitlist
  if (waitlist.value.length > 0 || props.event.config.waitlistLimit > 0) {
    text += `--候補--\n`
    const waitlistLimit = props.event.config.waitlistLimit || waitlist.value.length
    for (let i = 0; i < Math.max(waitlistLimit, waitlist.value.length); i++) {
      const num = String(i + 1).padStart(2, '0')
      if (i < waitlist.value.length) {
        const p = waitlist.value[i]
        const noteText = p.note ? `-${p.note}` : ''
        text += `${num}. ${p.displayName}${noteText}\n`
      } else {
        text += `${num}.\n`
      }
    }
  }
  
  try {
    await navigator.clipboard.writeText(text)
    showToast('已複製到剪貼簿！')
  } catch (e) {
    showToast('複製失敗: ' + e.message)
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
          <div v-if="waitlistCount > 0 || event.config.waitlistLimit > 0" class="text-center">
            <div class="text-2xl font-bold">{{ waitlistCount }}</div>
            <div class="text-xs opacity-80" v-if="event.config.waitlistLimit > 0">
              / {{ event.config.waitlistLimit }}
            </div>
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

    <!-- Registration Note Input -->
    <div class="bg-white p-4 rounded-xl shadow-sm border border-gray-100">
      <label class="block text-sm font-medium text-gray-700 mb-2">
        <i class="fas fa-sticky-note mr-1"></i>
        報名備註 (選填)
      </label>
      <input 
        v-model="registrationNote"
        type="text"
        placeholder="例如：幫朋友報名、攜帶裝備等"
        class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none text-sm"
      >
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
        :disabled="!canRegisterMore || isRegistrationClosed"
        class="bg-blue-600 text-white py-3 rounded-xl font-bold shadow-lg hover:bg-blue-700 active:scale-95 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <i class="fas fa-plus-circle mr-2"></i>
        報名+1
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
          :key="index"
          @click="openNoteEditor(participant)"
          class="px-4 py-3 hover:bg-gray-50 transition-colors"
          :class="[
            participant.isMe ? 'bg-blue-50 cursor-pointer' : '',
            participant.isMe ? 'hover:bg-blue-100' : ''
          ]"
        >
          <div class="flex items-center gap-3">
            <span class="text-sm font-bold text-gray-400 w-6">{{ index + 1 }}</span>
            <img 
              :src="getDisplayAvatar(participant)" 
              class="w-10 h-10 rounded-full bg-gray-200 border-2"
              :class="participant.isMe ? 'border-blue-500' : 'border-gray-300'"
              :alt="getMaskedName(participant)"
            >
            <div class="flex-1 min-w-0">
              <div class="font-medium text-gray-800 truncate">
                {{ getMaskedName(participant) }}
                <span v-if="participant.isMe" class="text-blue-600 text-xs ml-1">(您)</span>
              </div>
              <div v-if="participant.note" class="text-xs text-gray-500 mt-0.5">
                <i class="fas fa-sticky-note mr-1"></i>
                {{ participant.note }}
              </div>
            </div>
            <span class="text-xs font-bold text-green-600 bg-green-50 px-2 py-1 rounded">正取</span>
          </div>
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
          :key="index"
          @click="openNoteEditor(person)"
          class="px-4 py-3 hover:bg-gray-50 transition-colors"
          :class="[
            person.isMe ? 'bg-orange-50 cursor-pointer' : '',
            person.isMe ? 'hover:bg-orange-100' : ''
          ]"
        >
          <div class="flex items-center gap-3">
            <span class="text-sm font-bold text-gray-400 w-6">{{ index + 1 }}</span>
            <img 
              :src="getDisplayAvatar(person)" 
              class="w-10 h-10 rounded-full bg-gray-200 border-2"
              :class="person.isMe ? 'border-orange-500' : 'border-gray-300'"
              :alt="getMaskedName(person)"
            >
            <div class="flex-1 min-w-0">
              <div class="font-medium text-gray-800 truncate">
                {{ getMaskedName(person) }}
                <span v-if="person.isMe" class="text-orange-600 text-xs ml-1">(您)</span>
              </div>
              <div v-if="person.note" class="text-xs text-gray-500 mt-0.5">
                <i class="fas fa-sticky-note mr-1"></i>
                {{ person.note }}
              </div>
            </div>
            <span class="text-xs font-bold text-orange-600 bg-orange-50 px-2 py-1 rounded">候補</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Export Button -->
    <div class="bg-white p-4 rounded-xl shadow-sm border border-gray-100">
      <button 
        @click="exportToClipboard"
        class="w-full bg-gray-100 text-gray-700 py-3 rounded-xl font-bold hover:bg-gray-200 active:scale-95 transition-all flex items-center justify-center gap-2"
      >
        <i class="fas fa-copy"></i>
        匯出報名狀態
      </button>
    </div>

    <!-- Edit Note Modal -->
    <div 
      v-if="editingNote"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      @click.self="cancelNoteEdit"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md shadow-2xl">
        <h3 class="text-xl font-bold text-gray-800 mb-4">
          <i class="fas fa-edit mr-2 text-blue-600"></i>
          編輯報名備註
        </h3>
        
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-2">備註內容</label>
          <input 
            v-model="editingNoteValue"
            type="text"
            placeholder="例如：幫朋友報名、攜帶裝備等"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none"
            @keyup.enter="saveNote"
          >
        </div>

        <div class="flex gap-3">
          <button 
            @click="cancelNoteEdit"
            class="flex-1 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
          >
            取消
          </button>
          <button 
            @click="saveNote"
            class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            儲存
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

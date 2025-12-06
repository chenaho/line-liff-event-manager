<script setup>
import { ref, computed, nextTick, onMounted } from 'vue'
import { useEventStore } from '../stores/event'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'

const props = defineProps(['event', 'status'])
const eventStore = useEventStore()
const authStore = useAuthStore()
const { showToast } = useToast()

const content = ref('')
const showDialog = ref(false)
const messagesContainer = ref(null)
const likedMessages = ref(new Set())

const messages = computed(() => {
  if (!props.status || !props.status.records) return []
  return props.status.records.filter(r => r.type === 'MEMO')
})

const getAvatarUrl = (displayName) => {
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(displayName || 'User')}`
}

const isMyMessage = (message) => {
  return message.userId === authStore.user?.lineUserId
}

const formatTime = (timestamp) => {
  if (!timestamp) return 'Just now'
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-TW', { hour: '2-digit', minute: '2-digit' })
}

const openDialog = () => {
  showDialog.value = true
}

const closeDialog = () => {
  showDialog.value = false
  content.value = ''
}

const submitMemo = async () => {
  if (!content.value.trim()) {
    showToast('請輸入留言內容')
    return
  }
  
  try {
    await eventStore.submitAction(props.event.eventId, 'MEMO', {
      content: content.value,
      userDisplayName: authStore.user?.lineDisplayName
    })
    closeDialog()
    showToast('留言已送出')
    
    // Scroll to bottom after new message
    await nextTick()
    scrollToBottom()
  } catch (e) {
    showToast('發送失敗')
  }
}

const toggleLike = (messageId) => {
  if (likedMessages.value.has(messageId)) {
    likedMessages.value.delete(messageId)
  } else {
    likedMessages.value.add(messageId)
  }
}

const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

onMounted(() => {
  scrollToBottom()
})
</script>

<template>
  <div class="fade-in flex flex-col h-[calc(100vh-200px)]">
    <!-- Header Card -->
    <div class="bg-white p-4 rounded-xl shadow-sm border-l-4 border-purple-500 mb-4">
      <h2 class="text-xl font-bold text-gray-800">{{ event.title }}</h2>
    </div>

    <!-- Messages Container -->
    <div ref="messagesContainer" class="flex-1 overflow-y-auto space-y-4 p-2 hide-scrollbar">
      <div v-if="messages.length === 0" class="text-center text-gray-400 py-8">
        <i class="fas fa-comments text-4xl mb-2"></i>
        <p>還沒有留言，成為第一個留言的人吧！</p>
      </div>

      <div 
        v-for="(msg, idx) in messages" 
        :key="idx" 
        class="flex gap-3"
        :class="isMyMessage(msg) ? 'flex-row-reverse' : ''"
      >
        <!-- Avatar -->
        <img 
          :src="getAvatarUrl(msg.userDisplayName)" 
          class="w-8 h-8 rounded-full bg-gray-200 flex-shrink-0 mt-1"
        >
        
        <!-- Message Content -->
        <div class="max-w-[75%]">
          <div class="text-xs text-gray-400 mb-1" :class="isMyMessage(msg) ? 'text-right' : ''">
            {{ msg.userDisplayName }} • {{ formatTime(msg.timestamp) }}
          </div>
          <div 
            class="p-3 rounded-2xl shadow-sm text-gray-700 text-sm relative border"
            :class="isMyMessage(msg) 
              ? 'bg-purple-50 border-purple-100 rounded-tr-none' 
              : 'bg-white border-gray-100 rounded-tl-none'"
          >
            {{ msg.content }}
          </div>
          
          <!-- Like Button -->
          <div class="flex gap-2 mt-1" :class="isMyMessage(msg) ? 'justify-end' : ''">
            <button 
              @click="toggleLike(idx)"
              class="text-xs flex items-center gap-1 transition-all"
              :class="likedMessages.has(idx) ? 'text-pink-500 scale-110' : 'text-gray-400 hover:text-pink-500'"
            >
              <i :class="likedMessages.has(idx) ? 'fas fa-heart' : 'far fa-heart'"></i>
              <span v-if="likedMessages.has(idx)">1</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Floating Action Button -->
    <div class="fixed bottom-6 right-6 z-10">
      <button 
        @click="openDialog"
        class="w-14 h-14 bg-purple-600 text-white rounded-full shadow-lg flex items-center justify-center text-xl hover:bg-purple-700 active:scale-90 transition-transform"
      >
        <i class="fas fa-pen"></i>
      </button>
    </div>

    <!-- Dialog Modal -->
    <div 
      v-if="showDialog"
      class="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4"
      @click.self="closeDialog"
    >
      <div class="bg-white rounded-xl w-full max-w-sm p-4 shadow-2xl transform transition-all scale-100 fade-in">
        <h3 class="font-bold text-lg mb-2">發表留言</h3>
        <textarea 
          v-model="content"
          class="w-full border rounded-lg p-2 h-24 focus:outline-none focus:ring-2 focus:ring-purple-500" 
          placeholder="寫下你的想法..."
          @keydown.enter.ctrl="submitMemo"
        ></textarea>
        <div class="flex justify-end gap-2 mt-4">
          <button 
            @click="closeDialog" 
            class="px-4 py-2 text-gray-500 hover:bg-gray-100 rounded transition-colors"
          >
            取消
          </button>
          <button 
            @click="submitMemo" 
            class="px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700 transition-colors"
          >
            送出
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

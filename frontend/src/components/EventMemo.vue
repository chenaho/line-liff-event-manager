<script setup>
import { ref, computed, nextTick, onMounted, watch } from 'vue'
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

// Edit state
const editingMessage = ref(null)
const editingContent = ref('')

const messages = computed(() => {
  if (!props.status || !props.status.records) return []
  return props.status.records
    .filter(r => r.type === 'MEMO')
    .map(r => ({
      ...r,
      pictureUrl: r.userPictureUrl || null
    }))
})

const getAvatarUrl = (message) => {
  // Prioritize LINE pictureUrl with /small suffix
  if (message.pictureUrl) {
    return message.pictureUrl + '/small'
  }
  // Fallback to generated avatar
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(message.userDisplayName || 'User')}`
}

const isMyMessage = (message) => {
  return message.userId === authStore.user?.lineUserId
}

const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  return date.toLocaleString('zh-TW', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
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
      userDisplayName: authStore.user?.lineDisplayName,
      userPictureUrl: authStore.user?.pictureUrl
    })
    closeDialog()
    showToast('留言已送出')
    
    // Scroll to bottom after new message
    await nextTick()
    scrollToBottom()
  } catch (e) {
    console.error('Send memo error:', e)
    showToast('發送失敗: ' + (e.response?.data?.error || e.message))
  }
}

// Edit message
const openEditMessage = (message) => {
  if (!isMyMessage(message)) return
  editingMessage.value = message
  editingContent.value = message.content
}

const closeEditMessage = () => {
  editingMessage.value = null
  editingContent.value = ''
}

const saveEditedMessage = async () => {
  if (!editingContent.value.trim()) {
    showToast('請輸入留言內容')
    return
  }
  
  try {
    await eventStore.updateMemoContent(
      props.event.eventId,
      editingMessage.value.id,
      editingContent.value
    )
    showToast('留言已更新')
    closeEditMessage()
  } catch (e) {
    showToast('更新失敗: ' + (e.response?.data?.error || e.message))
  }
}

// Clap reaction
const handleClap = async (message) => {
  try {
    await eventStore.incrementClapCount(props.event.eventId, message.id)
  } catch (e) {
    showToast('鼓掌失敗')
  }
}

// Watch messages and scroll to bottom when new message arrives
watch(() => messages.value.length, () => {
  nextTick(() => scrollToBottom())
})

onMounted(() => {
  scrollToBottom()
})
</script>

<template>
  <div class="flex flex-col">
    <!-- Messages Container -->
    <div 
      ref="messagesContainer"
      class="overflow-y-auto p-4 space-y-4 bg-gray-50"
      style="max-height: 50vh;"
    >
      <div v-if="messages.length === 0" class="text-center text-gray-400 py-8">
        <i class="fas fa-comments text-4xl mb-2"></i>
        <p class="text-sm">尚無留言</p>
      </div>

      <div 
        v-for="msg in messages" 
        :key="msg.timestamp"
        class="flex gap-3"
        :class="isMyMessage(msg) ? 'flex-row-reverse' : ''"
      >
        <!-- Avatar -->
        <img 
          :src="getAvatarUrl(msg)" 
          class="w-8 h-8 rounded-full bg-gray-200 flex-shrink-0 mt-1"
          :alt="msg.userDisplayName"
        >
        
        <!-- Message Content -->
        <div class="max-w-[75%]">
          <div class="text-xs text-gray-400 mb-1" :class="isMyMessage(msg) ? 'text-right' : ''">
            {{ msg.userDisplayName }} · {{ formatTime(msg.timestamp) }}
          </div>
          <div 
            @click="openEditMessage(msg)"
            class="px-4 py-2 rounded-2xl break-words"
            :class="[
              isMyMessage(msg) 
                ? 'bg-blue-500 text-white rounded-br-sm' 
                : 'bg-gray-100 text-gray-800 rounded-bl-sm',
              isMyMessage(msg) ? 'cursor-pointer hover:bg-blue-600' : ''
            ]"
          >
            {{ msg.content }}
          </div>
          
          <!-- Clap Button -->
          <div class="flex gap-2 mt-1" :class="isMyMessage(msg) ? 'justify-end' : ''">
            <button 
              @click="handleClap(msg)"
              class="text-xs flex items-center gap-1 transition-all px-2 py-1 rounded hover:bg-gray-200"
              :class="(msg.clapCount || 0) > 0 ? 'text-orange-500' : 'text-gray-400'"
            >
              <i class="fas fa-hands-clapping"></i>
              <span v-if="(msg.clapCount || 0) > 0">{{ msg.clapCount }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Input Area -->
    <div class="p-4 bg-white border-t border-gray-200">
      <button 
        @click="openDialog"
        class="w-full bg-blue-600 text-white py-3 rounded-xl font-bold shadow-lg hover:bg-blue-700 active:scale-95 transition-transform"
      >
        <i class="fas fa-plus-circle mr-2"></i>
        發表留言
      </button>
    </div>

    <!-- New Message Dialog -->
    <div 
      v-if="showDialog"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      @click.self="closeDialog"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md shadow-2xl">
        <h3 class="text-xl font-bold text-gray-800 mb-4">
          <i class="fas fa-comment-dots mr-2 text-blue-600"></i>
          發表留言
        </h3>
        
        <textarea 
          v-model="content"
          placeholder="輸入您的留言..."
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none resize-none"
          rows="4"
        ></textarea>

        <div class="flex gap-3 mt-4">
          <button 
            @click="closeDialog"
            class="flex-1 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
          >
            取消
          </button>
          <button 
            @click="submitMemo"
            class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            送出
          </button>
        </div>
      </div>
    </div>

    <!-- Edit Message Dialog -->
    <div 
      v-if="editingMessage"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
      @click.self="closeEditMessage"
    >
      <div class="bg-white rounded-xl p-6 w-full max-w-md shadow-2xl">
        <h3 class="text-xl font-bold text-gray-800 mb-4">
          <i class="fas fa-edit mr-2 text-blue-600"></i>
          編輯留言
        </h3>
        
        <textarea 
          v-model="editingContent"
          placeholder="輸入您的留言..."
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none resize-none"
          rows="4"
        ></textarea>

        <div class="flex gap-3 mt-4">
          <button 
            @click="closeEditMessage"
            class="flex-1 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
          >
            取消
          </button>
          <button 
            @click="saveEditedMessage"
            class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            儲存
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

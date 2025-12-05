<script setup>
import { ref } from 'vue'
import { useEventStore } from '../stores/event'
import { useAuthStore } from '../stores/auth'

const props = defineProps(['event', 'status'])
const eventStore = useEventStore()
const authStore = useAuthStore()

const content = ref('')

const submitMemo = async () => {
  if (!content.value.trim()) return
  
  try {
    await eventStore.submitAction(props.event.eventId, 'MEMO', {
      content: content.value,
      userDisplayName: authStore.user?.lineDisplayName
    })
    content.value = ''
  } catch (e) {
    alert('Failed to post memo')
  }
}
</script>

<template>
  <div class="flex flex-col h-[calc(100vh-200px)]">
    <div class="flex-1 overflow-y-auto space-y-4 p-2">
      <div v-if="status && status.records">
        <div v-for="(rec, idx) in status.records.filter(r => r.type === 'MEMO')" :key="idx" class="flex gap-2">
          <div class="w-8 h-8 bg-gray-200 rounded-full flex-shrink-0"></div>
          <div class="bg-white p-3 rounded-lg rounded-tl-none shadow-sm max-w-[80%]">
            <p class="text-xs text-gray-500 mb-1">{{ rec.userDisplayName }}</p>
            <p>{{ rec.content }}</p>
          </div>
        </div>
      </div>
    </div>

    <div class="mt-4 flex gap-2">
      <input 
        v-model="content" 
        type="text" 
        placeholder="Say something..." 
        class="flex-1 border border-gray-300 rounded-full px-4 py-2 focus:outline-none focus:border-blue-500"
        @keyup.enter="submitMemo"
      >
      <button @click="submitMemo" class="bg-blue-600 text-white rounded-full w-10 h-10 flex items-center justify-center">
        ➤
      </button>
    </div>
  </div>
</template>

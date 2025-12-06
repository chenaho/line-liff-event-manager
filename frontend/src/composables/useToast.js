import { ref } from 'vue'

const toastMessage = ref('')
const toastVisible = ref(false)
let toastTimeout = null

export function useToast() {
    const showToast = (message, duration = 2000) => {
        // Clear existing timeout
        if (toastTimeout) {
            clearTimeout(toastTimeout)
        }

        toastMessage.value = message
        toastVisible.value = true

        toastTimeout = setTimeout(() => {
            toastVisible.value = false
        }, duration)
    }

    return {
        toastMessage,
        toastVisible,
        showToast
    }
}

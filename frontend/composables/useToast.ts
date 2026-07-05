import { ref } from 'vue'

interface Toast {
  id: number
  message: string
  type: 'success' | 'error' | 'info'
  visible: boolean
}

const toasts = ref<Toast[]>([])
let nextId = 0

export function useToast() {
  const showToast = (
    message: string,
    type: 'success' | 'error' | 'info' = 'info',
    duration = 3000,
  ) => {
    const id = nextId++
    const toast: Toast = { id, message, type, visible: true }
    toasts.value.push(toast)

    setTimeout(() => {
      const index = toasts.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        toasts.value[index].visible = false
        setTimeout(() => {
          toasts.value = toasts.value.filter((t) => t.id !== id)
        }, 300)
      }
    }, duration)
  }

  const removeToast = (id: number) => {
    const index = toasts.value.findIndex((t) => t.id === id)
    if (index !== -1) {
      toasts.value[index].visible = false
      setTimeout(() => {
        toasts.value = toasts.value.filter((t) => t.id !== id)
      }, 300)
    }
  }

  return { toasts, showToast, removeToast }
}

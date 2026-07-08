<template>
  <div class="fixed bottom-6 left-1/2 -translate-x-1/2 z-[9999] flex flex-col items-center gap-2 pointer-events-none">
    <div
      v-for="toast in toasts"
      :key="toast.id"
      class="pointer-events-auto px-4 py-3 rounded-[var(--radius-md)] shadow-lg text-sm flex items-center gap-3 transition-all duration-300"
      :class="[
        toast.visible ? 'opacity-100 translate-x-0' : 'opacity-0 translate-x-4',
        toastBg[toast.type],
      ]"
    >
      <span class="flex-shrink-0">
        <!-- Success icon -->
        <svg v-if="toast.type === 'success'" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-white">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
        </svg>
        <!-- Error icon -->
        <svg v-else-if="toast.type === 'error'" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-white">
          <circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/>
        </svg>
        <!-- Info icon -->
        <svg v-else-if="toast.type === 'info'" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-white">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/>
        </svg>
      </span>
      <span class="flex-1 text-white">{{ toast.message }}</span>
      <button class="flex-shrink-0 text-white/70 hover:text-white transition-colors" @click="removeToast(toast.id)">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useToast } from '~/composables/useToast'

const { toasts, removeToast } = useToast()

const toastBg: Record<string, string> = {
  success: 'bg-[var(--color-success)]',
  error: 'bg-[var(--color-danger)]',
  info: 'bg-[var(--color-primary)]',
}
</script>

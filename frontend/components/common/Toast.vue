<template>
  <div class="fixed top-4 right-4 z-[9999] flex flex-col gap-2 pointer-events-none">
    <div
      v-for="toast in toasts"
      :key="toast.id"
      class="pointer-events-auto px-4 py-3 rounded-[var(--radius-md)] shadow-lg text-sm flex items-center gap-3 transition-all duration-300"
      :class="[
        toast.visible ? 'opacity-100 translate-x-0' : 'opacity-0 translate-x-4',
        toastBg[toast.type],
      ]"
    >
      <span class="flex-shrink-0">{{ toastIcon[toast.type] }}</span>
      <span class="flex-1 text-white">{{ toast.message }}</span>
      <button class="flex-shrink-0 text-white/70 hover:text-white" @click="removeToast(toast.id)">
        &#10005;
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useToast } from '~/composables/useToast'

const { toasts, removeToast } = useToast()

const toastBg: Record<string, string> = {
  success: 'bg-green-600',
  error: 'bg-red-600',
  info: 'bg-blue-600',
}

const toastIcon: Record<string, string> = {
  success: '✔',
  error: '✘',
  info: 'ℹ',
}
</script>

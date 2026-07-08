<template>
  <div class="flex flex-col items-center justify-center py-16 text-center">
    <div class="mb-4 text-[var(--color-danger)]/60">
      <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10"/>
        <line x1="12" y1="8" x2="12" y2="12"/>
        <line x1="12" y1="16" x2="12.01" y2="16"/>
      </svg>
    </div>
    <p class="text-[var(--color-text-secondary)] text-sm mb-4">{{ message }}</p>
    <button
      v-if="onRetry"
      class="px-6 py-2 bg-[var(--color-primary)] text-white text-sm font-medium rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-all duration-[var(--transition-normal)] active:scale-[0.98] shadow-sm shadow-[var(--color-primary)]/25"
      @click="onRetry"
    >
      重试
    </button>
  </div>
</template>

<script setup lang="ts">
interface Props {
  message: string
  onRetry?: () => void
}

defineProps<Props>()
</script>

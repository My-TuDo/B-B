<template>
  <div class="flex flex-col items-center justify-center py-16 text-center">
    <div class="text-5xl mb-4 opacity-50">&#9888;</div>
    <p class="text-[var(--color-text-secondary)] text-base mb-4">{{ message }}</p>
    <button
      v-if="onRetry"
      class="px-6 py-2 bg-[var(--color-primary)] text-white text-sm rounded-[var(--radius-full)] hover:bg-[var(--color-primary-hover)] transition-colors"
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

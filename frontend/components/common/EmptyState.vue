<template>
  <div class="flex flex-col items-center justify-center py-16 text-center">
    <div class="text-5xl mb-4 opacity-50">{{ icon }}</div>
    <p class="text-[var(--color-text-secondary)] text-base">{{ message }}</p>
    <slot />
  </div>
</template>

<script setup lang="ts">
interface Props {
  message: string
  icon?: string
}

withDefaults(defineProps<Props>(), {
  icon: '📭',
})
</script>

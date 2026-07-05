<template>
  <div class="flex items-center justify-center py-12">
    <div
      class="animate-spin rounded-full border-2 border-[var(--color-border)]"
      :class="sizeClasses"
      style="border-top-color: var(--color-primary)"
    ></div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm': return 'w-5 h-5'
    case 'lg': return 'w-12 h-12'
    default: return 'w-8 h-8'
  }
})
</script>

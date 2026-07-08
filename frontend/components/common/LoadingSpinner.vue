<template>
  <div v-if="variant === 'spinner'" class="flex items-center justify-center py-12">
    <div
      class="animate-spin rounded-full border-2 border-[var(--color-border)]"
      :class="sizeClasses"
      style="border-top-color: var(--color-primary)"
    ></div>
  </div>
  <div v-else class="w-full py-12 px-4">
    <div class="animate-pulse space-y-4">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
        <div v-for="i in skeletonCount" :key="i">
          <div class="aspect-video bg-[var(--color-surface-hover)] rounded-[var(--radius-md)]"></div>
          <div class="mt-2.5 flex gap-2.5">
            <div class="w-8 h-8 rounded-full bg-[var(--color-surface-hover)] flex-shrink-0"></div>
            <div class="flex-1 space-y-2">
              <div class="h-3.5 bg-[var(--color-surface-hover)] rounded w-full"></div>
              <div class="h-3.5 bg-[var(--color-surface-hover)] rounded w-2/3"></div>
              <div class="h-2.5 bg-[var(--color-surface-hover)] rounded w-1/2"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  size?: 'sm' | 'md' | 'lg'
  variant?: 'spinner' | 'skeleton'
  skeletonCount?: number
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  variant: 'spinner',
  skeletonCount: 4,
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm': return 'w-5 h-5'
    case 'lg': return 'w-12 h-12'
    default: return 'w-8 h-8'
  }
})
</script>

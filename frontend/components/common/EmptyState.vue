<template>
  <div class="flex flex-col items-center justify-center py-16 text-center">
    <div class="mb-4 text-[var(--color-text-secondary)]/30">
      <!-- Default: empty box -->
      <svg v-if="iconType === 'empty'" xmlns="http://www.w3.org/2000/svg" width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
        <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
        <line x1="12" y1="22.08" x2="12" y2="12"/>
      </svg>
      <!-- No search results -->
      <svg v-else-if="iconType === 'search'" xmlns="http://www.w3.org/2000/svg" width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="11" cy="11" r="8"/>
        <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        <line x1="8" y1="11" x2="14" y2="11"/>
      </svg>
      <!-- No videos / media -->
      <svg v-else-if="iconType === 'video'" xmlns="http://www.w3.org/2000/svg" width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
        <line x1="8" y1="21" x2="16" y2="21"/>
        <line x1="12" y1="17" x2="12" y2="21"/>
      </svg>
      <!-- No history / clock -->
      <svg v-else-if="iconType === 'history'" xmlns="http://www.w3.org/2000/svg" width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10"/>
        <polyline points="12 6 12 12 16 14"/>
      </svg>
      <!-- No documents / drafts -->
      <svg v-else-if="iconType === 'document'" xmlns="http://www.w3.org/2000/svg" width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
        <polyline points="14 2 14 8 20 8"/>
        <line x1="16" y1="13" x2="8" y2="13"/>
        <line x1="16" y1="17" x2="8" y2="17"/>
      </svg>
    </div>
    <p class="text-[var(--color-text-secondary)] text-sm">{{ message }}</p>
    <slot />
  </div>
</template>

<script setup lang="ts">
interface Props {
  message: string
  icon?: string
}

const props = withDefaults(defineProps<Props>(), {
  icon: 'empty',
})

const iconType = computed(() => {
  // Map legacy emoji icons to our SVG types
  const map: Record<string, string> = {
    '📭': 'empty',
    '📺': 'video',
    '🔍': 'search',
    '🕐': 'history',
    '📝': 'document',
    '📹': 'video',
  }
  return map[props.icon] || props.icon
})
</script>

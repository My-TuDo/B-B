<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="fixed inset-0 z-[200] flex items-center justify-center p-4">
        <!-- Darkened backdrop -->
        <div class="absolute inset-0 bg-black/60" @click="$emit('close')"></div>
        <!-- Centered card -->
        <div
          class="relative bg-[var(--color-surface)] border border-[var(--color-border)] rounded-[var(--radius-lg)] shadow-2xl p-6 w-full"
          :class="width"
        >
          <!-- Title bar -->
          <div v-if="title || $slots.header" class="flex items-center justify-between mb-4">
            <h3 class="text-base font-semibold text-[var(--color-text)]">{{ title }}</h3>
            <button
              class="text-[var(--color-text-secondary)] hover:text-[var(--color-text)] transition-colors p-0.5"
              @click="$emit('close')"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>
          <!-- Content slot -->
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  visible: boolean
  title?: string
  width?: string
}>(), {
  width: 'max-w-sm',
})

defineEmits<{
  close: []
}>()
</script>

<style>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 200ms ease;
}
.modal-enter-active > div:last-child,
.modal-leave-active > div:last-child {
  transition: opacity 200ms ease, transform 200ms ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from > div:last-child,
.modal-leave-to > div:last-child {
  opacity: 0;
  transform: scale(0.95);
}
</style>

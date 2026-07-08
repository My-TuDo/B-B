<template>
  <div class="flex items-center justify-center min-h-[60vh]">
    <div class="flex flex-col items-center text-center">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="64"
        height="64"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="text-[var(--color-text-secondary)]/30 mb-4"
      >
        <path d="M12 20h9" />
        <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
      </svg>
      <p class="text-[var(--color-text-secondary)] text-sm">创作中心即将上线</p>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
})
</script>

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
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
      <p class="text-[var(--color-text-secondary)] text-sm">审核功能即将上线</p>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 relative overflow-hidden">
    <!-- Background gradient -->
    <div class="absolute inset-0 bg-gradient-to-br from-[var(--color-primary)]/5 via-transparent to-[var(--color-primary)]/5" />
    <div class="w-full max-w-md relative">
      <div class="text-center mb-8">
        <NuxtLink to="/" class="text-4xl font-bold text-[var(--color-primary)] font-display">B-B</NuxtLink>
      </div>
      <div class="bg-[var(--color-surface)] rounded-[var(--radius-lg)] p-8 border border-[var(--color-border)]/30 shadow-xl shadow-black/5">
        <slot />
      </div>
    </div>
    <Toast />
  </div>
</template>

<script setup lang="ts">
import Toast from '~/components/common/Toast.vue'
</script>

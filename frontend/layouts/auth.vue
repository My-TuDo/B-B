<template>
  <div class="min-h-screen flex items-center justify-center bg-[var(--color-bg)] p-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <NuxtLink to="/" class="text-4xl font-bold text-[var(--color-primary)]">B-B</NuxtLink>
      </div>
      <div class="bg-[var(--color-surface)] rounded-[var(--radius-lg)] p-8 border border-[var(--color-border)]">
        <slot />
      </div>
    </div>
    <Toast />
  </div>
</template>

<script setup lang="ts">
import Toast from '~/components/common/Toast.vue'
</script>

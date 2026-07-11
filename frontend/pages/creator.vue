<template>
  <div class="max-w-6xl mx-auto">
    <h1 class="text-2xl font-bold text-[var(--color-text)] mb-6">创作中心</h1>
    <CreatorPanel :user-id="userStore.userInfo?.id || 0" />
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '~/stores/userStore'
const userStore = useUserStore()
definePageMeta({ middleware: 'auth' })
</script>

<template>
  <div v-if="videos.length > 0" class="relative mb-8 rounded-2xl overflow-hidden bg-[var(--color-surface)] border border-[var(--color-border)]/30">
    <div class="relative w-full" style="aspect-ratio: 16/5">
      <NuxtLink v-for="(v, i) in videos" :key="v.id" :to="`/video/${v.id}`"
        class="absolute inset-0 transition-opacity duration-500"
        :class="i === current ? 'opacity-100 z-10' : 'opacity-0 z-0'">
        <img v-if="v.cover_url" :src="v.cover_url" class="w-full h-full object-cover" />
        <div v-else class="w-full h-full bg-gradient-to-br from-[var(--color-primary)]/20 to-[var(--color-surface)]" />
        <div class="absolute inset-0 bg-gradient-to-t from-black/70 via-transparent to-transparent" />
        <div class="absolute bottom-0 left-0 right-0 p-6">
          <h2 class="text-xl font-bold text-white line-clamp-1">{{ v.title }}</h2>
          <p class="text-sm text-white/70 mt-1">{{ v.user?.nickname || v.user?.username }} · {{ formatNum(v.views) }} 播放</p>
        </div>
      </NuxtLink>
    </div>
    <!-- Dots -->
    <div class="absolute bottom-3 right-4 z-20 flex gap-1.5">
      <button v-for="(_, i) in videos" :key="i"
        class="w-2 h-2 rounded-full transition-all"
        :class="i === current ? 'bg-white w-6' : 'bg-white/40'"
        @click="current = i" />
    </div>
    <!-- Arrows -->
    <button class="absolute left-3 top-1/2 -translate-y-1/2 z-20 w-10 h-10 rounded-full bg-white/20 hover:bg-white/30 flex items-center justify-center transition-colors" @click="prev">
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
    </button>
    <button class="absolute right-3 top-1/2 -translate-y-1/2 z-20 w-10 h-10 rounded-full bg-white/20 hover:bg-white/30 flex items-center justify-center transition-colors" @click="next">
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import type { VideoInfo } from '~/types'

defineProps<{ videos: VideoInfo[] }>()
const current = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

function prev() { current.value = (current.value - 1 + 4) % 4 }
function next() { current.value = (current.value + 1) % 4 }
function formatNum(n: number) { return n >= 10000 ? `${(n / 10000).toFixed(1)}万` : String(n) }

onMounted(() => { timer = setInterval(next, 5000) })
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

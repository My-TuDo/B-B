<template>
  <NuxtLink :to="`/video/${video.id}`" class="group block">
    <!-- Thumbnail -->
    <div class="relative w-full bg-[var(--color-surface)] rounded-[var(--radius-md)] overflow-hidden transition-transform duration-[var(--transition-normal)] group-hover:scale-[1.02] group-hover:shadow-[var(--shadow-hover)]" style="aspect-ratio: 16/9">
      <!-- Cover image or gradient placeholder -->
      <div
        v-if="video.cover_url"
        class="w-full h-full"
      >
        <img :src="video.cover_url" :alt="video.title" class="w-full h-full object-cover" />
      </div>
      <div
        v-else
        class="w-full h-full flex items-center justify-center"
        style="background: linear-gradient(135deg, var(--color-surface) 0%, var(--color-primary) 100%); opacity: 0.3;"
      >
        <span class="text-4xl">&#9654;</span>
      </div>

      <!-- Duration badge -->
      <div
        v-if="video.duration > 0"
        class="absolute bottom-1 right-1 px-1.5 py-0.5 bg-black/80 text-white text-xs rounded"
      >
        {{ formatDuration(video.duration) }}
      </div>

      <!-- Views badge -->
      <div class="absolute bottom-1 left-1 px-1.5 py-0.5 bg-black/80 text-white text-xs rounded">
        {{ formatViews(video.views) }} 播放
      </div>
    </div>

    <!-- Info -->
    <div class="mt-2 flex gap-2">
      <!-- Uploader avatar -->
      <div class="flex-shrink-0 w-6 h-6 rounded-full bg-[var(--color-primary)] flex items-center justify-center text-white text-[10px] font-bold overflow-hidden">
        <img v-if="video.user?.avatar" :src="video.user.avatar" class="w-full h-full object-cover" />
        <span v-else>{{ (video.user?.nickname || video.user?.username || 'U').charAt(0) }}</span>
      </div>

      <!-- Text info -->
      <div class="flex-1 min-w-0">
        <h3 class="text-sm text-[var(--color-text)] leading-snug line-clamp-2 group-hover:text-[var(--color-primary)] transition-colors">
          {{ video.title }}
        </h3>
        <p class="mt-1 text-xs text-[var(--color-text-secondary)]">
          {{ video.user?.nickname || video.user?.username || '未知用户' }} · {{ formatViews(video.views) }} 播放 · {{ formatTime(video.created_at) }}
        </p>
      </div>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
import type { VideoInfo } from '~/types'

interface Props {
  video: VideoInfo
}

defineProps<Props>()

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

function formatViews(views: number): string {
  if (views >= 10000) {
    return `${(views / 10000).toFixed(1)}万`
  }
  return String(views)
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 30) return `${days}天前`
  if (days < 365) return `${Math.floor(days / 30)}个月前`
  return `${Math.floor(days / 365)}年前`
}
</script>

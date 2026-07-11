<template>
  <NuxtLink :to="`/video/${video.id}`" class="group block">
    <!-- Thumbnail -->
    <div
      class="relative w-full bg-[var(--color-surface)] rounded-[var(--radius-md)] overflow-hidden ring-1 ring-[var(--color-border)]/30 transition-all duration-300 group-hover:shadow-[var(--shadow-hover)] group-hover:ring-[var(--color-primary)]/20"
      style="aspect-ratio: 16/9"
    >
      <!-- Cover image or gradient placeholder -->
      <div
        v-if="video.cover_url"
        class="w-full h-full"
      >
        <img
          :src="video.cover_url"
          :alt="video.title"
          class="w-full h-full object-cover transition-transform duration-[var(--transition-slow)] group-hover:scale-105"
          loading="lazy"
        />
      </div>
      <div
        v-else
        class="w-full h-full flex items-center justify-center bg-gradient-to-br from-[var(--color-surface)] via-[var(--color-surface-hover)] to-[var(--color-primary)]/10"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="currentColor" class="text-[var(--color-primary)]/25">
          <polygon points="8,5 19,12 8,19" />
        </svg>
      </div>

      <!-- Hover overlay with play icon -->
      <div class="absolute inset-0 bg-black/0 group-hover:bg-black/40 transition-colors duration-300 flex items-center justify-center">
        <div class="w-12 h-12 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center opacity-0 group-hover:opacity-100 transition-all duration-300 scale-75 group-hover:scale-100">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="22"
            height="22"
            viewBox="0 0 24 24"
            fill="white"
            class="ml-0.5"
          >
            <polygon points="8,5 19,12 8,19" />
          </svg>
        </div>
      </div>

      <!-- Duration badge -->
      <div
        v-if="video.duration > 0"
        class="absolute bottom-2 right-2 px-1.5 py-0.5 bg-[var(--color-badge-bg)] text-white text-[11px] rounded font-mono-data tracking-tight"
      >
        {{ formatDuration(video.duration) }}
      </div>

      <!-- Views badge -->
      <div class="absolute bottom-2 left-2 px-1.5 py-0.5 bg-[var(--color-badge-bg)] text-white text-[11px] rounded leading-tight">
        {{ formatViews(video.views) }} 播放
      </div>
    </div>

    <!-- Info -->
    <div class="mt-3 flex gap-3">
      <!-- Uploader avatar -->
      <NuxtLink
        :to="`/user/${video.user?.id}`"
        class="flex-shrink-0 w-9 h-9 rounded-full bg-[var(--color-surface-hover)] flex items-center justify-center text-[var(--color-primary)] text-xs font-bold overflow-hidden ring-1 ring-[var(--color-border)]/30 group-hover:ring-[var(--color-primary)]/40 transition-all"
        @click.stop
      >
        <img v-if="video.user?.avatar" :src="video.user.avatar" class="w-full h-full object-cover" />
        <span v-else>{{ (video.user?.nickname || video.user?.username || 'U').charAt(0) }}</span>
      </NuxtLink>

      <!-- Text info -->
      <div class="flex-1 min-w-0">
        <h3 class="text-sm text-[var(--color-text)] font-medium leading-snug line-clamp-2 group-hover:text-[var(--color-primary)] transition-colors duration-200">
          {{ video.title }}
        </h3>
        <p class="mt-1 text-xs text-[var(--color-text-secondary)] leading-relaxed truncate">
          {{ video.user?.nickname || video.user?.username || '未知用户' }}
        </p>
        <p class="text-[11px] text-[var(--color-text-secondary)]/70 leading-relaxed mt-0.5">
          {{ formatViews(video.views) }} 播放 <span class="mx-1 text-[var(--color-text-secondary)]/30">·</span> {{ formatTime(video.created_at) }}
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

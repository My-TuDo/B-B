<template>
  <aside
    class="fixed left-0 bg-[var(--color-surface)] border-r border-[var(--color-border)] flex flex-col overflow-hidden transition-all duration-200 z-40"
    :style="[sidebarStyle, { display: 'flex', flexDirection: 'column' }]"
    @mouseenter="hovering = true"
    @mouseleave="hovering = false"
  >
    <!-- Scrollable top area -->
    <nav class="flex-1 min-h-0 overflow-y-auto overflow-x-hidden py-2">
      <NuxtLink
        to="/"
        class="sidebar-item"
        :class="{ 'sidebar-item--active': route.path === '/' && !route.query.category_id }"
      >
        <IconHome class="sidebar-icon" />
        <span v-show="isExpanded" class="sidebar-label">首页</span>
      </NuxtLink>

      <NuxtLink
        to="/ranking"
        class="sidebar-item"
        :class="{ 'sidebar-item--active': route.path === '/ranking' }"
      >
        <IconRanking class="sidebar-icon" />
        <span v-show="isExpanded" class="sidebar-label">排行榜</span>
      </NuxtLink>

      <div class="my-2 mx-4 border-t border-[var(--color-border)]"></div>

      <!-- Categories accordion -->
      <button
        v-show="isExpanded"
        class="sidebar-item w-full"
        @click="categoriesOpen = !categoriesOpen"
      >
        <IconFolder class="sidebar-icon" />
        <span class="sidebar-label text-xs font-medium text-[var(--color-text-secondary)] uppercase tracking-wider">分类</span>
        <IconChevron class="sidebar-chevron" :class="{ 'rotate-90': categoriesOpen }" />
      </button>
      <template v-if="isExpanded && categoriesOpen">
        <NuxtLink
          v-for="cat in categories"
          :key="cat.id"
          :to="`/?category_id=${cat.id}`"
          class="sidebar-item"
          :class="{ 'sidebar-item--active': route.query.category_id === String(cat.id) }"
        >
          <component :is="categoryIcon(cat.slug)" class="sidebar-icon" />
          <span class="sidebar-label">{{ cat.name }}</span>
        </NuxtLink>
      </template>

      <!-- Creator center -->
    </nav>

    <!-- Pinned bottom area -->
    <div v-if="userStore.isLoggedIn" class="flex-shrink-0 border-t border-[var(--color-border)] py-1">
      <NuxtLink
        :to="`/user/${userStore.userInfo?.id}`"
        class="sidebar-item"
        :class="{ 'sidebar-item--active': route.path.startsWith('/user/') }"
      >
        <IconUser class="sidebar-icon" />
        <span v-show="isExpanded" class="sidebar-label">个人中心</span>
      </NuxtLink>

      <NuxtLink
        to="/drafts"
        class="sidebar-item"
        :class="{ 'sidebar-item--active': route.path === '/drafts' }"
      >
        <IconDraft class="sidebar-icon" />
        <span v-show="isExpanded" class="sidebar-label">稿件管理</span>
      </NuxtLink>

      <NuxtLink
        to="/history"
        class="sidebar-item"
        :class="{ 'sidebar-item--active': route.path === '/history' }"
      >
        <IconHistory class="sidebar-icon" />
        <span v-show="isExpanded" class="sidebar-label">观看历史</span>
      </NuxtLink>
    </div>

    <!-- Collapse button -->
    <div class="flex-shrink-0 border-t border-[var(--color-border)] py-1">
      <button class="sidebar-item w-full" @click="$emit('toggle')">
        <IconCollapse class="sidebar-icon" :class="{ 'rotate-180': !collapsed }" />
        <span v-show="isExpanded" class="sidebar-label text-[var(--color-text-secondary)]">收起</span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { useApi } from '~/composables/useApi'
import { useUserStore } from '~/stores/userStore'
import type { Category } from '~/types'
import { h } from 'vue'
import type { FunctionalComponent, SVGAttributes } from 'vue'

const props = defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  toggle: []
}>()

const route = useRoute()
const userStore = useUserStore()
const { get } = useApi()

const hovering = ref(false)
const categoriesOpen = ref(false)
const categories = ref<Category[]>([])

const isExpanded = computed(() => !props.collapsed || hovering.value)

const sidebarStyle = computed(() => ({
  top: 'var(--header-height)',
  bottom: '0',
  height: 'calc(100vh - var(--header-height))',
  width: isExpanded.value ? '240px' : '72px',
}))

// ===== Line-style SVG Icons =====
const strokeProps = {
  xmlns: 'http://www.w3.org/2000/svg',
  width: '22',
  height: '22',
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  'stroke-width': '1.75',
  'stroke-linecap': 'round' as const,
  'stroke-linejoin': 'round' as const,
}

const IconHome: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'm3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z' }), h('polyline', { points: '9 22 9 12 15 12 15 22' })])

const IconUser: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2' }), h('circle', { cx: '12', cy: '7', r: '4' })])

const IconUpload: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4' }), h('polyline', { points: '17 8 12 3 7 8' }), h('line', { x1: '12', y1: '3', x2: '12', y2: '15' })])

const IconDraft: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z' }), h('polyline', { points: '14 2 14 8 20 8' }), h('line', { x1: '16', y1: '13', x2: '8', y2: '13' }), h('line', { x1: '16', y1: '17', x2: '8', y2: '17' })])

const IconCollapse: FunctionalComponent = (p: SVGAttributes) =>
  h('svg', { ...strokeProps, class: ['transition-transform duration-200', (p as any).class] }, [h('polyline', { points: '15 18 9 12 15 6' })])

const IconFolder: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z' })])

const IconChevron: FunctionalComponent = (p: SVGAttributes) =>
  h('svg', { ...strokeProps, width: '16', height: '16', class: ['transition-transform duration-200 flex-shrink-0', (p as any).class] }, [h('polyline', { points: '9 18 15 12 9 6' })])

const IconAnime: FunctionalComponent = () =>
  h('svg', strokeProps, [h('rect', { x: '2', y: '3', width: '20', height: '14', rx: '2', ry: '2' }), h('line', { x1: '8', y1: '21', x2: '16', y2: '21' }), h('line', { x1: '12', y1: '17', x2: '12', y2: '21' })])

const IconMusic: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'M9 18V5l12-2v13' }), h('circle', { cx: '6', cy: '18', r: '3' }), h('circle', { cx: '18', cy: '16', r: '3' })])

const IconGame: FunctionalComponent = () =>
  h('svg', strokeProps, [h('line', { x1: '6', y1: '11', x2: '10', y2: '11' }), h('line', { x1: '8', y1: '9', x2: '8', y2: '13' }), h('line', { x1: '15', y1: '12', x2: '15.01', y2: '12' }), h('line', { x1: '18', y1: '10', x2: '18.01', y2: '10' }), h('path', { d: 'M17.32 5H6.68a4 4 0 0 0-3.978 3.59c-.006.052-.01.101-.017.152C2.604 9.416 2 14.456 2 16a3 3 0 0 0 3 3c1 0 1.5-.5 2-1l1.414-1.414A2 2 0 0 1 9.828 16h4.344a2 2 0 0 1 1.414.586L17 18c.5.5 1 1 2 1a3 3 0 0 0 3-3c0-1.545-.604-6.584-.685-7.258-.007-.05-.011-.1-.017-.151A4 4 0 0 0 17.32 5z' })])

const IconKnowledge: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'M4 19.5A2.5 2.5 0 0 1 6.5 17H20' }), h('path', { d: 'M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z' })])

const IconLife: FunctionalComponent = () =>
  h('svg', strokeProps, [h('circle', { cx: '12', cy: '12', r: '10' }), h('path', { d: 'M12 6v6l4 2' })])

const IconMovie: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'm13.18 16.26-3.01 1.85c-.75.46-1.67-.1-1.67-.94V6.83c0-.84.92-1.4 1.67-.94l3.01 1.85 3.01 1.85c.75.46.75 1.62 0 2.08l-3.01 1.85' })])

const IconTech: FunctionalComponent = () =>
  h('svg', strokeProps, [h('rect', { x: '2', y: '2', width: '20', height: '8', rx: '2', ry: '2' }), h('rect', { x: '2', y: '14', width: '20', height: '8', rx: '2', ry: '2' }), h('line', { x1: '6', y1: '6', x2: '6.01', y2: '6' }), h('line', { x1: '6', y1: '18', x2: '6.01', y2: '18' })])

const IconDefault: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z' }), h('polyline', { points: '14 2 14 8 20 8' })])

const IconRanking: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'M6 9H4.5a2.5 2.5 0 0 1 0-5C7 4 7 7 7 7' }), h('path', { d: 'M18 9h1.5a2.5 2.5 0 0 0 0-5C17 4 17 7 17 7' }), h('path', { d: 'M4 22h16' }), h('path', { d: 'M10 22V8c0-1.1.9-2 2-2s2 .9 2 2v14' })])

const IconHistory: FunctionalComponent = () =>
  h('svg', strokeProps, [h('circle', { cx: '12', cy: '12', r: '10' }), h('polyline', { points: '12 6 12 12 16 14' })])

const IconCreator: FunctionalComponent = () =>
  h('svg', strokeProps, [h('path', { d: 'M12 20h9' }), h('path', { d: 'M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z' })])

const iconMap: Record<string, FunctionalComponent> = {
  anime: IconAnime,
  music: IconMusic,
  game: IconGame,
  knowledge: IconKnowledge,
  life: IconLife,
  movie: IconMovie,
  tech: IconTech,
}

function categoryIcon(slug: string): FunctionalComponent {
  return iconMap[slug] || IconDefault
}

onMounted(async () => {
  try {
    categories.value = await get<Category[]>('/api/v1/categories/')
  } catch {
    // non-critical
  }
})
</script>

<style scoped>
.sidebar-item {
  display: flex;
  align-items: center;
  gap: 16px;
  height: 40px;
  padding: 0 16px;
  margin: 1px 8px;
  border-radius: var(--radius-md);
  text-decoration: none;
  color: var(--color-text);
  font-size: 14px;
  transition: background-color var(--transition-fast), color var(--transition-fast);
  white-space: nowrap;
  overflow: hidden;
  position: relative;
}
.sidebar-item:hover {
  background-color: var(--color-surface-hover);
  color: var(--color-text);
}
.sidebar-item--active {
  background-color: var(--color-primary-soft);
  color: var(--color-primary);
  font-weight: 500;
}
.sidebar-item--active:hover {
  background-color: var(--color-primary-softer);
  color: var(--color-primary);
}
.sidebar-item--active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 18px;
  border-radius: 0 3px 3px 0;
  background-color: var(--color-primary);
}
.sidebar-icon {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 18px;
}
.sidebar-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sidebar-chevron {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  color: var(--color-text-secondary);
}
</style>

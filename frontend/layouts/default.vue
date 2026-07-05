<template>
  <div class="min-h-screen bg-[var(--color-bg)]">
    <AppHeader />
    <AppSidebar :collapsed="sidebarCollapsed" @toggle="sidebarCollapsed = !sidebarCollapsed" />
    <main
      class="min-h-[calc(100vh-var(--header-height))] px-6 pb-8 transition-all duration-200"
      :style="mainStyle"
    >
      <slot />
    </main>
    <AppFooter :style="footerStyle" />
    <Toast />
  </div>
</template>

<script setup lang="ts">
import AppHeader from '~/components/common/AppHeader.vue'
import AppSidebar from '~/components/common/AppSidebar.vue'
import AppFooter from '~/components/common/AppFooter.vue'
import Toast from '~/components/common/Toast.vue'

const sidebarCollapsed = ref(false)

const mainStyle = computed(() => ({
  paddingTop: 'calc(var(--header-height) + 16px)',
  maxWidth: '1920px',
  marginLeft: sidebarCollapsed.value ? '72px' : '240px',
}))

const footerStyle = computed(() => ({
  marginLeft: sidebarCollapsed.value ? '72px' : '240px',
}))
</script>

<script setup>
import '../assets/css/main.css'

// Initialize theme system
const { isDark } = useTheme()
</script>

<template>
  <div class="app-root">
    <NuxtRouteAnnouncer />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>

<style scoped>
.app-root {
  width: 100%;
  height: 100vh;
  overflow: hidden;
}
</style>

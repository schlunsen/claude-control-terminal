<script setup>
import '../assets/css/main.css'

// Initialize dark mode
const { isDark } = useDarkMode()
</script>

<template>
  <div>
    <NuxtRouteAnnouncer />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>

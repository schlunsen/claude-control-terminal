<script setup>
import '../assets/css/main.css'

// Initialize theme system
const { isDark } = useTheme()
</script>

<template>
  <div>
    <NuxtRouteAnnouncer />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>

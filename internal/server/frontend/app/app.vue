<template>
  <div id="app">
    <NuxtRouteAnnouncer />
    <NuxtPage />
  </div>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Inter', 'Segoe UI', sans-serif;
  background: #f5f5f5;
  color: #1a1a1a;
  min-height: 100vh;
  line-height: 1.6;
}

#app {
  min-height: 100vh;
}
</style>

<script setup>
import { ref, computed } from 'vue'
import Home from './routes/Home.vue'
import About from './routes/About.vue'
import NotFound from './NotFound.vue'

const routes = {
  '/': Home,
  '/about': About,
}

const currentPath = ref(window.location.hash)

window.addEventListener('hashchange', () => {
  currentPath.value = window.location.hash
})

const currentView = computed(() => {
  return routes[currentPath.value.slice(1) || '/'] || NotFound
})
</script>

<template>
  <a href="/">Home</a> | <a href="/about">About</a> |
  <a href="/non-existent-path">Broken Link</a>
  <suspense>
    <component :is="currentView" />
  </suspense>
</template>

<script setup>
import { ref } from 'vue'

let data = ref([])
async function useFetch() {
  data = await fetch('https://jsonplaceholder.typicode.com/posts/1')
    .then((response) => response.json())
    .then((json) => json)
}
await useFetch()
</script>

<template>
  <h1>About</h1>
  <div class="size-full">{{ JSON.stringify(data) }}</div>
</template>

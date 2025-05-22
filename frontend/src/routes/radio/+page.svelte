<script>
	import { onMount } from "svelte";

    let message = $state("")

    onMount(async () => {
        message = await fetch("/api/ping").then(res => res.text())
    })
</script>
<h1 class="text-3xl">Hello</h1>

{message}
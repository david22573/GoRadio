<script>
	import '../app.css';
	import Nav from '$lib/components/nav/Nav.svelte';
	let { children } = $props();
</script>

<div class="flex flex-col min-h-screen size-full p-4">
	<Nav />

	{@render children?.()}
</div>

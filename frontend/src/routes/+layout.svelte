<script>
	import '../app.css';
	import Nav from '$lib/components/nav/Nav.svelte';
	let { children } = $props();
</script>

<Nav />

{@render children?.()}

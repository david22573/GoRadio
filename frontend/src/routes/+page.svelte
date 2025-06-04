<script lang="ts">
	let { data } = $props();
</script>

{#each data.stations as station}
	<div class="flex items-center gap-4">
		<a href="/radio/{station.id}">{JSON.stringify(station, null, 2)}</a>
		<audio src={station.url} controls> </audio>
	</div>
{/each}

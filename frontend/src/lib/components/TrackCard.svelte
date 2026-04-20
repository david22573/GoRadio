<script lang="ts">
	import { player } from '$lib/player.svelte';

	let { track } = $props<{
		track: any;
	}>();

	let isPlayingThis = $derived(
		(player.currentTrack?.id === track.id && track.id !== 0 || player.currentTrack?.url === track.url) && 
		player.isPlaying
	);
	let isLoadingThis = $derived(
		(player.currentTrack?.id === track.id && track.id !== 0 || player.currentTrack?.url === track.url) && 
		player.isLoading
	);

	function handlePlay() {
		player.startContinuous(track);
	}
</script>

<div
	class="bg-surface-container p-4 rounded-xl border border-outline-variant/10 flex items-center gap-4 group hover:border-primary/30 transition-colors"
>
	<div class="h-12 w-12 rounded-lg bg-surface-container-highest flex items-center justify-center text-primary shrink-0">
		<span class="material-symbols-outlined">music_note</span>
	</div>
	
	<div class="overflow-hidden flex-1">
		<h3 class="font-bold text-white truncate" title={track.title}>{track.title}</h3>
		<p class="text-xs text-on-surface-variant truncate">{track.artist}</p>
	</div>

	<button
		onclick={handlePlay}
		class="h-10 w-10 rounded-full bg-surface-container-highest hover:bg-primary text-white hover:text-on-primary-fixed flex items-center justify-center transition-all shrink-0"
	>
		{#if isPlayingThis}
			<span class="material-symbols-outlined text-[20px]" style="font-variation-settings: 'FILL' 1;">pause</span>
		{:else if isLoadingThis}
			<span class="material-symbols-outlined text-[20px] animate-spin">progress_activity</span>
		{:else}
			<span class="material-symbols-outlined text-[20px]" style="font-variation-settings: 'FILL' 1;">play_arrow</span>
		{/if}
	</button>
</div>

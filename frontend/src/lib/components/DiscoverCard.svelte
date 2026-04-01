<script lang="ts">
	import { player } from '$lib/player.svelte';

	let { result, onAdd } = $props<{
		result: any;
		onAdd: (result: any) => void;
	}>();

	// Computed state for the preview player
	let previewId = $derived(`preview-${result.stationuuid}`);
	let isPlayingThis = $derived(player.currentStation?.id === previewId && player.isPlaying);
	let isLoadingThis = $derived(player.currentStation?.id === previewId && player.isLoading);

	function handlePreview(e: Event) {
		e.preventDefault();
		player.play({
			id: previewId,
			name: result.name,
			url: result.url_resolved || result.url
		});
	}
</script>

<div
	class="bg-surface-container p-5 rounded-2xl border border-outline-variant/10 flex flex-col gap-4 group hover:border-primary/30 transition-colors"
>
	<div class="flex justify-between items-start gap-4">
		<div class="overflow-hidden">
			<h3 class="font-bold text-white truncate" title={result.name}>{result.name}</h3>
			<div class="flex items-center gap-2 mt-1">
				<span class="material-symbols-outlined text-[14px] text-primary">label</span>
				<p class="text-xs text-on-surface-variant uppercase tracking-widest font-label truncate">
					{result.tags || 'Generic'}
				</p>
			</div>
		</div>
		<button
			onclick={() => onAdd(result)}
			class="shrink-0 h-10 w-10 rounded-full bg-surface-container-highest hover:bg-primary text-white hover:text-on-primary-fixed flex items-center justify-center transition-all"
			title="Add to Library"
		>
			<span class="material-symbols-outlined text-xl">add</span>
		</button>
	</div>

	<button
		onclick={handlePreview}
		class="w-full py-2.5 rounded-lg bg-surface-container-highest hover:bg-primary hover:text-on-primary-fixed text-on-surface-variant transition-colors flex items-center justify-center gap-2 text-sm font-bold mt-auto"
	>
		{#if isPlayingThis}
			<span class="material-symbols-outlined text-[18px]" style="font-variation-settings: 'FILL' 1;"
				>pause</span
			>
			Playing
		{:else if isLoadingThis}
			<span class="material-symbols-outlined text-[18px] animate-spin">progress_activity</span>
			Connecting
		{:else}
			<span class="material-symbols-outlined text-[18px]" style="font-variation-settings: 'FILL' 1;"
				>play_arrow</span
			>
			Preview Stream
		{/if}
	</button>
</div>

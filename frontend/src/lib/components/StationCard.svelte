<script lang="ts">
	import { player } from '$lib/player.svelte';

	let { station, onDelete } = $props<{
		station: any;
		onDelete: (id: number) => void;
	}>();
</script>

<div
	class="group relative rounded-2xl bg-surface-container-low border border-outline-variant/10 p-6 hover:bg-surface-container transition-all duration-300 hover:shadow-[0_8px_30px_rgb(0,0,0,0.4)]"
>
	<a
		href={`/station/${station.id}`}
		class="absolute inset-0 z-10 rounded-2xl"
		aria-label={`View station details for ${station.name}`}
	></a>

	<div class="flex items-start justify-between mb-6 relative z-20">
		<button
			onclick={(e) => {
				e.preventDefault();
				e.stopPropagation();
				player.play(station);
			}}
			class="h-14 w-14 rounded-xl bg-surface-container-highest flex items-center justify-center text-primary hover:bg-gradient-to-br hover:from-primary hover:to-primary-container hover:text-on-primary-fixed transition-all duration-300 shadow-inner shadow-white/5"
			title="Play Station"
		>
			{#if player.currentStation?.id === station.id && player.isPlaying}
				<span class="material-symbols-outlined text-3xl" style="font-variation-settings: 'FILL' 1;"
					>volume_up</span
				>
			{:else if player.currentStation?.id === station.id && player.isLoading}
				<span class="material-symbols-outlined text-3xl animate-spin">progress_activity</span>
			{:else}
				<span class="material-symbols-outlined text-3xl" style="font-variation-settings: 'FILL' 1;"
					>play_arrow</span
				>
			{/if}
		</button>

		<button
			onclick={(e) => {
				e.preventDefault();
				e.stopPropagation();
				onDelete(station.id);
			}}
			class="h-10 w-10 rounded-full flex items-center justify-center text-on-surface-variant hover:text-error hover:bg-error/10 transition-colors"
			title="Delete Station"
		>
			<span class="material-symbols-outlined">delete</span>
		</button>
	</div>

	<div>
		<h3
			class="font-headline font-bold text-xl text-white mb-1.5 truncate group-hover:text-primary transition-colors"
		>
			{station.name}
		</h3>
		<div class="flex items-center gap-2">
			<span class="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"
			></span>
			<p class="text-xs text-on-surface-variant font-label uppercase tracking-widest truncate">
				{station.url}
			</p>
		</div>
	</div>
</div>

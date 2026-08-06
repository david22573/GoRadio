<script lang="ts">
	import { queueManager } from '$lib/queue.svelte';
</script>

<div class="p-4 md:p-5 bg-surface-container rounded-2xl border border-outline-variant/10 backdrop-blur-sm">
	<h3 class="text-xs font-label text-on-surface-variant uppercase tracking-widest mb-4">Up Next</h3>

	{#if queueManager.upcoming.length === 0}
		<p class="text-on-surface-variant/70 text-sm italic">Queue is empty...</p>
	{:else}
		<div class="space-y-3">
			{#each queueManager.upcoming as track, i}
				<div
					class="group flex items-center gap-3 p-2 rounded-xl hover:bg-surface-container-highest transition-colors"
				>
					<span class="text-xs font-mono text-on-surface-variant/70 w-4">{i + 1}</span>
					<div class="flex-1 min-w-0">
						<div class="text-sm font-bold text-white truncate">{track.title}</div>
						<div class="text-xs text-on-surface-variant truncate">{track.artist}</div>
					</div>
					<!-- Mode indicator would go here if tracked per-track in upcoming -->
				</div>
			{/each}
		</div>
	{/if}
</div>

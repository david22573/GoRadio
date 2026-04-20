<script lang="ts">
	import { queueManager } from '$lib/queue.svelte';
</script>

<div class="p-4 bg-zinc-900/50 rounded-xl border border-zinc-800 backdrop-blur-sm">
	<h3 class="text-sm font-semibold text-zinc-400 uppercase tracking-wider mb-4">Up Next</h3>
	
	{#if queueManager.upcoming.length === 0}
		<p class="text-zinc-500 text-sm italic">Queue is empty...</p>
	{:else}
		<div class="space-y-3">
			{#each queueManager.upcoming as track, i}
				<div class="group flex items-center gap-3 p-2 rounded-lg hover:bg-zinc-800/50 transition-colors">
					<span class="text-xs font-mono text-zinc-600 w-4">{i + 1}</span>
					<div class="flex-1 min-w-0">
						<div class="text-sm font-medium text-zinc-200 truncate">{track.title}</div>
						<div class="text-xs text-zinc-500 truncate">{track.artist}</div>
					</div>
					<!-- Mode indicator would go here if tracked per-track in upcoming -->
				</div>
			{/each}
		</div>
	{/if}
</div>

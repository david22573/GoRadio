<script lang="ts">
	import { sessionManager } from '$lib/session.svelte';
	import { fade } from 'svelte/transition';

	let isExploration = $derived(
		sessionManager.session?.exploration_rate && sessionManager.session.exploration_rate > 0.15
	);
</script>

{#if sessionManager.isActive}
	<div
		in:fade
		out:fade
		class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider transition-all duration-500 {isExploration
			? 'bg-purple-500/20 text-purple-400 border border-purple-500/30'
			: 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'}"
	>
		{#if isExploration}
			<span class="relative flex h-1.5 w-1.5 mr-1.5">
				<span
					class="animate-ping absolute inline-flex h-full w-full rounded-full bg-purple-400 opacity-75"
				></span>
				<span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-purple-500"></span>
			</span>
			Exploration
		{:else}
			<span class="h-1.5 w-1.5 rounded-full bg-emerald-500 mr-1.5"></span>
			Similar
		{/if}
	</div>
{/if}

<script lang="ts">
	import { sessionManager } from '$lib/session.svelte';
	import { onMount } from 'svelte';

	let metrics = $state<any>(null);

	async function refreshMetrics() {
		const id = sessionManager.getSessionId();
		if (!id) return;

		const res = await fetch(`/api/sessions/${id}/metrics`);
		if (res.ok) {
			metrics = await res.json();
		}
	}

	onMount(() => {
		const interval = setInterval(refreshMetrics, 5000);
		refreshMetrics();
		return () => clearInterval(interval);
	});
</script>

{#if sessionManager.isActive && metrics}
	<div
		class="grid grid-cols-2 gap-4 p-4 bg-zinc-900/50 rounded-xl border border-zinc-800 backdrop-blur-sm mt-4"
	>
		<div class="flex flex-col gap-1">
			<span class="text-[10px] text-zinc-500 font-bold uppercase tracking-widest">Skip Rate</span>
			<div class="flex items-end gap-2">
				<span class="text-xl font-mono text-zinc-100">{Math.round(metrics.SkipRate * 100)}%</span>
				<div class="h-2 flex-1 bg-zinc-800 rounded-full overflow-hidden mb-1.5">
					<div
						class="h-full bg-red-500/50 transition-all duration-500"
						style="width: {metrics.SkipRate * 100}%"
					></div>
				</div>
			</div>
		</div>

		<div class="flex flex-col gap-1 text-right">
			<span class="text-[10px] text-zinc-500 font-bold uppercase tracking-widest"
				>Acoustic Drift</span
			>
			<span class="text-xl font-mono text-blue-400">{metrics.VectorDrift.toFixed(3)}</span>
		</div>
	</div>
{/if}

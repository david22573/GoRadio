<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { player } from '$lib/player.svelte';

	let { data } = $props();

	let stationName = $state(data.station?.name || '');
	let stationUrl = $state(data.station?.url || '');

	async function updateStation(e: Event) {
		e.preventDefault();
		try {
			const res = await fetch(`/api/stations/${data.station.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: stationName, url: stationUrl })
			});
			if (res.ok) await invalidateAll();
			else alert('Error updating station');
		} catch (err) {
			console.error(err);
		}
	}

	async function deleteStation() {
		if (!confirm('Delete this station?')) return;
		try {
			const res = await fetch(`/api/stations/${data.station.id}`, { method: 'DELETE' });
			if (res.ok) {
				window.location.href = '/';
			} else alert('Error deleting station');
		} catch (err) {
			console.error(err);
		}
	}
</script>

<main class="h-full">
	<section class="relative px-8 pt-12 pb-16 hero-gradient overflow-hidden">
		<div class="flex flex-col md:flex-row gap-10 items-end relative z-10">
			<div
				class="w-64 h-64 shrink-0 rounded-xl overflow-hidden shadow-[0_20px_40px_rgba(0,0,0,0.6)] relative group bg-surface-container-highest"
			>
				<img
					alt="Station Art"
					class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-110"
					src="https://lh3.googleusercontent.com/aida-public/AB6AXuBdl783Gx32qan-8ff4rvZ0ayMAJQkt1M3SfFIfA6nl5r1eclzBY3z9q_pB_bGgE4hp9EDr2-n6cTmKkq51dqobXXeCvjD03phybkoo2Wm0lkqm3s1kcMgtuBIcUbjr0rE2usYfb0c1vUQTHFupsALOPRIpVoI5nt1VArzId5oszyNs7ErcKiIF_Wh7_r1TCD8oCUJbbwvExfPNW3pzyjoI9j1kbxm_FBSPq3Z2OlJZOvkmHPnJEr_eLHQQnYdMnCHyloCJASLDxv0"
				/>
				<div
					class="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
				>
					<span class="material-symbols-outlined text-white text-6xl">graphic_eq</span>
				</div>
			</div>
			<div class="flex-1">
				<div class="flex items-center gap-3 mb-2">
					<span
						class="px-3 py-1 rounded-full bg-primary/20 text-primary text-[10px] font-label font-bold uppercase tracking-widest"
						>Live Broadcast</span
					>
					<span class="text-on-surface-variant font-label text-sm uppercase tracking-tighter"
						>Stereo Audio</span
					>
				</div>
				<h2
					class="font-headline text-5xl md:text-7xl font-extrabold tracking-tighter mb-6 text-white"
				>
					{stationName}
				</h2>
				<div class="flex items-center gap-4">
					<button
						onclick={() => player.play(data.station)}
						class="flex items-center gap-3 px-8 py-4 rounded-full bg-gradient-to-br from-primary to-primary-container text-on-primary-fixed font-headline font-bold text-lg hover:scale-105 active:scale-95 transition-all shadow-lg shadow-primary/20"
					>
						{#if player.currentStation?.id === data.station.id && player.isPlaying}
							<span class="material-symbols-outlined" style="font-variation-settings: 'FILL' 1;"
								>volume_up</span
							>
							Playing
						{:else}
							<span class="material-symbols-outlined" style="font-variation-settings: 'FILL' 1;"
								>play_arrow</span
							>
							Play Now
						{/if}
					</button>

					<button
						class="h-14 w-14 rounded-full flex items-center justify-center border border-outline-variant/30 text-on-surface hover:bg-surface-container-highest transition-colors"
					>
						<span class="material-symbols-outlined">favorite</span>
					</button>
					<button
						class="h-14 w-14 rounded-full flex items-center justify-center border border-outline-variant/30 text-on-surface hover:bg-surface-container-highest transition-colors"
					>
						<span class="material-symbols-outlined">share</span>
					</button>
				</div>
			</div>
		</div>
		<div class="absolute -top-32 -left-32 w-96 h-96 bg-primary/10 rounded-full blur-[120px]"></div>
	</section>

	<div class="grid grid-cols-1 lg:grid-cols-3 gap-8 px-8 pb-12">
		<section class="lg:col-span-2 space-y-6">
			<div class="flex items-center justify-between">
				<h3
					class="font-headline text-2xl font-bold tracking-tight text-white flex items-center gap-2"
				>
					<span class="material-symbols-outlined text-primary">history_edu</span>
					Recent Recordings
				</h3>
				<button
					class="text-primary text-sm font-label font-bold uppercase tracking-wider hover:underline"
					>View All</button
				>
			</div>

			<div class="space-y-3">
				<div
					class="flex items-center justify-between p-4 rounded-lg bg-surface-container-low hover:bg-surface-container transition-colors group"
				>
					<div class="flex items-center gap-4">
						<div
							class="h-12 w-12 rounded-lg bg-surface-container-highest flex items-center justify-center text-primary group-hover:bg-primary group-hover:text-on-primary-fixed transition-colors"
						>
							<span class="material-symbols-outlined">mic</span>
						</div>
						<div>
							<h4 class="text-sm font-bold text-white">Midnight Synth Sessions Vol. 4</h4>
							<p
								class="text-xs text-on-surface-variant font-label uppercase tracking-widest mt-0.5"
							>
								REC_20231024_AAC • 12.4 MB
							</p>
						</div>
					</div>
					<div class="flex items-center gap-6">
						<span class="text-xs text-on-surface-variant font-label tracking-widest">48:12</span>
						<button class="text-primary hover:scale-110 transition-transform">
							<span class="material-symbols-outlined">play_circle</span>
						</button>
						<button class="text-on-surface-variant hover:text-white transition-colors">
							<span class="material-symbols-outlined">more_vert</span>
						</button>
					</div>
				</div>
			</div>
		</section>

		<section class="space-y-6">
			<div class="p-8 rounded-xl bg-surface-container-low border border-outline-variant/10">
				<h3 class="font-headline text-xl font-bold mb-6 text-white">Station Settings</h3>
				<form onsubmit={updateStation} class="space-y-6">
					<div>
						<label
							class="block text-xs font-label uppercase tracking-widest text-on-surface-variant mb-2"
							>Station Name</label
						>
						<input
							bind:value={stationName}
							class="w-full bg-surface-container-highest border-none rounded-lg p-3 text-sm focus:ring-1 focus:ring-primary text-on-surface"
							type="text"
						/>
					</div>
					<div>
						<label
							class="block text-xs font-label uppercase tracking-widest text-on-surface-variant mb-2"
							>Broadcast URL</label
						>
						<input
							bind:value={stationUrl}
							class="w-full bg-surface-container-highest border-none rounded-lg p-3 text-sm focus:ring-1 focus:ring-primary text-on-surface"
							type="url"
						/>
					</div>
					<div class="pt-4 space-y-3">
						<button
							type="submit"
							class="w-full py-3 rounded-lg bg-surface-container-highest text-white font-bold text-sm hover:bg-surface-variant transition-colors border border-outline-variant/20"
						>
							Update Information
						</button>
						<button
							type="button"
							onclick={deleteStation}
							class="w-full py-3 rounded-lg text-error text-sm font-bold hover:bg-error/10 transition-colors"
						>
							Delete Station
						</button>
					</div>
				</form>
			</div>

			<div class="p-6 rounded-xl bg-surface-container shadow-inner grid grid-cols-2 gap-4">
				<div class="bg-surface-container-lowest p-4 rounded-lg">
					<p class="text-[10px] font-label text-on-surface-variant uppercase tracking-widest mb-1">
						Peak Listeners
					</p>
					<p class="text-2xl font-headline font-extrabold text-primary">12.4K</p>
				</div>
				<div class="bg-surface-container-lowest p-4 rounded-lg">
					<p class="text-[10px] font-label text-on-surface-variant uppercase tracking-widest mb-1">
						Uptime
					</p>
					<p class="text-2xl font-headline font-extrabold text-secondary">99.8%</p>
				</div>
			</div>
		</section>
	</div>
</main>

<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import StationCard from '$lib/components/StationCard.svelte';
	import DiscoverCard from '$lib/components/DiscoverCard.svelte';
	import TrackCard from '$lib/components/TrackCard.svelte';

	let { data } = $props();

	// Station Search State
	let searchQuery = $state('');
	let searchResults = $state<any[]>([]);
	let isSearching = $state(false);

	// Track Search State
	let trackSearchQuery = $state('');
	let trackResults = $state<any[]>([]);
	let isSearchingTracks = $state(false);

	// Custom Add State
	let customName = $state('');
	let customUrl = $state('');
	let isSubmitting = $state(false);

	async function searchStations(e: Event) {
		e.preventDefault();
		if (!searchQuery) return;

		isSearching = true;
		try {
			const res = await fetch(`/api/search?q=${encodeURIComponent(searchQuery)}`);
			if (!res.ok) throw new Error('Search failed');
			searchResults = await res.json();
		} catch (err) {
			console.error(err);
			alert('Failed to fetch stations');
		} finally {
			isSearching = false;
		}
	}

	async function searchTracks(e: Event) {
		e.preventDefault();
		if (!trackSearchQuery) return;

		isSearchingTracks = true;
		try {
			const res = await fetch(`/api/tracks/search?q=${encodeURIComponent(trackSearchQuery)}`);
			if (!res.ok) throw new Error('Track search failed');
			const data = await res.json();
			trackResults = data.tracks || [];
		} catch (err) {
			console.error(err);
			alert('Failed to fetch tracks');
		} finally {
			isSearchingTracks = false;
		}
	}

	async function addFromSearch(result: any) {
		try {
			const res = await fetch('/api/stations', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: result.name, url: result.url_resolved || result.url })
			});
			if (res.ok) {
				await invalidateAll();
			} else {
				const err = await res.json();
				alert('Error adding station: ' + err.error);
			}
		} catch (err) {
			console.error(err);
		}
	}

	async function addCustomStation(e: Event) {
		e.preventDefault();
		if (!customName || !customUrl) return;

		isSubmitting = true;
		try {
			const res = await fetch('/api/stations', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: customName, url: customUrl })
			});
			if (res.ok) {
				customName = '';
				customUrl = '';
				await invalidateAll();
			} else {
				const err = await res.json();
				alert('Error adding station: ' + err.error);
			}
		} catch (err) {
			console.error(err);
		} finally {
			isSubmitting = false;
		}
	}

	async function deleteStation(id: number) {
		if (!confirm('Delete this station?')) return;

		try {
			const res = await fetch(`/api/stations/${id}`, { method: 'DELETE' });
			if (res.ok) {
				await invalidateAll();
			} else {
				const err = await res.json();
				alert('Error deleting station: ' + err.error);
			}
		} catch (err) {
			console.error(err);
		}
	}
</script>

<div class="max-w-7xl mx-auto space-y-16 mt-8 px-8 pb-12">
	<section>
		<div class="flex items-center justify-between mb-8">
			<h2 class="font-headline text-3xl md:text-4xl font-extrabold tracking-tighter text-white">
				Your Frequencies
			</h2>
			<span
				class="px-4 py-1.5 rounded-full bg-surface-container-highest text-primary font-label text-xs font-bold uppercase tracking-widest border border-outline-variant/10"
			>
				{data.stations?.length || 0} Saved
			</span>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
			{#if data.stations && data.stations.length > 0}
				{#each data.stations as station (station.id)}
					<StationCard {station} onDelete={deleteStation} />
				{/each}
			{:else}
				<div
					class="col-span-full flex flex-col items-center justify-center py-20 text-center bg-surface-container-low rounded-3xl border border-outline-variant/10 border-dashed"
				>
					<span class="material-symbols-outlined text-6xl text-on-surface-variant mb-4 opacity-50"
						>sensors_off</span
					>
					<h3 class="text-xl font-bold text-white mb-2">No stations tuned in</h3>
					<p class="text-on-surface-variant max-w-sm">
						Use the search below to discover global broadcasts or manually add a direct stream URL.
					</p>
				</div>
			{/if}
		</div>
	</section>

	<section
		class="p-8 rounded-3xl bg-surface-container-low border border-outline-variant/10 shadow-2xl shadow-black/50 relative overflow-hidden"
	>
		<div
			class="absolute -top-32 -right-32 w-96 h-96 bg-primary/5 rounded-full blur-[100px] pointer-events-none"
		></div>

		<div class="relative z-10">
			<h2
				class="font-headline text-3xl font-extrabold tracking-tighter mb-6 text-white flex items-center gap-3"
			>
				<span class="material-symbols-outlined text-primary">explore</span>
				Discover
			</h2>

			<form onsubmit={searchStations} class="flex flex-col md:flex-row gap-4 mb-8">
				<div class="relative flex-1">
					<span
						class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant"
						>search</span
					>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search by genre, callsign, or city..."
						class="w-full bg-surface-container-highest border border-outline-variant/20 rounded-xl py-4 pl-12 pr-4 text-sm focus:ring-2 focus:ring-primary text-white focus:outline-none transition-all"
						required
					/>
				</div>
				<button
					type="submit"
					disabled={isSearching}
					class="bg-white hover:bg-gray-200 text-black font-headline font-bold py-4 px-8 rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
				>
					{isSearching ? 'Scanning...' : 'Scan Airways'}
					{#if !isSearching}
						<span class="material-symbols-outlined text-sm">radar</span>
					{/if}
				</button>
			</form>

			{#if searchResults.length > 0}
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
					{#each searchResults as result (result.stationuuid)}
						<DiscoverCard {result} onAdd={addFromSearch} />
					{/each}
				</div>
			{/if}
		</div>
	</section>

	<section
		class="p-8 rounded-3xl bg-surface-container-low border border-outline-variant/10 shadow-2xl shadow-black/50 relative overflow-hidden"
	>
		<div
			class="absolute -top-32 -left-32 w-96 h-96 bg-secondary/5 rounded-full blur-[100px] pointer-events-none"
		></div>

		<div class="relative z-10">
			<h2
				class="font-headline text-3xl font-extrabold tracking-tighter mb-6 text-white flex items-center gap-3"
			>
				<span class="material-symbols-outlined text-secondary">music_note</span>
				Sonic Library
			</h2>

			<form onsubmit={searchTracks} class="flex flex-col md:flex-row gap-4 mb-8">
				<div class="relative flex-1">
					<span
						class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant"
						>search</span
					>
					<input
						type="text"
						bind:value={trackSearchQuery}
						placeholder="Search songs by title or artist..."
						class="w-full bg-surface-container-highest border border-outline-variant/20 rounded-xl py-4 pl-12 pr-4 text-sm focus:ring-2 focus:ring-secondary text-white focus:outline-none transition-all"
						required
					/>
				</div>
				<button
					type="submit"
					disabled={isSearchingTracks}
					class="bg-white hover:bg-gray-200 text-black font-headline font-bold py-4 px-8 rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
				>
					{isSearchingTracks ? 'Searching...' : 'Search Songs'}
					{#if !isSearchingTracks}
						<span class="material-symbols-outlined text-sm">library_music</span>
					{/if}
				</button>
			</form>

			{#if trackResults.length > 0}
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
					{#each trackResults as track (track.url)}
						<TrackCard {track} />
					{/each}
				</div>
			{:else if trackSearchQuery && !isSearchingTracks}
				<p class="text-on-surface-variant text-center py-8">No matching songs found in your archive.</p>
			{/if}
		</div>
	</section>

	<section class="p-8 rounded-3xl bg-surface-container border border-outline-variant/10">
		<div class="max-w-3xl">
			<h3 class="font-headline text-xl font-bold mb-2 text-white flex items-center gap-2">
				<span class="material-symbols-outlined text-tertiary-fixed">link</span>
				Direct Stream Connection
			</h3>
			<p class="text-sm text-on-surface-variant mb-6">
				Have a private icecast or shoutcast link? Manually patch it into your library here.
			</p>

			<form onsubmit={addCustomStation} class="flex flex-col md:flex-row gap-4">
				<input
					type="text"
					bind:value={customName}
					placeholder="Station Identity"
					class="flex-1 bg-surface-container-highest border border-outline-variant/20 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary text-white focus:outline-none"
					required
				/>
				<input
					type="url"
					bind:value={customUrl}
					placeholder="wss:// or http://..."
					class="flex-[2] bg-surface-container-highest border border-outline-variant/20 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary text-white focus:outline-none"
					required
				/>
				<button
					type="submit"
					disabled={isSubmitting}
					class="bg-surface-container-highest hover:bg-surface-variant text-white border border-outline-variant/20 font-headline font-bold py-3 px-8 rounded-xl transition-colors disabled:opacity-50"
				>
					{isSubmitting ? 'Patching...' : 'Connect'}
				</button>
			</form>
		</div>
	</section>
</div>

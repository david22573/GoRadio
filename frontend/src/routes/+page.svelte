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
	let searchPage = $state(1);
	let searchHasMore = $state(false);
	let isLoadingMore = $state(false);
	let searchError = $state('');
	let loadMoreError = $state('');
	
	let searchAbortController: AbortController | null = null;
	let searchTimeout: ReturnType<typeof setTimeout> | null = null;

	// Track Search State
	let trackSearchQuery = $state('');
	let trackResults = $state<any[]>([]);
	let isSearchingTracks = $state(false);

	// Custom Add State
	let customName = $state('');
	let customUrl = $state('');
	let isSubmitting = $state(false);

	function handleSearchInput() {
		if (searchTimeout) clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			executeSearch(true);
		}, 400);
	}

	async function executeSearch(isNewQuery: boolean) {
		if (!searchQuery) {
			searchResults = [];
			searchPage = 1;
			searchHasMore = false;
			searchError = '';
			loadMoreError = '';
			return;
		}

		if (isNewQuery) {
			searchPage = 1;
			searchError = '';
			loadMoreError = '';
			if (searchAbortController) searchAbortController.abort();
			searchAbortController = new AbortController();
			isSearching = true;
		} else {
			if (searchAbortController) searchAbortController.abort();
			searchAbortController = new AbortController();
			isLoadingMore = true;
			loadMoreError = '';
		}

		try {
			const res = await fetch(`/api/search?q=${encodeURIComponent(searchQuery)}&page=${searchPage}`, {
				signal: searchAbortController.signal
			});
			if (!res.ok) throw new Error('Search failed');
			const data = await res.json();
			
			if (isNewQuery) {
				searchResults = data.results || [];
			} else {
				const existingIds = new Set(searchResults.map(r => r.stationuuid));
				const newResults = (data.results || []).filter((r: any) => !existingIds.has(r.stationuuid));
				searchResults = [...searchResults, ...newResults];
			}
			searchPage = data.page;
			searchHasMore = data.has_more;
		} catch (err: any) {
			if (err.name === 'AbortError') return;
			console.error(err);
			if (isNewQuery) {
				searchError = 'Failed to fetch stations. Please try again.';
				searchResults = [];
			} else {
				loadMoreError = 'Failed to load more stations.';
			}
		} finally {
			if (isNewQuery) isSearching = false;
			isLoadingMore = false;
		}
	}

	function searchStations(e: Event) {
		e.preventDefault();
		if (searchTimeout) clearTimeout(searchTimeout);
		executeSearch(true);
	}

	function loadMoreStations() {
		if (isLoadingMore || !searchHasMore) return;
		searchPage += 1;
		executeSearch(false);
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

<div class="max-w-7xl mx-auto space-y-8 md:space-y-12 mt-4 md:mt-8 px-4 md:px-8">
	<!-- Section: Saved Stations (Your Frequencies) -->
	<section>
		<div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6 md:mb-8">
			<div>
				<h2 class="font-headline text-3xl md:text-5xl font-extrabold tracking-tight text-white mb-2">
					Your Frequencies
				</h2>
				<p class="text-on-surface-variant font-label text-sm">Your personalized sonic landscape</p>
			</div>
			<div class="flex items-center gap-3">
				<span class="px-4 py-2 rounded-full bg-surface-container-highest text-primary font-label text-xs font-bold uppercase tracking-widest border border-outline-variant/10 shadow-sm">
					{data.stations?.length || 0} Saved
				</span>
			</div>
		</div>

		{#if data.stations && data.stations.length > 0}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 md:gap-6">
				{#each data.stations as station (station.id)}
					<StationCard {station} onDelete={deleteStation} />
				{/each}
			</div>
		{:else}
			<div class="flex flex-col items-center justify-center py-24 px-4 text-center bg-surface-container-low rounded-[2rem] border border-outline-variant/10 border-dashed relative overflow-hidden group">
				<div class="absolute inset-0 bg-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
				<span class="material-symbols-outlined text-7xl text-primary/40 mb-6 drop-shadow-[0_0_15px_rgba(94,180,255,0.3)]">sensors_off</span>
				<h3 class="font-headline text-2xl font-bold text-white mb-3">No Stations Tuned In</h3>
				<p class="text-on-surface-variant max-w-md mx-auto text-sm leading-relaxed mb-8">
					Your dial is currently static. Use the discover section below to scan global broadcasts or manually patch in a direct stream.
				</p>
				<button onclick={() => document.getElementById('discover')?.scrollIntoView({ behavior: 'smooth' })} class="px-8 py-3 rounded-full bg-primary text-on-primary-fixed font-bold font-label tracking-widest text-xs uppercase hover:scale-105 active:scale-95 transition-all shadow-[0_0_20px_rgba(94,180,255,0.3)] min-h-[44px]">
					Start Discovering
				</button>
			</div>
		{/if}
	</section>

	<div class="grid grid-cols-1 xl:grid-cols-2 gap-8 md:gap-12" id="discover">
		<!-- Section: Discover (Stations) -->
		<section class="p-6 md:p-10 rounded-[2.5rem] bg-surface-container-low border border-outline-variant/10 shadow-2xl shadow-black/40 relative overflow-hidden flex flex-col">
			<div class="absolute -top-40 -right-40 w-[30rem] h-[30rem] bg-primary/10 rounded-full blur-[100px] pointer-events-none"></div>

			<div class="relative z-10 flex-1 flex flex-col">
				<div class="flex items-center gap-4 mb-3">
					<div class="h-12 w-12 rounded-2xl bg-primary/20 text-primary flex items-center justify-center">
						<span class="material-symbols-outlined text-2xl">explore</span>
					</div>
					<h2 class="font-headline text-2xl md:text-3xl font-extrabold tracking-tight text-white">
						Global Radar
					</h2>
				</div>
				<p class="text-on-surface-variant text-sm mb-8">Scan thousands of live internet radio stations worldwide.</p>

				<form onsubmit={searchStations} class="flex flex-col sm:flex-row gap-3 mb-8">
					<div class="relative flex-1 group">
						<span class="material-symbols-outlined absolute left-5 top-1/2 -translate-y-1/2 text-on-surface-variant group-focus-within:text-primary transition-colors">search</span>
						<input
							type="text"
							bind:value={searchQuery}
							oninput={handleSearchInput}
							placeholder="Genre, callsign, or city..."
							class="w-full bg-surface-container-highest border border-outline-variant/20 rounded-2xl py-4 pl-14 pr-4 text-sm focus:ring-2 focus:ring-primary focus:border-transparent text-white focus:outline-none transition-all shadow-inner min-h-[56px]"
							required
						/>
					</div>
					<button
						type="submit"
						disabled={isSearching}
						class="min-h-[56px] bg-white hover:bg-gray-200 text-black font-headline font-bold px-8 rounded-2xl transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 hover:shadow-[0_0_20px_rgba(255,255,255,0.2)] active:scale-95 shrink-0"
					>
						{isSearching ? 'Scanning...' : 'Scan'}
						{#if !isSearching}
							<span class="material-symbols-outlined text-[18px]">radar</span>
						{/if}
					</button>
				</form>

				<div class="flex-1">
					{#if searchError}
						<div class="flex flex-col items-center justify-center py-12 text-center bg-error-container/20 rounded-2xl border border-error/20">
							<span class="material-symbols-outlined text-4xl text-error mb-2">error</span>
							<p class="text-error text-sm mb-4">{searchError}</p>
							<button onclick={() => executeSearch(true)} class="px-6 py-2 rounded-full bg-error text-error-container font-bold text-xs uppercase min-h-[44px]">Retry Search</button>
						</div>
					{:else if isSearching}
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-4 animate-pulse">
							{#each [1, 2, 3, 4] as _}
								<div class="h-32 bg-surface-container-highest rounded-2xl border border-outline-variant/5"></div>
							{/each}
						</div>
					{:else if searchResults.length > 0}
						<div class="flex flex-col h-full max-h-[600px]">
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-4 overflow-y-auto pr-2 custom-scrollbar flex-1 pb-4">
								{#each searchResults as result (result.stationuuid)}
									<DiscoverCard {result} onAdd={addFromSearch} />
								{/each}
								
								{#if isLoadingMore}
									<div class="col-span-1 sm:col-span-2 grid grid-cols-1 sm:grid-cols-2 gap-4 animate-pulse mt-4">
										{#each [1, 2] as _}
											<div class="h-32 bg-surface-container-highest rounded-2xl border border-outline-variant/5"></div>
										{/each}
									</div>
								{/if}
							</div>

							{#if loadMoreError}
								<div class="mt-4 p-4 rounded-2xl bg-error-container/20 border border-error/20 flex items-center justify-between">
									<span class="text-error text-sm">{loadMoreError}</span>
									<button onclick={() => executeSearch(false)} class="px-4 py-2 rounded-full bg-error/20 text-error font-bold text-xs hover:bg-error/30 min-h-[44px]">Retry</button>
								</div>
							{:else if searchHasMore}
								<div class="mt-4 flex flex-col sm:flex-row items-center justify-between gap-4 border-t border-outline-variant/10 pt-4">
									<span class="text-on-surface-variant text-xs font-label uppercase tracking-widest hidden sm:block" aria-live="polite">Page {searchPage}</span>
									<button 
										onclick={loadMoreStations} 
										disabled={isLoadingMore}
										class="w-full sm:w-auto px-6 py-3 rounded-full bg-surface-container-highest text-primary font-bold text-sm hover:bg-surface-variant transition-colors disabled:opacity-50 min-h-[44px]"
									>
										{isLoadingMore ? 'Loading...' : 'Load More Stations'}
									</button>
								</div>
							{:else}
								<div class="mt-6 text-center border-t border-outline-variant/10 pt-4">
									<span class="text-on-surface-variant text-sm" aria-live="polite">End of results</span>
								</div>
							{/if}
						</div>
					{:else if searchQuery && !isSearching}
						<div class="flex flex-col items-center justify-center py-12 text-center bg-surface-container-highest/30 rounded-2xl border border-outline-variant/5">
							<span class="material-symbols-outlined text-4xl text-on-surface-variant mb-2 opacity-50">search_off</span>
							<p class="text-on-surface-variant text-sm">No frequencies found.</p>
						</div>
					{/if}
				</div>
			</div>
		</section>

		<!-- Section: Sonic Library (Tracks) -->
		<section class="p-6 md:p-10 rounded-[2.5rem] bg-surface-container-low border border-outline-variant/10 shadow-2xl shadow-black/40 relative overflow-hidden flex flex-col">
			<div class="absolute -top-40 -left-40 w-[30rem] h-[30rem] bg-secondary/10 rounded-full blur-[100px] pointer-events-none"></div>

			<div class="relative z-10 flex-1 flex flex-col">
				<div class="flex items-center gap-4 mb-3">
					<div class="h-12 w-12 rounded-2xl bg-secondary/20 text-secondary flex items-center justify-center">
						<span class="material-symbols-outlined text-2xl">music_note</span>
					</div>
					<h2 class="font-headline text-2xl md:text-3xl font-extrabold tracking-tight text-white">
						Sonic Library
					</h2>
				</div>
				<p class="text-on-surface-variant text-sm mb-8">Search for specific tracks across your indexed audio archives.</p>

				<form onsubmit={searchTracks} class="flex flex-col sm:flex-row gap-3 mb-8">
					<div class="relative flex-1 group">
						<span class="material-symbols-outlined absolute left-5 top-1/2 -translate-y-1/2 text-on-surface-variant group-focus-within:text-secondary transition-colors">search</span>
						<input
							type="text"
							bind:value={trackSearchQuery}
							placeholder="Track title or artist..."
							class="w-full bg-surface-container-highest border border-outline-variant/20 rounded-2xl py-4 pl-14 pr-4 text-sm focus:ring-2 focus:ring-secondary focus:border-transparent text-white focus:outline-none transition-all shadow-inner min-h-[56px]"
							required
						/>
					</div>
					<button
						type="submit"
						disabled={isSearchingTracks}
						class="min-h-[56px] bg-white hover:bg-gray-200 text-black font-headline font-bold px-8 rounded-2xl transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 hover:shadow-[0_0_20px_rgba(255,255,255,0.2)] active:scale-95 shrink-0"
					>
						{isSearchingTracks ? 'Searching...' : 'Search'}
						{#if !isSearchingTracks}
							<span class="material-symbols-outlined text-[18px]">library_music</span>
						{/if}
					</button>
				</form>

				<div class="flex-1">
					{#if isSearchingTracks}
						<div class="grid grid-cols-1 gap-4 animate-pulse">
							{#each [1, 2, 3] as _}
								<div class="h-[72px] bg-surface-container-highest rounded-xl border border-outline-variant/5"></div>
							{/each}
						</div>
					{:else if trackResults.length > 0}
						<div class="grid grid-cols-1 gap-4 max-h-[600px] overflow-y-auto pr-2 custom-scrollbar">
							{#each trackResults as track (track.url)}
								<TrackCard {track} />
							{/each}
						</div>
					{:else if trackSearchQuery && !isSearchingTracks}
						<div class="flex flex-col items-center justify-center py-12 text-center bg-surface-container-highest/30 rounded-2xl border border-outline-variant/5">
							<span class="material-symbols-outlined text-4xl text-on-surface-variant mb-2 opacity-50">search_off</span>
							<p class="text-on-surface-variant text-sm">No matching songs found.</p>
						</div>
					{/if}
				</div>
			</div>
		</section>
	</div>

	<!-- Section: Direct Stream Connection -->
	<section class="p-6 md:p-8 rounded-[2rem] bg-surface-container border border-outline-variant/10 shadow-lg mb-12">
		<div class="flex flex-col lg:flex-row lg:items-center gap-8">
			<div class="flex-1">
				<h3 class="font-headline text-xl font-bold mb-2 text-white flex items-center gap-3">
					<div class="h-10 w-10 rounded-xl bg-tertiary-fixed/20 text-tertiary-fixed flex items-center justify-center shrink-0">
						<span class="material-symbols-outlined text-xl">link</span>
					</div>
					Direct Stream Connection
				</h3>
				<p class="text-sm text-on-surface-variant lg:ml-13 lg:mt-2">
					Have a private icecast or shoutcast link? Manually patch it into your library.
				</p>
			</div>

			<form onsubmit={addCustomStation} class="flex-[2] flex flex-col sm:flex-row gap-3">
				<input
					type="text"
					bind:value={customName}
					placeholder="Station Identity"
					class="w-full sm:w-1/3 bg-surface-container-highest border border-outline-variant/20 rounded-xl px-4 py-4 text-sm focus:ring-2 focus:ring-primary focus:border-transparent text-white focus:outline-none transition-all shadow-inner min-h-[56px]"
					required
				/>
				<input
					type="url"
					bind:value={customUrl}
					placeholder="wss:// or http://..."
					class="w-full sm:flex-1 bg-surface-container-highest border border-outline-variant/20 rounded-xl px-4 py-4 text-sm focus:ring-2 focus:ring-primary focus:border-transparent text-white focus:outline-none transition-all shadow-inner min-h-[56px]"
					required
				/>
				<button
					type="submit"
					disabled={isSubmitting}
					class="w-full sm:w-auto bg-surface-variant hover:bg-surface-container-highest text-white border border-outline-variant/20 font-headline font-bold py-4 px-8 rounded-xl transition-all disabled:opacity-50 hover:shadow-lg active:scale-95 shrink-0 min-h-[56px]"
				>
					{isSubmitting ? 'Patching...' : 'Connect'}
				</button>
			</form>
		</div>
	</section>
</div>

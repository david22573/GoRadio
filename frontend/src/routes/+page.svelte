<script lang="ts">
	import { invalidateAll } from '$app/navigation';

	let { data } = $props();

	// Search State
	let searchQuery = $state('');
	let searchResults = $state<any[]>([]);
	let isSearching = $state(false);

	// Custom Add State
	let customName = $state('');
	let customUrl = $state('');
	let isSubmitting = $state(false);

	async function searchStations(e: Event) {
		e.preventDefault();
		if (!searchQuery) return;

		isSearching = true;
		try {
			// Using one of the main community nodes, filtering out broken streams
			const res = await fetch(
				`https://de1.api.radio-browser.info/json/stations/search?name=${encodeURIComponent(searchQuery)}&limit=12&hidebroken=true`
			);
			searchResults = await res.json();
		} catch (err) {
			console.error(err);
			alert('Failed to fetch from Radio-Browser');
		} finally {
			isSearching = false;
		}
	}

	async function addFromSearch(result: any) {
		try {
			// url_resolved is better because it bypasses redirects
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

<div class="max-w-6xl mx-auto space-y-12 mt-8 px-4 pb-12">
	<section>
		<h2 class="text-3xl font-bold mb-6 text-sky-400">Your Stations</h2>
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#if data.stations && data.stations.length > 0}
				{#each data.stations as station}
					<div
						class="bg-zinc-800 p-5 rounded-xl shadow-lg border border-zinc-700 flex flex-col gap-4"
					>
						<div class="flex justify-between items-start">
							<h3 class="text-xl font-bold truncate pr-4 text-zinc-100">{station.name}</h3>
							<button
								onclick={() => deleteStation(station.id)}
								class="text-red-400 hover:text-red-300 text-sm font-medium transition-colors"
							>
								Delete
							</button>
						</div>
						<audio src={station.url} controls class="w-full h-10 rounded outline-none"></audio>
					</div>
				{/each}
			{:else}
				<div
					class="col-span-full text-center py-12 text-zinc-500 bg-zinc-800/50 rounded-xl border border-zinc-700 border-dashed"
				>
					No saved stations yet. Search below to add some.
				</div>
			{/if}
		</div>
	</section>

	<section class="bg-zinc-800 p-6 rounded-xl shadow-lg border border-zinc-700">
		<h2 class="text-2xl font-bold mb-4 text-sky-400">Discover Stations</h2>
		<form onsubmit={searchStations} class="flex flex-col md:flex-row gap-4 mb-6">
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search by name, genre, or location..."
				class="flex-1 bg-zinc-900 border border-zinc-700 rounded-lg px-4 py-2 focus:outline-none focus:border-sky-500 text-white"
				required
			/>
			<button
				type="submit"
				disabled={isSearching}
				class="bg-sky-600 hover:bg-sky-500 text-white font-semibold py-2 px-6 rounded-lg transition-colors disabled:opacity-50"
			>
				{isSearching ? 'Searching...' : 'Search'}
			</button>
		</form>

		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			{#each searchResults as result}
				<div class="bg-zinc-900 p-4 rounded-lg border border-zinc-700 flex flex-col gap-3">
					<div class="flex justify-between items-start gap-2">
						<div class="overflow-hidden">
							<h3 class="font-bold text-zinc-100 truncate" title={result.name}>{result.name}</h3>
							<p class="text-xs text-zinc-400 truncate">{result.tags || 'No tags'}</p>
						</div>
						<button
							onclick={() => addFromSearch(result)}
							class="shrink-0 bg-zinc-700 hover:bg-zinc-600 text-white text-sm font-medium py-1.5 px-3 rounded transition-colors"
						>
							Add +
						</button>
					</div>
					<audio
						src={result.url_resolved || result.url}
						controls
						preload="none"
						class="w-full h-8 rounded outline-none"
					></audio>
				</div>
			{/each}
		</div>
	</section>

	<section class="bg-zinc-800/50 p-6 rounded-xl shadow border border-zinc-700 border-dashed">
		<h3 class="text-lg font-bold mb-4 text-zinc-300">Add via Custom URL</h3>
		<form onsubmit={addCustomStation} class="flex flex-col md:flex-row gap-4">
			<input
				type="text"
				bind:value={customName}
				placeholder="Station Name"
				class="flex-1 bg-zinc-900 border border-zinc-700 rounded-lg px-4 py-2 focus:outline-none focus:border-sky-500 text-white"
				required
			/>
			<input
				type="url"
				bind:value={customUrl}
				placeholder="Stream URL (http://...)"
				class="flex-[2] bg-zinc-900 border border-zinc-700 rounded-lg px-4 py-2 focus:outline-none focus:border-sky-500 text-white"
				required
			/>
			<button
				type="submit"
				disabled={isSubmitting}
				class="bg-zinc-700 hover:bg-zinc-600 text-white font-medium py-2 px-6 rounded-lg transition-colors disabled:opacity-50"
			>
				{isSubmitting ? 'Adding...' : 'Add'}
			</button>
		</form>
	</section>
</div>

<script lang="ts">
	import '../app.css';
	import { player } from '$lib/player.svelte';
	let { children } = $props();

	// Volume slider input handler
	function handleVolume(e: Event) {
		const target = e.target as HTMLInputElement;
		player.setVolume(parseFloat(target.value));
	}
</script>

<aside
	class="fixed left-0 top-0 hidden md:flex flex-col h-screen w-64 border-r-0 bg-gradient-to-r from-[#0e0e0e] to-[#262626] py-8 px-6 z-50"
>
	<div class="mb-10">
		<h1 class="text-2xl font-black text-[#5eb4ff] tracking-tighter">GoRadio</h1>
		<p class="text-on-surface-variant text-xs mt-1 font-label">The Digital Conductor</p>
	</div>
	<nav class="flex-1 space-y-4">
		<a
			class="flex items-center gap-4 px-4 py-3 rounded-xl text-[#adaaaa] hover:text-[#5eb4ff] hover:bg-[#262626] transition-colors duration-300"
			href="/"
		>
			<span class="material-symbols-outlined">home</span>
			<span class="font-headline tracking-tight font-bold text-lg">Home</span>
		</a>
		<a
			class="flex items-center gap-4 px-4 py-3 rounded-xl text-[#adaaaa] hover:text-[#5eb4ff] hover:bg-[#262626] transition-colors duration-300"
			href="/"
		>
			<span class="material-symbols-outlined">explore</span>
			<span class="font-headline tracking-tight font-bold text-lg">Discover</span>
		</a>
		<a
			class="flex items-center gap-4 px-4 py-3 rounded-xl text-[#5eb4ff] font-bold border-r-4 border-[#5eb4ff] bg-[#262626]/30"
			href="/"
		>
			<span class="material-symbols-outlined">podcasts</span>
			<span class="font-headline tracking-tight font-bold text-lg">Your Shows</span>
		</a>
		<a
			class="flex items-center gap-4 px-4 py-3 rounded-xl text-[#adaaaa] hover:text-[#5eb4ff] hover:bg-[#262626] transition-colors duration-300"
			href="/"
		>
			<span class="material-symbols-outlined">favorite</span>
			<span class="font-headline tracking-tight font-bold text-lg">Liked Stations</span>
		</a>
	</nav>
	<div class="mt-auto">
		<button
			class="w-full py-4 rounded-xl bg-gradient-to-br from-primary to-primary-container text-on-primary-fixed font-headline font-extrabold uppercase tracking-widest text-sm transition-transform active:scale-95 duration-200"
		>
			Go Premium
		</button>
	</div>
</aside>

<header
	class="fixed top-0 right-0 w-full md:w-[calc(100%-16rem)] h-16 bg-[#0e0e0e]/70 backdrop-blur-xl flex justify-between items-center px-8 z-40"
>
	<div class="flex items-center gap-4 w-1/2">
		<div class="relative w-full max-w-md">
			<span
				class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant"
				>search</span
			>
			<input
				class="w-full bg-surface-container-highest border-none rounded-full py-2 pl-10 pr-4 text-sm focus:ring-1 focus:ring-primary text-white focus:outline-none"
				placeholder="Search frequencies..."
				type="text"
			/>
		</div>
	</div>
	<div class="flex items-center gap-6">
		<button
			class="text-on-surface-variant hover:text-white transition-opacity opacity-80 hover:opacity-100"
		>
			<span class="material-symbols-outlined">notifications</span>
		</button>
		<button
			class="text-on-surface-variant hover:text-white transition-opacity opacity-80 hover:opacity-100"
		>
			<span class="material-symbols-outlined">settings</span>
		</button>
		<div
			class="h-10 w-10 rounded-full bg-surface-container-highest overflow-hidden border border-outline-variant/20"
		>
			<img
				alt="User profile"
				class="w-full h-full object-cover"
				src="https://lh3.googleusercontent.com/aida-public/AB6AXuBFjaiFWW1M1PMR8v5YBY_oTZADhJVu-wqZfHMdcr-PznYUwfSy1MASJ3Cn8NWW31x8U6aDApDnZipiBV58Wsc-DPDV93DxwT6DZAEQJ17bF208jbkGprzFNbHvBWKpIw-CyrJZHadp-YELQKJPwdbRTHVhdfWNG5D8QK5AhkxEHrdcEQtdOHMwhckwrgV2sgv6BS8RtipPxhvYm--JDhPIVmk0_0h6lEyijCoZCNQOskLb0r12D4dtyXDkAA9_CHIjULSLS4B5aP0"
			/>
		</div>
	</div>
</header>

<div class="ml-0 md:ml-64 mt-16 h-[calc(100vh-64px)] overflow-y-auto pb-32 relative z-10">
	{@render children?.()}
</div>

<footer
	class="fixed bottom-0 left-0 w-full z-50 flex items-center justify-between px-12 py-4 bg-[#262626]/80 backdrop-blur-[24px] h-24 rounded-t-[3rem] border-t border-white/5 shadow-[0_-20px_40px_rgba(0,0,0,0.5)]"
>
	<div class="flex items-center gap-4 w-1/4">
		<div
			class="h-14 w-14 rounded-xl overflow-hidden shrink-0 bg-surface-container flex items-center justify-center border border-white/5"
		>
			{#if player.currentStation}
				<span class="material-symbols-outlined text-3xl text-primary">radio</span>
			{:else}
				<span class="material-symbols-outlined text-2xl text-on-surface-variant">music_off</span>
			{/if}
		</div>
		<div class="overflow-hidden">
			<h5 class="text-sm font-bold text-white truncate">
				{player.currentStation?.name || 'Ready to Broadcast'}
			</h5>
			{#if player.isPlaying}
				<div class="flex items-center gap-2 mt-0.5">
					<span class="w-2 h-2 rounded-full bg-red-500 animate-pulse"></span>
					<p class="text-xs text-primary font-medium">LIVE</p>
				</div>
			{:else if player.currentStation}
				<p class="text-xs text-on-surface-variant font-medium mt-0.5">Tuned Out</p>
			{/if}
		</div>
	</div>

	<div class="flex flex-col items-center gap-2 flex-1 max-w-lg">
		<div class="flex items-center gap-8">
			<button
				disabled={!player.currentStation}
				onclick={() => player.togglePlay()}
				class="h-14 w-14 rounded-full bg-gradient-to-br from-primary to-primary-container flex items-center justify-center text-on-primary-fixed hover:scale-110 active:scale-95 transition-all shadow-[0_0_20px_rgba(94,180,255,0.3)] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
			>
				{#if player.isLoading}
					<span class="material-symbols-outlined text-4xl animate-spin">progress_activity</span>
				{:else if player.isPlaying}
					<span
						class="material-symbols-outlined text-4xl"
						style="font-variation-settings: 'FILL' 1;">pause</span
					>
				{:else}
					<span
						class="material-symbols-outlined text-4xl"
						style="font-variation-settings: 'FILL' 1;">play_arrow</span
					>
				{/if}
			</button>
		</div>

		{#if player.currentStation}
			<div class="w-full flex justify-center mt-1">
				{#if player.isPlaying}
					<div class="flex gap-1 items-end h-4">
						<div class="w-1 bg-primary rounded-t animate-[bounce_1s_ease-in-out_infinite]"></div>
						<div
							class="w-1 bg-primary rounded-t animate-[bounce_1.2s_ease-in-out_infinite_0.1s]"
						></div>
						<div
							class="w-1 bg-primary rounded-t animate-[bounce_0.8s_ease-in-out_infinite_0.2s]"
						></div>
						<div
							class="w-1 bg-primary rounded-t animate-[bounce_1.5s_ease-in-out_infinite_0.3s]"
						></div>
						<div
							class="w-1 bg-primary rounded-t animate-[bounce_1.1s_ease-in-out_infinite_0.4s]"
						></div>
					</div>
				{:else}
					<span class="text-[10px] font-label text-on-surface-variant uppercase tracking-widest"
						>Stream Paused</span
					>
				{/if}
			</div>
		{/if}
	</div>

	<div class="flex items-center justify-end gap-6 w-1/4">
		<div class="flex items-center gap-3">
			<span class="material-symbols-outlined text-on-surface-variant text-xl">
				{player.volume === 0 ? 'volume_off' : player.volume < 0.5 ? 'volume_down' : 'volume_up'}
			</span>
			<input
				type="range"
				min="0"
				max="1"
				step="0.05"
				value={player.volume}
				oninput={handleVolume}
				class="w-24 accent-primary cursor-pointer"
			/>
		</div>
	</div>
</footer>

<script lang="ts">
	import '../app.css';
	import { player } from '$lib/player.svelte';
	import { sessionManager } from '$lib/session.svelte';
	import QueuePanel from '$lib/components/QueuePanel.svelte';
	import SessionMetrics from '$lib/components/SessionMetrics.svelte';
	import AcousticJourney from '$lib/components/AcousticJourney.svelte';
	let { children } = $props();

	// Volume slider input handler
	function handleVolume(e: Event) {
		const target = e.target as HTMLInputElement;
		player.setVolume(parseFloat(target.value));
	}
	
	let isPlayerExpanded = $state(false);
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
	class="fixed top-0 right-0 w-full md:w-[calc(100%-16rem)] h-16 pt-safe box-content bg-[#0e0e0e]/70 backdrop-blur-xl flex justify-between items-center px-4 md:px-8 z-40"
>
	<div class="flex items-center gap-4 w-[60%] md:w-1/2">
		<div class="relative w-full max-w-md">
			<span
				class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant"
				>search</span
			>
			<input
				aria-label="Search frequencies globally"
				class="w-full bg-surface-container-highest border-none rounded-full py-2 pl-10 pr-4 text-sm focus:ring-1 focus:ring-primary text-white focus:outline-none"
				placeholder="Search frequencies..."
				type="text"
			/>
		</div>
	</div>
	<div class="flex items-center gap-3 md:gap-6">
		<button
			class="h-11 w-11 flex items-center justify-center text-on-surface-variant hover:text-white transition-opacity opacity-80 hover:opacity-100 rounded-full hover:bg-white/5 active:bg-white/10"
			aria-label="Notifications"
		>
			<span class="material-symbols-outlined">notifications</span>
		</button>
		<button
			class="h-11 w-11 flex items-center justify-center text-on-surface-variant hover:text-white transition-opacity opacity-80 hover:opacity-100 rounded-full hover:bg-white/5 active:bg-white/10"
			aria-label="Settings"
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

<div class="ml-0 md:ml-64 mt-[calc(4rem+env(safe-area-inset-top))] h-[calc(100dvh-4rem-env(safe-area-inset-top))] overflow-y-auto pb-[calc(10rem+env(safe-area-inset-bottom))] md:pb-32 relative z-10 transition-all {sessionManager.isActive ? 'md:mr-80' : ''}">
	{@render children?.()}
</div>

{#if sessionManager.isActive}
	<aside
		class="fixed inset-0 top-16 md:inset-auto md:right-0 md:top-16 flex flex-col h-[calc(100vh-4rem)] md:w-80 bg-[#121212]/95 md:bg-[#121212]/80 backdrop-blur-2xl border-l border-white/5 p-4 z-30 overflow-y-auto pb-48 md:pb-32"
	>
		<div class="space-y-6">
			<QueuePanel />
			<AcousticJourney />
			<SessionMetrics />
		</div>
	</aside>
{/if}

<footer
	class="fixed {isPlayerExpanded ? 'inset-0 z-[60] flex-col !bg-[#0e0e0e] p-6' : 'bottom-[calc(4rem+env(safe-area-inset-bottom))] md:bottom-0 left-0 w-full z-40 flex items-center justify-between px-4 md:px-12 py-2 md:py-4 bg-[#262626]/95 md:bg-[#262626]/80 backdrop-blur-[24px] h-16 md:h-24 md:rounded-t-[3rem] border-t border-white/5 shadow-[0_-20px_40px_rgba(0,0,0,0.5)] transition-all duration-300'}"
	onclick={(e) => {
		// Only expand if clicking on the background, not buttons
		if (!isPlayerExpanded && (e.target as HTMLElement).tagName !== 'BUTTON' && (e.target as HTMLElement).tagName !== 'INPUT' && (e.target as HTMLElement).tagName !== 'SPAN') {
			isPlayerExpanded = true;
		}
	}}
>
	{#if isPlayerExpanded}
		<div class="w-full flex justify-between items-center mb-8 pt-safe">
			<button class="h-12 w-12 flex items-center justify-center rounded-full hover:bg-white/10" onclick={(e) => { e.stopPropagation(); isPlayerExpanded = false; }}>
				<span class="material-symbols-outlined text-3xl">expand_more</span>
			</button>
			<span class="font-label text-xs uppercase tracking-widest text-on-surface-variant font-bold">Now Playing</span>
			<button class="h-12 w-12 flex items-center justify-center rounded-full hover:bg-white/10">
				<span class="material-symbols-outlined text-xl">more_vert</span>
			</button>
		</div>
		<div class="w-full aspect-square bg-surface-container rounded-2xl mb-8 flex items-center justify-center border border-white/5 shadow-2xl">
			{#if player.currentStation || player.currentTrack}
				<span class="material-symbols-outlined text-9xl text-primary opacity-50">radio</span>
			{:else}
				<span class="material-symbols-outlined text-8xl text-on-surface-variant opacity-30">music_off</span>
			{/if}
		</div>
		<div class="w-full mb-8">
			<h2 class="text-2xl font-bold text-white mb-1">
				{#if player.currentTrack}
					{player.currentTrack.title}
				{:else}
					{player.currentStation?.name || 'Ready to Broadcast'}
				{/if}
			</h2>
			<p class="text-lg text-primary">
				{#if player.currentTrack}
					{player.currentTrack.artist}
				{:else if player.isPlaying}
					LIVE
				{:else if player.currentStation}
					Tuned Out
				{/if}
			</p>
		</div>
		<div class="w-full mb-10">
			<input
				type="range"
				min="0"
				max={player.duration || 100}
				value={player.currentTime}
				oninput={(e) => player.seek(parseFloat((e.target as HTMLInputElement).value))}
				class="w-full h-1 bg-white/20 rounded-full appearance-none cursor-pointer accent-primary"
			/>
			<div class="flex justify-between w-full mt-2 text-xs text-on-surface-variant font-label">
				<span>0:00</span>
				<span>-:-</span>
			</div>
		</div>
		<div class="w-full flex justify-center items-center gap-8 mb-auto">
			<button class="text-on-surface-variant hover:text-white h-12 w-12 flex items-center justify-center rounded-full">
				<span class="material-symbols-outlined text-2xl">shuffle</span>
			</button>
			<button class="text-white hover:text-primary h-14 w-14 flex items-center justify-center rounded-full">
				<span class="material-symbols-outlined text-4xl" style="font-variation-settings: 'FILL' 1;">skip_previous</span>
			</button>
			<button
				disabled={!player.currentStation && !player.currentTrack}
				onclick={(e) => { e.stopPropagation(); player.togglePlay(); }}
				class="h-20 w-20 rounded-full bg-gradient-to-br from-primary to-primary-container flex items-center justify-center text-on-primary-fixed hover:scale-105 active:scale-95 transition-all shadow-[0_0_30px_rgba(94,180,255,0.4)] disabled:opacity-50"
			>
				{#if player.isLoading}
					<span class="material-symbols-outlined text-4xl animate-spin">progress_activity</span>
				{:else if player.isPlaying}
					<span class="material-symbols-outlined text-5xl" style="font-variation-settings: 'FILL' 1;">pause</span>
				{:else}
					<span class="material-symbols-outlined text-5xl" style="font-variation-settings: 'FILL' 1;">play_arrow</span>
				{/if}
			</button>
			<button
				disabled={!!player.currentStation || !player.currentTrack}
				onclick={(e) => { e.stopPropagation(); player.skip(); }}
				class="text-white hover:text-primary h-14 w-14 flex items-center justify-center rounded-full disabled:opacity-50"
			>
				<span class="material-symbols-outlined text-4xl" style="font-variation-settings: 'FILL' 1;">skip_next</span>
			</button>
			<button class="text-on-surface-variant hover:text-white h-12 w-12 flex items-center justify-center rounded-full">
				<span class="material-symbols-outlined text-2xl">repeat</span>
			</button>
		</div>
	{:else}
	{#if player.currentTrack}
		<input
			type="range"
			min="0"
			max={player.duration || 100}
			value={player.currentTime}
			oninput={(e) => player.seek(parseFloat((e.target as HTMLInputElement).value))}
			class="absolute top-0 left-1/2 -translate-x-1/2 w-[calc(100%-6rem)] h-1 bg-white/20 rounded-full appearance-none cursor-pointer accent-primary -translate-y-1/2 z-50"
		/>
	{/if}
	<div class="flex items-center gap-3 md:gap-4 w-[60%] md:w-1/4">
		<div
			class="h-12 w-12 md:h-14 md:w-14 rounded-lg md:rounded-xl overflow-hidden shrink-0 bg-surface-container flex items-center justify-center border border-white/5"
		>
			{#if player.currentStation || player.currentTrack}
				<span class="material-symbols-outlined text-3xl text-primary">radio</span>
			{:else}
				<span class="material-symbols-outlined text-2xl text-on-surface-variant">music_off</span>
			{/if}
		</div>
		<div class="overflow-hidden min-w-0">
			<h5 class="text-xs md:text-sm font-bold text-white truncate">
				{#if player.currentTrack}
					{player.currentTrack.title}
				{:else}
					{player.currentStation?.name || 'Ready to Broadcast'}
				{/if}
			</h5>
			{#if player.currentTrack}
				<p class="text-xs text-on-surface-variant font-medium mt-0.5 truncate">{player.currentTrack.artist}</p>
			{:else if player.isPlaying}
				<div class="flex items-center gap-2 mt-0.5">
					<span class="w-2 h-2 rounded-full bg-red-500 animate-pulse"></span>
					<p class="text-xs text-primary font-medium">LIVE</p>
				</div>
			{:else if player.currentStation}
				<p class="text-xs text-on-surface-variant font-medium mt-0.5">Tuned Out</p>
			{/if}
		</div>
	</div>

	<div class="flex flex-col md:flex-row items-end md:items-center gap-2 flex-1 justify-end md:justify-center max-w-lg">
		<div class="flex items-center gap-4 md:gap-8">
			<button
				disabled={!player.currentStation && !player.currentTrack}
				onclick={() => player.togglePlay()}
				class="h-10 w-10 md:h-14 md:w-14 rounded-full bg-gradient-to-br from-primary to-primary-container flex items-center justify-center text-on-primary-fixed hover:scale-110 active:scale-95 transition-all shadow-[0_0_20px_rgba(94,180,255,0.3)] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
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
			<button
				disabled={!!player.currentStation || !player.currentTrack}
				onclick={() => player.skip()}
				class="text-on-surface-variant hover:text-white transition-opacity disabled:opacity-50 disabled:cursor-not-allowed hidden md:flex items-center justify-center"
			>
				<span class="material-symbols-outlined text-3xl" style="font-variation-settings: 'FILL' 1;">fast_forward</span>
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


	<div class="hidden md:flex items-center justify-end gap-6 w-1/4">
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
	{/if}
</footer>

<nav class="fixed bottom-0 left-0 w-full h-16 box-content pb-safe bg-[#0e0e0e]/90 backdrop-blur-xl border-t border-white/5 z-50 flex md:hidden items-center justify-around px-2 transition-transform duration-300 {isPlayerExpanded ? 'translate-y-full' : 'translate-y-0'}">
	<a href="/" class="flex flex-col items-center justify-center h-full min-w-[4.5rem] gap-1 text-[#5eb4ff] rounded-xl active:bg-white/5">
		<span class="material-symbols-outlined text-[24px]" style="font-variation-settings: 'FILL' 1;">home</span>
		<span class="text-[10px] font-bold tracking-wider">Home</span>
	</a>
	<a href="/" class="flex flex-col items-center justify-center h-full min-w-[4.5rem] gap-1 text-[#adaaaa] hover:text-[#5eb4ff] rounded-xl active:bg-white/5 transition-colors">
		<span class="material-symbols-outlined text-[24px]">explore</span>
		<span class="text-[10px] font-bold tracking-wider">Discover</span>
	</a>
	<button onclick={() => sessionManager.isActive = !sessionManager.isActive} class="flex flex-col items-center justify-center h-full min-w-[4.5rem] gap-1 rounded-xl active:bg-white/5 transition-colors {sessionManager.isActive ? 'text-[#5eb4ff]' : 'text-[#adaaaa] hover:text-[#5eb4ff]'}" aria-label="Toggle Queue and Session Metrics">
		<span class="material-symbols-outlined text-[24px]">queue_music</span>
		<span class="text-[10px] font-bold tracking-wider">Queue</span>
	</button>
</nav>

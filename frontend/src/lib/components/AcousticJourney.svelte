<script lang="ts">
	import { sessionManager } from '$lib/session.svelte';
	import { onMount } from 'svelte';

	let canvas: HTMLCanvasElement;
	let points = $state<any[]>([]);

	async function fetchJourney() {
		const id = sessionManager.getSessionId();
		if (!id) return;

		const res = await fetch(`/api/sessions/${id}/journey`);
		if (res.ok) {
			const data = await res.json();
			points = data.journey;
			draw();
		}
	}

	function draw() {
		if (!canvas) return;
		const ctx = canvas.getContext('2d');
		if (!ctx) return;

		ctx.clearRect(0, 0, canvas.width, canvas.height);
		const w = canvas.width;
		const h = canvas.height;

		// Draw path
		ctx.beginPath();
		ctx.strokeStyle = '#484847'; // outline-variant
		ctx.lineWidth = 1;
		points.forEach((p, i) => {
			const x = (p.x / 100) * w;
			const y = (p.y / 100) * h;
			if (i === 0) ctx.moveTo(x, y);
			else ctx.lineTo(x, y);
		});
		ctx.stroke();

		// Draw points
		points.forEach((p, i) => {
			const x = (p.x / 100) * w;
			const y = (p.y / 100) * h;
			ctx.beginPath();
			ctx.fillStyle = p.mode === 'exploration' ? '#a855f7' : '#10b981';
			if (i === 0) ctx.fillStyle = '#f59e0b'; // Start
			ctx.arc(x, y, i === points.length - 1 ? 6 : 4, 0, Math.PI * 2);
			ctx.fill();

			// Glow for current point
			if (i === points.length - 1) {
				ctx.shadowBlur = 10;
				ctx.shadowColor = ctx.fillStyle;
				ctx.stroke();
				ctx.shadowBlur = 0;
			}
		});
	}

	onMount(() => {
		const interval = setInterval(fetchJourney, 10000);
		fetchJourney();
		return () => clearInterval(interval);
	});
</script>

<div class="p-4 md:p-5 bg-surface-container rounded-2xl border border-outline-variant/10 backdrop-blur-sm mt-4">
	<h3 class="text-[10px] text-on-surface-variant font-label font-bold uppercase tracking-widest mb-4">
		Acoustic Journey
	</h3>
	<canvas
		bind:this={canvas}
		width="300"
		height="200"
		class="w-full h-auto rounded-xl bg-surface-container-lowest"
	></canvas>
	<div class="flex flex-wrap justify-between gap-2 mt-3">
		<div class="flex items-center gap-1.5 text-[10px] font-label text-on-surface-variant">
			<span class="w-2 h-2 rounded-full bg-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.5)]"></span> Start
		</div>
		<div class="flex items-center gap-1.5 text-[10px] font-label text-on-surface-variant">
			<span class="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></span> Similar
		</div>
		<div class="flex items-center gap-1.5 text-[10px] font-label text-on-surface-variant">
			<span class="w-2 h-2 rounded-full bg-purple-500 shadow-[0_0_8px_rgba(168,85,247,0.5)]"></span> Exploration
		</div>
	</div>
</div>

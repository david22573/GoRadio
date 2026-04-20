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
		ctx.strokeStyle = '#3f3f46';
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

<div class="p-4 bg-zinc-900/50 rounded-xl border border-zinc-800 backdrop-blur-sm mt-4">
	<h3 class="text-[10px] text-zinc-500 font-bold uppercase tracking-widest mb-4">
		Acoustic Journey
	</h3>
	<canvas
		bind:this={canvas}
		width="300"
		height="200"
		class="w-full h-auto rounded-lg bg-zinc-950/50"
	></canvas>
	<div class="flex gap-4 mt-3">
		<div class="flex items-center gap-1.5 text-[10px] text-zinc-500">
			<span class="w-2 h-2 rounded-full bg-amber-500"></span> Start
		</div>
		<div class="flex items-center gap-1.5 text-[10px] text-zinc-500">
			<span class="w-2 h-2 rounded-full bg-emerald-500"></span> Similar
		</div>
		<div class="flex items-center gap-1.5 text-[10px] text-zinc-500">
			<span class="w-2 h-2 rounded-full bg-purple-500"></span> Exploration
		</div>
	</div>
</div>

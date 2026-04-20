import { sessionManager } from './session.svelte';
import { queueManager } from './queue.svelte';

class AudioPlayer {
	currentStation = $state<any>(null);
	currentTrack = $state<any>(null);
	isPlaying = $state(false);
	isLoading = $state(false);
	volume = $state(0.75);

	private audio: HTMLAudioElement | null = null;
	private startTime: number = 0;

	init() {
		if (typeof window === 'undefined' || this.audio) return;

		this.audio = new Audio();
		this.audio.volume = this.volume;

		this.audio.addEventListener('playing', () => {
			this.isPlaying = true;
			this.isLoading = false;
			this.startTime = Date.now();
		});

		this.audio.addEventListener('pause', () => {
			this.isPlaying = false;
		});

		this.audio.addEventListener('waiting', () => {
			this.isLoading = true;
		});

		this.audio.addEventListener('ended', () => {
			this.handleTrackEnd();
		});

		this.audio.addEventListener('error', (e) => {
			console.error('Stream playback error:', e);
			this.isPlaying = false;
			this.isLoading = false;
		});
	}

	async startContinuous(seedTrack: any) {
		await sessionManager.createSession(seedTrack.id);
		await queueManager.fetchQueue();
		this.playTrack(seedTrack);
	}

	play(station: any) {
		this.init();
		if (!this.audio) return;
		this.currentTrack = null;

		if (this.currentStation?.id === station.id && this.isPlaying) return;

		this.currentStation = station;
		this.isLoading = true;

		this.audio.src = station.url_resolved || station.url;
		this.audio.load();
		this.audio.play().catch((err) => {
			console.error('Playback failed:', err);
			this.isLoading = false;
			this.isPlaying = false;
		});
	}

	playTrack(track: any) {
		this.init();
		if (!this.audio) return;
		this.currentStation = null;
		this.currentTrack = track;
		this.isLoading = true;

		this.audio.src = track.url;
		this.audio.load();
		this.audio.play().catch((err) => {
			console.error('Track playback failed:', err);
			this.isLoading = false;
			this.isPlaying = false;
		});
	}

	async skip() {
		if (!this.audio || !this.currentTrack) return;

		const completion = this.audio.currentTime / this.audio.duration;
		await this.recordEvent('skip', completion);

		await queueManager.advance();
		if (queueManager.currentTrack) {
			this.playTrack(queueManager.currentTrack);
		}
	}

	private async handleTrackEnd() {
		if (this.currentTrack) {
			await this.recordEvent('play', 1.0);
			await queueManager.advance();
			if (queueManager.currentTrack) {
				this.playTrack(queueManager.currentTrack);
			}
		}
	}

	private async recordEvent(type: 'play' | 'skip', completion: number) {
		const sessionId = sessionManager.getSessionId();
		if (!sessionId || !this.currentTrack) return;

		const body: any = {
			session_id: sessionId,
			track_id: this.currentTrack.id,
			completion: completion
		};

		if (type === 'play') {
			body.started_at = new time.Time(); // JS needs to format this correctly for Go
			// Simplified for now
		}

		await fetch(`/api/events/${type}`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		});
	}

	togglePlay() {
		if (!this.audio) return;
		if (!this.currentStation && !this.currentTrack) return;

		if (this.isPlaying) {
			this.audio.pause();
			if (this.currentStation) {
				this.audio.removeAttribute('src');
				this.audio.load();
			}
		} else {
			this.isLoading = true;
			const src = this.currentStation
				? this.currentStation.url_resolved || this.currentStation.url
				: this.currentTrack.url;
			this.audio.src = src;
			this.audio.load();
			this.audio.play().catch(console.error);
		}
	}

	setVolume(val: number) {
		this.volume = Math.max(0, Math.min(1, val));
		if (this.audio) {
			this.audio.volume = this.volume;
		}
	}
}

export const player = new AudioPlayer();

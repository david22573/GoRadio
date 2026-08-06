import { sessionManager } from './session.svelte';
import { queueManager } from './queue.svelte';

class AudioPlayer {
	currentStation = $state<any>(null);
	currentTrack = $state<any>(null);
	isPlaying = $state(false);
	isLoading = $state(false);
	volume = $state(0.75);
	currentTime = $state(0);
	duration = $state(0);

	private audio: HTMLAudioElement | null = null;
	private nextAudio: HTMLAudioElement | null = null;
	private startTime: number = 0;

	constructor() {
		if (typeof window !== 'undefined') {
			const saved = localStorage.getItem('goradio_last_station');
			if (saved) {
				try {
					const parsedStation = JSON.parse(saved);
					this.play(parsedStation);
				} catch (e) {
					console.error('Failed to parse saved station', e);
				}
			}
		}

		$effect.root(() => {
			$effect(() => {
				if (this.currentStation) {
					localStorage.setItem('goradio_last_station', JSON.stringify(this.currentStation));
				} else {
					localStorage.removeItem('goradio_last_station');
				}
			});

			$effect(() => {
				const next = queueManager.nextTrack;
				if (next && this.nextAudio) {
					fetch(`/api/tracks/resolve?url=${encodeURIComponent(next.url)}`)
						.then(res => res.json())
						.then(data => {
							if (this.nextAudio) {
								this.nextAudio.src = data.url || next.url;
								this.nextAudio.preload = "auto";
							}
						})
						.catch(err => console.error("Failed to preload next track", err));
				}
			});
		});
	}

	private setupListeners(audioElement: HTMLAudioElement) {
		audioElement.addEventListener('playing', () => {
			if (this.audio !== audioElement) return;
			this.isPlaying = true;
			this.isLoading = false;
			this.startTime = Date.now();
		});

		audioElement.addEventListener('pause', () => {
			if (this.audio !== audioElement) return;
			this.isPlaying = false;
		});

		audioElement.addEventListener('waiting', () => {
			if (this.audio !== audioElement) return;
			this.isLoading = true;
		});

		audioElement.addEventListener('ended', () => {
			if (this.audio !== audioElement) return;
			this.handleTrackEnd();
		});

		audioElement.addEventListener('timeupdate', () => {
			if (this.audio !== audioElement) return;
			this.currentTime = audioElement.currentTime || 0;
		});

		audioElement.addEventListener('loadedmetadata', () => {
			if (this.audio !== audioElement) return;
			this.duration = audioElement.duration || 0;
		});

		audioElement.addEventListener('error', (e) => {
			if (this.audio !== audioElement) return;
			console.error('Stream playback error:', e);
			this.isPlaying = false;
			this.isLoading = false;
			if (this.currentTrack) {
				this.skip();
			}
		});
	}

	init() {
		if (typeof window === 'undefined' || this.audio) return;

		this.audio = new Audio();
		this.audio.volume = this.volume;
		this.setupListeners(this.audio);

		this.nextAudio = new Audio();
		this.nextAudio.volume = this.volume;
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
			if (err.name === 'NotAllowedError') {
				console.warn('Autoplay blocked by browser. User interaction required.');
			} else {
				console.error('Playback failed:', err);
			}
			this.isLoading = false;
			this.isPlaying = false;
		});
	}

	async playTrack(track: any) {
		this.init();
		if (!this.audio) return;
		this.currentStation = null;
		this.currentTrack = track;
		this.isLoading = true;

		try {
			// Resolve URL if needed (YouTube etc)
			const res = await fetch(`/api/tracks/resolve?url=${encodeURIComponent(track.url)}`);
			const data = await res.json();
			const resolvedUrl = data.url || track.url;

			this.audio.src = resolvedUrl;
			this.audio.load();
			await this.audio.play();
		} catch (err) {
			console.error('Track playback failed:', err);
			this.isLoading = false;
			this.isPlaying = false;
		}
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

	seek(time: number) {
		if (this.audio) {
			this.audio.currentTime = time;
		}
	}

	private async handleTrackEnd() {
		if (this.currentTrack) {
			await this.recordEvent('play', 1.0);
			
			if (this.nextAudio && this.nextAudio.src) {
				const oldAudio = this.audio;
				
				// Swap and play immediately
				this.audio = this.nextAudio;
				this.setupListeners(this.audio);
				this.audio.volume = this.volume;
				this.audio.play().catch(console.error);
				
				// Recreate nextAudio
				this.nextAudio = new Audio();
				this.nextAudio.volume = this.volume;
				
				// Cleanup old audio
				if (oldAudio) {
					oldAudio.pause();
					oldAudio.removeAttribute('src');
				}
				
				await queueManager.advance();
				if (queueManager.currentTrack) {
					this.currentTrack = queueManager.currentTrack;
					this.currentStation = null;
				}
			} else {
				await queueManager.advance();
				if (queueManager.currentTrack) {
					this.playTrack(queueManager.currentTrack);
				}
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
			body.started_at = new Date().toISOString();
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

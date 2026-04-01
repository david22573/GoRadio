class AudioPlayer {
	currentStation = $state<any>(null);
	isPlaying = $state(false);
	isLoading = $state(false);
	volume = $state(0.75);

	private audio: HTMLAudioElement | null = null;

	init() {
		if (typeof window === 'undefined' || this.audio) return;

		this.audio = new Audio();
		this.audio.volume = this.volume;

		this.audio.addEventListener('playing', () => {
			this.isPlaying = true;
			this.isLoading = false;
		});

		this.audio.addEventListener('pause', () => {
			this.isPlaying = false;
		});

		this.audio.addEventListener('waiting', () => {
			this.isLoading = true;
		});

		this.audio.addEventListener('error', (e) => {
			console.error('Stream playback error:', e);
			this.isPlaying = false;
			this.isLoading = false;
		});
	}

	play(station: any) {
		this.init();
		if (!this.audio) return;

		// If it's the same station and already playing, do nothing
		if (this.currentStation?.id === station.id && this.isPlaying) return;

		this.currentStation = station;
		this.isLoading = true;

		// For live streams, setting src forces a fresh connection
		this.audio.src = station.url_resolved || station.url;
		this.audio.load();
		this.audio.play().catch((err) => {
			console.error('Playback failed:', err);
			this.isLoading = false;
			this.isPlaying = false;
		});
	}

	togglePlay() {
		if (!this.audio || !this.currentStation) return;

		if (this.isPlaying) {
			this.audio.pause();
			// OPTIMIZATION: Sever the connection to stop background buffering
			this.audio.removeAttribute('src');
			this.audio.load();
		} else {
			this.isLoading = true;
			// Reconnect the stream
			this.audio.src = this.currentStation.url_resolved || this.currentStation.url;
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

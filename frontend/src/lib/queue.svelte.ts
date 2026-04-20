import { sessionManager } from './session.svelte';

interface Track {
	id: number;
	title: string;
	artist: string;
	url: string;
	duration: number;
}

class QueueManager {
	currentTrack = $state<Track | null>(null);
	nextTrack = $state<Track | null>(null);
	upcoming = $state<Track[]>([]);
	isLoading = $state(false);

	async fetchQueue() {
		const sessionId = sessionManager.getSessionId();
		if (!sessionId) return;

		try {
			const res = await fetch(`/api/queue/${sessionId}`);
			if (res.ok) {
				const data = await res.json();
				this.currentTrack = data.current;
				this.nextTrack = data.next;
				this.upcoming = data.upcoming || [];
			}
		} catch (err) {
			console.error('Fetch queue error:', err);
		}
	}

	async advance() {
		const sessionId = sessionManager.getSessionId();
		if (!sessionId) return;

		this.isLoading = true;
		try {
			const res = await fetch(`/api/queue/${sessionId}/advance`, { method: 'POST' });
			if (res.ok) {
				const track = await res.json();
				this.currentTrack = track;
				// After advancing, the queue state on backend changes, so refresh
				await this.fetchQueue();
			}
		} catch (err) {
			console.error('Advance queue error:', err);
		} finally {
			this.isLoading = false;
		}
	}
}

export const queueManager = new QueueManager();

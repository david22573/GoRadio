interface Session {
	id: string;
	current_vector: number[];
	exploration_rate: number;
}

class SessionManager {
	session = $state<Session | null>(null);
	isActive = $state(false);

	constructor() {
		if (typeof window !== 'undefined') {
			this.resumeSession();
		}
	}

	async createSession(seedTrackId: number): Promise<void> {
		try {
			const res = await fetch('/api/sessions', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ seed_track_id: seedTrackId })
			});
			if (!res.ok) throw new Error('Failed to create session');
			
			const data = await res.json();
			this.session = data;
			this.isActive = true;
			this.saveToStorage();
		} catch (err) {
			console.error('Session creation error:', err);
		}
	}

	async endSession(): Promise<void> {
		if (this.session) {
			// Backend DELETE optional if we just want client-side expiry
			this.session = null;
			this.isActive = false;
			localStorage.removeItem('goradio_session');
		}
	}

	private saveToStorage() {
		if (this.session) {
			localStorage.setItem('goradio_session', JSON.stringify(this.session));
		}
	}

	private async resumeSession() {
		const stored = localStorage.getItem('goradio_session');
		if (stored) {
			try {
				const parsed = JSON.parse(stored);
				// Validate session still exists on backend
				const res = await fetch(`/api/sessions/${parsed.id}`);
				if (res.ok) {
					this.session = await res.json();
					this.isActive = true;
				} else {
					localStorage.removeItem('goradio_session');
				}
			} catch (e) {
				localStorage.removeItem('goradio_session');
			}
		}
	}

	getSessionId(): string | null {
		return this.session?.id || null;
	}
}

export const sessionManager = new SessionManager();

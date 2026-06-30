import { writable } from 'svelte/store';
import { browser } from '$app/environment';

const STORAGE_KEY = 'dm-companion-token';

function createAuthStore() {
	const initial = browser ? localStorage.getItem(STORAGE_KEY) : null;
	const { subscribe, set } = writable<string | null>(initial);

	return {
		subscribe,
		setToken(token: string) {
			if (browser) localStorage.setItem(STORAGE_KEY, token);
			set(token);
		},
		clear() {
			if (browser) localStorage.removeItem(STORAGE_KEY);
			set(null);
		}
	};
}

export const authToken = createAuthStore();

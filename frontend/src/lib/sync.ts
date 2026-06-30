import { db } from './db';
import { api, ApiError } from './api';

let flushing = false;

/** Sends every queued mutation to the server in order. Stops at the first
 * failure so retries don't reorder writes; a failed item stays queued. */
export async function flushMutationQueue(): Promise<void> {
	if (flushing || !navigator.onLine) return;
	flushing = true;
	try {
		const items = await db.mutationQueue.orderBy('createdAt').toArray();
		for (const item of items) {
			try {
				await api[item.method.toLowerCase() as 'post' | 'put' | 'patch' | 'delete'](
					item.path,
					item.body
				);
				await db.mutationQueue.delete(item.id!);
			} catch (err) {
				if (err instanceof ApiError && err.status >= 400 && err.status < 500) {
					// Won't succeed on retry (e.g. record deleted server-side); drop it
					// rather than blocking the queue forever.
					await db.mutationQueue.delete(item.id!);
					continue;
				}
				break;
			}
		}
	} finally {
		flushing = false;
	}
}

export function startSyncListeners() {
	window.addEventListener('online', () => void flushMutationQueue());
	void flushMutationQueue();
}

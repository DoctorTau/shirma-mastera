// The whole app is a personal, authenticated tool with no public pages worth
// prerendering, and several lib modules (Dexie/IndexedDB) only work in the
// browser. Build every route as part of the adapter-static SPA fallback.
export const prerender = false;

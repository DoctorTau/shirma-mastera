import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			adapter: adapter({
				fallback: 'index.html'
			})
		}),
		VitePWA({
			registerType: 'autoUpdate',
			manifest: {
				name: 'DM Companion',
				short_name: 'DM Companion',
				description: 'Боевая ширма и справочник монстров для мастера D&D',
				start_url: '/',
				display: 'standalone',
				orientation: 'any',
				background_color: '#111111',
				theme_color: '#111111',
				icons: [
					{ src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' },
					{ src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png' }
				]
			},
			workbox: {
				globPatterns: ['**/*.{js,css,html,ico,png,svg,webmanifest}'],
				navigateFallback: '/index.html',
				runtimeCaching: [
					{
						urlPattern: ({ url }: { url: URL }) => url.pathname.startsWith('/api/'),
						handler: 'NetworkFirst',
						options: {
							cacheName: 'api-cache',
							networkTimeoutSeconds: 5,
							cacheableResponse: { statuses: [0, 200] }
						}
					}
				]
			}
		})
	],
	server: {
		proxy: {
			'/api': 'http://localhost:8080'
		}
	}
});

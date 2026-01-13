import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			'/api': {
				target: 'http://backend:8090',
				changeOrigin: true
			},
			'/_': {
				target: 'http://backend:8090',
				changeOrigin: true
			}
		},
		host: true,
    	port: 5173,
	}
});

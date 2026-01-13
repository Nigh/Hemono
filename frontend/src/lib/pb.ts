import PocketBase from 'pocketbase';
import { writable } from 'svelte/store';

// 自动检测 API 地址
const getApiUrl = () => {
	// 优先使用环境变量
	if (import.meta.env.VITE_API_URL) {
		return import.meta.env.VITE_API_URL;
	}

	// 生产环境：使用当前 origin（通过 Caddy 反代）
	if (import.meta.env.PROD) {
		return window.location.origin;
	}

	// 开发环境：指向本地后端
	return 'http://127.0.0.1:8090';
};

export const pb = new PocketBase(getApiUrl());

export const currentUser = writable(pb.authStore.model);

pb.authStore.onChange((token, model) => {
	console.log('Auth state changed:', model);
	currentUser.set(model);
}, true);

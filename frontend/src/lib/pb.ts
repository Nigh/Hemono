import PocketBase, { ClientResponseError } from 'pocketbase';
import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// 自动检测 API 地址
const getApiUrl = () => {
	if (import.meta.env.VITE_API_URL) {
		return import.meta.env.VITE_API_URL;
	}

	if (import.meta.env.PROD) {
		// 只有在浏览器端才访问 window
		if (browser) {
			return window.location.origin;
		}
		// 服务器端渲染时，返回一个占位符或内网地址
		return '';
	}

	return 'http://localhost:8090';
};
export const pb = new PocketBase(getApiUrl());

pb.beforeSend = (url, options) => {
	if (pb.authStore.token && !pb.authStore.isValid && !url.includes('/auth-')) {
		pb.authStore.clear();
		throw new ClientResponseError({ status: 401, response: { message: '登录已过期，请重新登录' } });
	}
	return { url, options };
};

export const currentUser = writable(pb.authStore.model);

pb.authStore.onChange((_token, model) => {
	currentUser.set(model);
}, true);

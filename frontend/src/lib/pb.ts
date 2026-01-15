import PocketBase from 'pocketbase';
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

export const currentUser = writable(pb.authStore.model);

pb.authStore.onChange((token, model) => {
	console.log('Auth state changed:', model);
	currentUser.set(model);
}, true);

import PocketBase from 'pocketbase';
import { writable } from 'svelte/store';

export const pb = new PocketBase('http://127.0.0.1:8090');

// 初始化时直接读取 authStore 的当前模型
export const currentUser = writable(pb.authStore.model);

// 监听 auth 状态变化（登录、退出、Token 过期）
pb.authStore.onChange((token, model) => {
	console.log('Auth state changed:', model);
	currentUser.set(model);
}, true); // true 表示立即触发一次当前状态

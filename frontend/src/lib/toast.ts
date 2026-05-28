import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
	id: number;
	message: string;
	type: ToastType;
}

const { subscribe, update } = writable<Toast[]>([]);

let nextId = 1;

export const toasts = {
	subscribe,
	add: (message: string, type: ToastType = 'info', duration = 3000) => {
		const id = nextId++;
		update((all) => [...all, { id, message, type }]);
		if (duration > 0) {
			setTimeout(() => {
				toasts.remove(id);
			}, duration);
		}
	},
	success: (message: string, duration = 3000) => toasts.add(message, 'success', duration),
	error: (message: string, duration = 5000) => toasts.add(message, 'error', duration),
	warning: (message: string, duration = 3000) => toasts.add(message, 'warning', duration),
	info: (message: string, duration = 3000) => toasts.add(message, 'info', duration),
	remove: (id: number) => {
		update((all) => all.filter((t) => t.id !== id));
	},
	clear: () => {
		update(() => []);
		nextId = 1;
	}
};

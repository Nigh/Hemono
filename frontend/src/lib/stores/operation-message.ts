import { writable } from 'svelte/store';

export type OperationMessageType = 'success' | 'error' | 'warning';
export type OperationMessageScope = 'createLedger' | 'addTransaction' | 'general';

interface OperationMessage {
	type: OperationMessageType;
	text: string;
	scope: OperationMessageScope;
}

function createOperationMessageStore() {
	const { subscribe, set } = writable<OperationMessage | null>(null);
	let timer: ReturnType<typeof setTimeout> | null = null;

	return {
		subscribe,
		show: (
			message: string,
			type: OperationMessageType,
			scope: OperationMessageScope = 'general',
			durationMs: number = 3000,
			onTimeout?: () => void
		) => {
			if (timer) {
				clearTimeout(timer);
				timer = null;
			}
			set({ type, text: message, scope });
			timer = setTimeout(() => {
				onTimeout?.();
				set(null);
				timer = null;
			}, durationMs);
		},
		clear: () => {
			if (timer) {
				clearTimeout(timer);
				timer = null;
			}
			set(null);
		}
	};
}

export const operationMessageStore = createOperationMessageStore();

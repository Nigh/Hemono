import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { operationMessageStore } from './operation-message';

describe('operationMessageStore', () => {
	beforeEach(() => {
		vi.useFakeTimers();
		operationMessageStore.clear();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('should start with null', () => {
		let current: any = 'unset';
		operationMessageStore.subscribe((v) => (current = v))();
		expect(current).toBeNull();
	});

	it('should show a message', () => {
		let current: any = null;
		const unsub = operationMessageStore.subscribe((v) => (current = v));

		operationMessageStore.show('hello', 'success', 'general');

		expect(current).toEqual({
			type: 'success',
			text: 'hello',
			scope: 'general'
		});

		unsub();
	});

	it('should use default scope general', () => {
		let current: any = null;
		const unsub = operationMessageStore.subscribe((v) => (current = v));

		operationMessageStore.show('msg', 'error');

		expect(current.scope).toBe('general');

		unsub();
	});

	it('should clear message after duration', () => {
		let current: any = null;
		const unsub = operationMessageStore.subscribe((v) => (current = v));

		operationMessageStore.show('temporary', 'warning', 'general', 2000);
		expect(current).not.toBeNull();

		vi.advanceTimersByTime(2000);
		expect(current).toBeNull();

		unsub();
	});

	it('should call onTimeout callback', () => {
		const callback = vi.fn();
		const unsub = operationMessageStore.subscribe(() => {});

		operationMessageStore.show('msg', 'success', 'general', 1000, callback);
		vi.advanceTimersByTime(1000);

		expect(callback).toHaveBeenCalledOnce();

		unsub();
	});

	it('should clear message manually', () => {
		let current: any = null;
		const unsub = operationMessageStore.subscribe((v) => (current = v));

		operationMessageStore.show('msg', 'success');
		expect(current).not.toBeNull();

		operationMessageStore.clear();
		expect(current).toBeNull();

		unsub();
	});

	it('should replace previous message on new show()', () => {
		let current: any = null;
		const unsub = operationMessageStore.subscribe((v) => (current = v));

		operationMessageStore.show('first', 'success');
		operationMessageStore.show('second', 'error', 'createLedger');

		expect(current.text).toBe('second');
		expect(current.type).toBe('error');
		expect(current.scope).toBe('createLedger');

		unsub();
	});

	it('should reset timer on new show()', () => {
		let current: any = null;
		const unsub = operationMessageStore.subscribe((v) => (current = v));

		operationMessageStore.show('first', 'success', 'general', 5000);
		vi.advanceTimersByTime(3000);

		operationMessageStore.show('second', 'error', 'general', 5000);
		vi.advanceTimersByTime(3000);
		expect(current).not.toBeNull();

		vi.advanceTimersByTime(2000);
		expect(current).toBeNull();

		unsub();
	});

	it('should support different scopes', () => {
		let current: any = null;
		const unsub = operationMessageStore.subscribe((v) => (current = v));

		operationMessageStore.show('msg', 'success', 'addTransaction');
		expect(current.scope).toBe('addTransaction');

		unsub();
	});
});

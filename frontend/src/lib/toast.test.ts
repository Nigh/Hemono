import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { toasts } from './toast';

describe('toast store', () => {
	beforeEach(() => {
		vi.useFakeTimers();
		toasts.clear();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('should start with empty toasts', () => {
		let current: any[] = [];
		toasts.subscribe((t) => (current = t))();
		expect(current).toEqual([]);
	});

	it('should add a toast with add()', () => {
		let current: any[] = [];
		const unsub = toasts.subscribe((t) => (current = t));

		toasts.add('test message', 'success');

		expect(current).toHaveLength(1);
		expect(current[0].message).toBe('test message');
		expect(current[0].type).toBe('success');

		unsub();
	});

	it('should add toast with default type info', () => {
		let current: any[] = [];
		const unsub = toasts.subscribe((t) => (current = t));

		toasts.add('info message');

		expect(current[0].type).toBe('info');

		unsub();
	});

	it('should add success toast via shortcut', () => {
		let current: any[] = [];
		const unsub = toasts.subscribe((t) => (current = t));

		toasts.success('done');

		expect(current[0].type).toBe('success');
		expect(current[0].message).toBe('done');

		unsub();
	});

	it('should add error toast via shortcut', () => {
		let current: any[] = [];
		const unsub = toasts.subscribe((t) => (current = t));

		toasts.error('fail');

		expect(current[0].type).toBe('error');

		unsub();
	});

	it('should add warning toast via shortcut', () => {
		let current: any[] = [];
		const unsub = toasts.subscribe((t) => (current = t));

		toasts.warning('warn');

		expect(current[0].type).toBe('warning');

		unsub();
	});

	it('should remove a toast by id', () => {
		let current: any[] = [];
		const unsub = toasts.subscribe((t) => (current = t));

		toasts.add('first');
		toasts.add('second');
		expect(current).toHaveLength(2);

		const id = current[0].id;
		toasts.remove(id);
		expect(current).toHaveLength(1);
		expect(current[0].message).toBe('second');

		unsub();
	});

	it('should auto-remove toast after duration', () => {
		let current: any[] = [];
		const unsub = toasts.subscribe((t) => (current = t));

		toasts.add('temporary', 'info', 1000);
		expect(current).toHaveLength(1);

		vi.advanceTimersByTime(1000);
		expect(current).toHaveLength(0);

		unsub();
	});

	it('should not auto-remove when duration is 0', () => {
		let current: any[] = [];
		const unsub = toasts.subscribe((t) => (current = t));

		toasts.add('persistent', 'info', 0);
		expect(current).toHaveLength(1);

		vi.advanceTimersByTime(10000);
		expect(current).toHaveLength(1);

		unsub();
	});

	it('should assign incrementing ids', () => {
		let current: any[] = [];
		const unsub = toasts.subscribe((t) => (current = t));

		toasts.add('a');
		toasts.add('b');
		toasts.add('c');

		expect(current[0].id).toBeLessThan(current[1].id);
		expect(current[1].id).toBeLessThan(current[2].id);

		unsub();
	});
});

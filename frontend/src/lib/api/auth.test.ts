import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { get } from 'svelte/store';
import { pb, currentUser } from '$lib/pb';
import { restoreAuth } from './auth';
import { fetchLedgers } from './ledger';

const record = { id: 'test-user', collectionName: 'users', collectionId: '_pb_users_auth_' };
const token = (expiresIn: number) =>
	`eyJhbGciOiJIUzI1NiJ9.${btoa(JSON.stringify({ exp: Math.floor(Date.now() / 1000) + expiresIn }))}.test`;

beforeEach(() => {
	pb.authStore.save(token(3600), record);
});

afterEach(() => {
	vi.restoreAllMocks();
	pb.authStore.clear();
});

it('clears an expired cached session without making a refresh request', async () => {
	pb.authStore.save(token(-60), record);
	const fetch = vi.spyOn(globalThis, 'fetch');
	await restoreAuth();
	expect(pb.authStore.token).toBe('');
	expect(get(currentUser)).toBeNull();
	expect(fetch).not.toHaveBeenCalled();
});

it('refreshes a valid session and stores the renewed token', async () => {
	const renewed = token(604800);
	vi.spyOn(globalThis, 'fetch').mockResolvedValue(
		new Response(JSON.stringify({ token: renewed, record }), { status: 200 })
	);
	await restoreAuth();
	expect(pb.authStore.token).toBe(renewed);
	expect(get(currentUser)?.id).toBe(record.id);
});

it.each([401, 403])('clears a session rejected by the server with %i', async (status) => {
	vi.spyOn(globalThis, 'fetch').mockResolvedValue(
		new Response(JSON.stringify({ message: 'Invalid authentication' }), { status })
	);
	await restoreAuth();
	expect(get(currentUser)).toBeNull();
});

it('preserves the cached session on a network failure', async () => {
	const original = pb.authStore.token;
	vi.spyOn(globalThis, 'fetch').mockRejectedValue(new TypeError('Failed to fetch'));
	await expect(restoreAuth()).rejects.toThrow();
	expect(pb.authStore.token).toBe(original);
	expect(get(currentUser)?.id).toBe(record.id);
});

it('blocks an expired token before a ledger request can return an empty list', async () => {
	pb.authStore.save(token(-60), record);
	const fetch = vi.spyOn(globalThis, 'fetch');
	await expect(fetchLedgers()).rejects.toMatchObject({ status: 401 });
	expect(fetch).not.toHaveBeenCalled();
	expect(get(currentUser)).toBeNull();
});

import { afterEach, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, waitFor } from '@testing-library/svelte';
import { createRawSnippet } from 'svelte';
import { pb } from '$lib/pb';
import { toasts } from '$lib/toast';
import { restoreAuth } from '$lib/api/auth';
import Layout from './+layout.svelte';

vi.mock('$lib/api/auth', () => ({ restoreAuth: vi.fn() }));

const children = createRawSnippet(() => ({ render: () => '<p>账本内容</p>' }));
const record = { id: 'test-user', collectionName: 'users', collectionId: '_pb_users_auth_' };
const token = `eyJhbGciOiJIUzI1NiJ9.${btoa(JSON.stringify({ exp: Math.floor(Date.now() / 1000) + 3600 }))}.test`;

afterEach(() => {
	cleanup();
	vi.restoreAllMocks();
	vi.resetAllMocks();
	pb.authStore.clear();
});

it('waits for authentication and refreshes on focus without unmounting a valid session', async () => {
	pb.authStore.save(token, record);
	let finish!: () => void;
	vi.mocked(restoreAuth).mockImplementation(() => new Promise((resolve) => (finish = resolve)));
	const view = render(Layout, { children });
	expect(view.queryByText('账本内容')).not.toBeInTheDocument();
	finish();
	await waitFor(() => expect(view.getByText('账本内容')).toBeInTheDocument());
	await fireEvent.focus(window);
	expect(restoreAuth).toHaveBeenCalledTimes(2);
	expect(view.getByText('账本内容')).toBeInTheDocument();
	finish();
	await waitFor(() => expect(view.getByText('账本内容')).toBeInTheDocument());
});

it('cancels pending requests when logging out during a background refresh', async () => {
	pb.authStore.save(token, record);
	vi.mocked(restoreAuth).mockResolvedValueOnce(undefined);
	const cancel = vi.spyOn(pb, 'cancelAllRequests');
	vi.spyOn(toasts, 'success').mockReturnValue(undefined);
	const view = render(Layout, { children });
	await waitFor(() => expect(view.getByText('账本内容')).toBeInTheDocument());
	vi.mocked(restoreAuth).mockImplementation(() => new Promise(() => {}));
	await fireEvent.focus(window);
	await fireEvent.click(view.getByRole('button', { name: '退出登录' }));
	expect(cancel).toHaveBeenCalledOnce();
	expect(pb.authStore.record).toBeNull();
	expect(view.queryByText('账本内容')).not.toBeInTheDocument();
});

it('shows retry instead of protected content when session verification fails', async () => {
	pb.authStore.save('cached-token', record);
	vi.mocked(restoreAuth).mockRejectedValueOnce(new Error('offline')).mockResolvedValue(undefined);
	const view = render(Layout, { children });
	await waitFor(() => expect(view.getByRole('alert')).toHaveTextContent('无法验证登录状态'));
	expect(view.queryByText('账本内容')).not.toBeInTheDocument();
	await fireEvent.click(view.getByRole('button', { name: '重试' }));
	await waitFor(() => expect(view.getByText('账本内容')).toBeInTheDocument());
});

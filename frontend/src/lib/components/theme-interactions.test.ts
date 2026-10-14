import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, waitFor } from '@testing-library/svelte';
import { pb } from '$lib/pb';
import TransactionCard from './TransactionCard.svelte';
import AddTransactionModal from './modals/AddTransactionModal.svelte';
import GenerateInviteModal from './GenerateInviteModal.svelte';
import MembersDrawer from './MembersDrawer.svelte';

beforeEach(() => {
	Object.defineProperty(HTMLDialogElement.prototype, 'showModal', {
		configurable: true,
		writable: true,
		value: vi.fn(function (this: HTMLDialogElement) {
			this.open = true;
		})
	});
	Object.defineProperty(HTMLDialogElement.prototype, 'close', {
		configurable: true,
		writable: true,
		value: vi.fn(function (this: HTMLDialogElement) {
			this.open = false;
		})
	});
});

afterEach(() => {
	cleanup();
	vi.restoreAllMocks();
	pb.authStore.clear();
});

describe('theme interactions', () => {
	it('opens and closes the member drawer with its controlled state', async () => {
		const onclose = vi.fn();
		const view = render(MembersDrawer, { members: [], open: false, onclose });
		const dialog = view.container.querySelector('dialog')!;
		expect(dialog.open).toBe(false);
		await view.rerender({ members: [], open: true, onclose });
		expect(view.getByRole('dialog', { name: '成员列表' })).toBeVisible();
		await fireEvent.click(view.getByRole('button', { name: '关闭成员列表' }));
		expect(onclose).toHaveBeenCalledOnce();
		await view.rerender({ members: [], open: false, onclose });
		expect(dialog.open).toBe(false);
	});

	it('requires explicit confirmation before deleting a transaction', async () => {
		pb.authStore.save('test-token', {
			id: 'payer',
			collectionId: 'users',
			collectionName: 'users'
		});
		const ondelete = vi.fn();
		const showModal = vi.fn(function (this: HTMLDialogElement) {
			this.open = true;
		});
		vi.spyOn(HTMLDialogElement.prototype, 'showModal').mockImplementation(showModal);
		const view = render(TransactionCard, {
			transaction: {
				id: 'transaction',
				amount: 1234,
				type: 'AA',
				date: '2026-10-15',
				expand: { payer: { id: 'payer', email: 'payer@example.com' } }
			},
			ondelete
		});
		await fireEvent.click(view.getByTitle('删除'));
		expect(showModal).toHaveBeenCalledOnce();
		expect(ondelete).not.toHaveBeenCalled();
		await fireEvent.click(view.getByRole('button', { name: '取消' }));
		expect(ondelete).not.toHaveBeenCalled();
		await fireEvent.click(view.getByRole('button', { name: '确认删除' }));
		expect(ondelete).toHaveBeenCalledExactlyOnceWith('transaction');
	});

	it('exposes selected choices and a labelled beneficiary field', async () => {
		const view = render(AddTransactionModal, {
			ledgerId: 'ledger',
			ledgerName: '旅行',
			members: [{ id: 'member', email: 'member@example.com' }]
		});
		const dialog = view.container.querySelector('dialog')!;
		dialog.open = true;
		const expense = view.getByRole('button', { name: '支出' });
		const income = view.getByRole('button', { name: '收入' });
		expect(expense).toHaveAttribute('aria-pressed', 'true');
		await fireEvent.click(income);
		expect(income).toHaveAttribute('aria-pressed', 'true');
		expect(expense).toHaveAttribute('aria-pressed', 'false');
		await fireEvent.click(view.getByRole('button', { name: '单人承担' }));
		expect(view.getByLabelText('由谁承担？')).toBeVisible();
		expect(view.getByLabelText('金额（元）')).toBeVisible();
	});

	it('registers the invitation modal and renders asynchronously loaded codes', async () => {
		vi.spyOn(HTMLDialogElement.prototype, 'showModal').mockImplementation(function (
			this: HTMLDialogElement
		) {
			this.open = true;
		});
		vi.spyOn(pb, 'send').mockResolvedValue({
			exists: true,
			code: 'ABC-123456',
			expiresAt: '2026-10-16',
			maxUses: 10,
			usedCount: 1,
			id: 'invitation'
		});
		const onregister = vi.fn();
		const view = render(GenerateInviteModal, {
			ledgerId: 'ledger',
			ledgerName: '旅行',
			isOwner: true,
			onregister
		});
		await waitFor(() => expect(onregister).toHaveBeenCalledOnce());
		expect(onregister.mock.calls[0][0]).toBe('ledger');
		onregister.mock.calls[0][1].open();
		await waitFor(() => expect(view.getByText('ABC-123456')).toBeVisible());
	});
});

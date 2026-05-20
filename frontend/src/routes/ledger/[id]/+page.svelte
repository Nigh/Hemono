<script lang="ts">
	import { page } from '$app/stores';
	import {
		fetchLedger,
		fetchLedgerMembers,
		fetchTransactions,
		deleteTransaction
	} from '$lib/api/ledger';
	import { toasts } from '$lib/toast';
	import MembersDrawer from '$lib/components/MembersDrawer.svelte';
	import TransactionCard from '$lib/components/TransactionCard.svelte';

	let ledgerId = $derived($page.params.id ?? '');
	let ledger: any = $state(null);
	let members: any[] = $state([]);
	let transactions: any[] = $state([]);
	let isLoading = $state(true);
	let error: string | null = $state(null);
	let drawerOpen = $state(false);

	async function loadData() {
		isLoading = true;
		error = null;
		try {
			[ledger, members, transactions] = await Promise.all([
				fetchLedger(ledgerId),
				fetchLedgerMembers(ledgerId),
				fetchTransactions(ledgerId)
			]);
		} catch (err: any) {
			error = err.message || '加载失败';
		} finally {
			isLoading = false;
		}
	}

	$effect(() => {
		if (ledgerId) {
			loadData();
		}
	});

	async function handleDelete(id: string) {
		try {
			await deleteTransaction(id);
			transactions = transactions.filter((t) => t.id !== id);
			toasts.success('已删除');
		} catch (err: any) {
			toasts.error(`删除失败: ${err.message}`);
		}
	}
</script>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<a href="/" class="btn btn-ghost btn-sm">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="h-4 w-4"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
			返回
		</a>

		{#if !isLoading && !error}
			<button
				class="btn btn-ghost btn-sm btn-circle indicator"
				onclick={() => (drawerOpen = true)}
				title="查看成员"
			>
				<span class="badge badge-sm badge-primary indicator-item">{members.length}</span>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-5 w-5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
					/>
				</svg>
			</button>
		{/if}
	</div>

	{#if isLoading}
		<div class="py-20 flex justify-center">
			<span class="loading loading-spinner loading-md"></span>
		</div>
	{:else if error}
		<div class="alert alert-error">
			<span>{error}</span>
		</div>
	{:else if ledger}
		<div class="card border-base-300 bg-base-100 border">
			<div class="card-body p-4">
				<h2 class="card-title text-lg">{ledger.name}</h2>
				<p class="text-base-content/60 text-xs tracking-widest uppercase">{ledgerId}</p>
			</div>
		</div>

		<div class="space-y-2">
			{#each transactions as tx (tx.id)}
				<TransactionCard transaction={tx} ondelete={handleDelete} />
			{:else}
				<div class="text-center py-10 opacity-30">
					<p>暂无记账记录</p>
				</div>
			{/each}
		</div>
	{/if}
</div>

<MembersDrawer {members} open={drawerOpen} onclose={() => (drawerOpen = false)} />

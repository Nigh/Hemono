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
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	let ledger: any = $state(null);
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	let members: any[] = $state([]);
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	let transactions: any[] = $state([]);
	let isLoading = $state(true);
	let error: string | null = $state(null);
	let drawerOpen = $state(false);

	let searchKeyword = $state('');
	let debouncedKeyword = $state('');
	let searchTimer: ReturnType<typeof setTimeout> | null = null;

	let currentPage = $state(1);
	let totalItems = $state(0);
	let totalPages = $state(0);
	const perPage = 20;

	function debounceSearch(value: string) {
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => {
			debouncedKeyword = value;
			currentPage = 1;
		}, 300);
	}

	$effect(() => {
		debounceSearch(searchKeyword);
	});

	async function loadTransactions() {
		const result = await fetchTransactions({
			ledgerId,
			page: currentPage,
			perPage,
			keyword: debouncedKeyword
		});
		transactions = result.items;
		totalItems = result.totalItems;
		totalPages = result.totalPages;
	}

	async function loadData() {
		isLoading = true;
		error = null;
		try {
			const [ledgerResult, membersResult] = await Promise.all([
				fetchLedger(ledgerId),
				fetchLedgerMembers(ledgerId)
			]);
			ledger = ledgerResult;
			members = membersResult;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : '加载失败';
		} finally {
			isLoading = false;
		}
	}

	$effect(() => {
		if (ledgerId) {
			loadData();
		}
	});

	$effect(() => {
		void debouncedKeyword;
		void currentPage;
		if (!isLoading && ledgerId) {
			loadTransactions();
		}
	});

	async function handleDelete(id: string) {
		try {
			await deleteTransaction(id);
			toasts.success('已删除');
			await loadTransactions();
		} catch (err: unknown) {
			toasts.error(`删除失败: ${err instanceof Error ? err.message : '未知错误'}`);
		}
	}

	function goToPage(page: number) {
		if (page >= 1 && page <= totalPages) {
			currentPage = page;
		}
	}

	function clearSearch() {
		searchKeyword = '';
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

		<label class="input input-bordered gap-2 flex w-full items-center">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="h-4 w-4 opacity-40"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
				/>
			</svg>
			<input
				type="text"
				class="grow"
				placeholder="搜索备注、付款人、金额…"
				bind:value={searchKeyword}
			/>
			{#if searchKeyword}
				<button class="btn btn-ghost btn-xs btn-circle" onclick={clearSearch} title="清除搜索">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-4 w-4"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
				</button>
			{/if}
		</label>

		{#if totalItems > 0}
			<div class="text-base-content/50 text-xs">
				共 {totalItems} 条记录{#if totalPages > 1}，第 {currentPage}/{totalPages} 页{/if}
			</div>
		{/if}

		<div class="space-y-2">
			{#each transactions as tx (tx.id)}
				<TransactionCard transaction={tx} ondelete={handleDelete} />
			{:else}
				<div class="text-center py-10 opacity-30">
					<p>{debouncedKeyword ? '没有匹配的记录' : '暂无记账记录'}</p>
				</div>
			{/each}
		</div>

		{#if totalPages > 1}
			<div class="flex justify-center">
				<div class="join">
					<button
						class="join-item btn btn-sm"
						disabled={currentPage <= 1}
						onclick={() => goToPage(currentPage - 1)}
					>
						«
					</button>
					{#each Array.from({ length: totalPages }, (_, i) => i + 1) as p (p)}
						{#if totalPages <= 7 || p === 1 || p === totalPages || Math.abs(p - currentPage) <= 1}
							<button
								class="join-item btn btn-sm"
								class:btn-active={p === currentPage}
								onclick={() => goToPage(p)}
							>
								{p}
							</button>
					{:else if p === 2 && currentPage > 3}
						<button class="join-item btn btn-sm btn-disabled">…</button>
					{:else if p === totalPages - 1 && currentPage < totalPages - 2}
							<button class="join-item btn btn-sm btn-disabled">…</button>
						{/if}
					{/each}
					<button
						class="join-item btn btn-sm"
						disabled={currentPage >= totalPages}
						onclick={() => goToPage(currentPage + 1)}
					>
						»
					</button>
				</div>
			</div>
		{/if}
	{/if}
</div>

<MembersDrawer
	{members}
	open={drawerOpen}
	onclose={() => (drawerOpen = false)}
	ownerId={ledger?.owner}
/>

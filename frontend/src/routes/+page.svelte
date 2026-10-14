<script lang="ts">
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/pb';
	import { toasts } from '$lib/toast';
	import { fetchLedgers, fetchLedgerMembers } from '$lib/api/ledger';
	import GenerateInviteModal from '$lib/components/GenerateInviteModal.svelte';
	import CreateLedgerModal from '$lib/components/modals/CreateLedgerModal.svelte';
	import AddTransactionModal from '$lib/components/modals/AddTransactionModal.svelte';
	import LedgerCard from '$lib/components/ledger/LedgerCard.svelte';
	import LedgerExpandedContent from '$lib/components/ledger/LedgerExpandedContent.svelte';

	let ledgers: any[] = $state([]);
	let members: any[] = $state([]);
	let expandedLedger = $state('');
	let activeLedgerName = $state('');
	let statsRefreshCounter = $state(0);
	let inviteModals: Map<string, { open: () => void; resetState: () => void }> = new Map();

	function handleRegisterInviteModal(
		ledgerId: string,
		methods: { open: () => void; resetState: () => void }
	) {
		inviteModals.set(ledgerId, methods);
	}

	onMount(() => {
		loadLedgers();
	});

	async function loadLedgers() {
		try {
			ledgers = await fetchLedgers();
		} catch (err: any) {
			console.error('Failed to load ledgers:', err);
			toasts.error(`加载账本失败: ${err.message || '未知错误'}`);
		}
	}

	$effect(() => {
		if (expandedLedger) {
			const ledger = ledgers.find((l) => l.id === expandedLedger);
			if (ledger) {
				activeLedgerName = ledger.name;
			}
			fetchLedgerMembers(expandedLedger)
				.then((fetchedMembers) => {
					members = fetchedMembers;
				})
				.catch((err) => {
					console.error('Failed to load members:', err);
					toasts.error(`加载成员失败: ${err.message || '未知错误'}`);
				});
		}
	});

	function toggleLedger(ledgerId: string) {
		if (expandedLedger === ledgerId) {
			expandedLedger = '';
			activeLedgerName = '';
		} else {
			expandedLedger = ledgerId;
		}
	}

	function openCreateLedgerModal() {
		(window as any).create_ledger_modal.showModal();
	}

	function openAddTransactionModal() {
		(window as any).add_modal.showModal();
	}

	function openGenerateInviteModal(ledgerId: string) {
		const modal = inviteModals.get(ledgerId);
		if (modal) {
			modal.open();
		}
	}

	function handleLedgerCreated() {
		loadLedgers();
	}

	function handleTransactionAdded() {
		statsRefreshCounter++;
	}
</script>

<div class="space-y-4">
	<div class="mb-6 gap-4 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold">我的账本</h1>
			<p class="text-text-muted mt-2 text-sm">共同的开销，清楚的记录。</p>
		</div>
		<button class="btn btn-primary btn-sm shrink-0" onclick={openCreateLedgerModal}>新建账本</button
		>
	</div>
	<div class="gap-3 grid">
		{#each ledgers as ledger}
			<div
				class="card bg-base-100 border transition-colors duration-200 {expandedLedger === ledger.id
					? 'border-primary'
					: 'border-base-300 hover:border-primary/50'}"
			>
				<LedgerCard
					{ledger}
					isExpanded={expandedLedger === ledger.id}
					ontoggle={() => toggleLedger(ledger.id)}
				/>

				{#if expandedLedger === ledger.id}
					<LedgerExpandedContent
						{ledger}
						{members}
						isOwner={$currentUser?.id === ledger.owner}
						onaddtransaction={openAddTransactionModal}
						ongenerateinvite={() => openGenerateInviteModal(ledger.id)}
						{statsRefreshCounter}
					/>
				{/if}
			</div>
		{:else}
			<div class="card bg-base-100 border-base-300 border py-16 text-center text-text-muted">
				<p>还没有账本，点击创建一个</p>
			</div>
		{/each}
	</div>
</div>

<CreateLedgerModal oncreated={handleLedgerCreated} />

<AddTransactionModal
	ledgerId={expandedLedger}
	ledgerName={activeLedgerName}
	{members}
	onadded={handleTransactionAdded}
/>

{#each ledgers as ledger}
	<GenerateInviteModal
		ledgerId={ledger.id}
		ledgerName={ledger.name}
		isOwner={$currentUser?.id === ledger.owner}
		onregister={handleRegisterInviteModal}
	/>
{/each}

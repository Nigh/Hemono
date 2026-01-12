<script lang="ts">
	import { pb, currentUser } from '$lib/pb';
	import { onMount } from 'svelte';
	import { toasts } from '$lib/toast';
	import GenerateInviteModal from '$lib/components/GenerateInviteModal.svelte';

	type OperationMessageType = 'success' | 'error' | 'warning';
	type OperationMessageScope = 'createLedger' | 'addTransaction';

	let ledgers: any[] = [];
	let members: any[] = [];
	let expandedLedger = '';
	let activeLedgerName = ''; // 当前展开的账本名称
	let inviteModals: Map<string, { open: () => void; resetState: () => void }> = new Map();

	function handleRegisterInviteModal(e: CustomEvent) {
		const { ledgerId, methods } = e.detail;
		inviteModals.set(ledgerId, methods);
	}

	let operationMessage: {
		type: OperationMessageType;
		text: string;
		scope: OperationMessageScope;
	} | null = null;
	let operationMessageTimer: ReturnType<typeof setTimeout> | null = null;

	function clearOperationMessage() {
		if (operationMessageTimer) {
			clearTimeout(operationMessageTimer);
			operationMessageTimer = null;
		}
		operationMessage = null;
	}

	function showOperationMessage(
		message: { type: OperationMessageType; text: string; scope: OperationMessageScope },
		durationMs: number,
		onTimeout?: () => void
	) {
		clearOperationMessage();
		operationMessage = message;
		operationMessageTimer = setTimeout(() => {
			onTimeout?.();
			operationMessage = null;
			operationMessageTimer = null;
		}, durationMs);
	}

	function operationMessageAlertClass(type: OperationMessageType) {
		if (type === 'success') return 'alert-success';
		if (type === 'error') return 'alert-error';
		return 'alert-warning';
	}

	function openCreateLedgerModal() {
		clearOperationMessage();
		(window as any).create_ledger_modal.showModal();
	}

	function openAddTransactionModal() {
		clearOperationMessage();
		(window as any).add_modal.showModal();
	}

	// 打开生成邀请码模态框
	function openGenerateInviteModal(ledgerId: string, ledgerName: string) {
		const modal = inviteModals.get(ledgerId);
		if (modal) {
			modal.open();
		}
	}

	function onCreateLedgerModalClose() {
		if (operationMessage?.scope === 'createLedger') clearOperationMessage();
	}

	function onAddTransactionModalClose() {
		if (operationMessage?.scope === 'addTransaction') clearOperationMessage();
	}

	let isCreateLedgerSuccess = false;
	let isAddTransactionSuccess = false;

	$: isCreateLedgerSuccess =
		operationMessage?.scope === 'createLedger' && operationMessage?.type === 'success';
	$: isAddTransactionSuccess =
		operationMessage?.scope === 'addTransaction' && operationMessage?.type === 'success';

	// 表单变量
	let newLedgerName = '';
	let amount = '';
	let note = '';
	let splitType = 'AA';
	let beneficiary = '';

	onMount(() => {
		loadLedgers();
	});

	async function loadLedgers() {
		try {
			ledgers = await pb.collection('ledgers').getFullList({ sort: '-created' });
		} catch (err: any) {
			console.error('Failed to load ledgers:', err);
			toasts.error(`加载账本失败: ${err.message || '未知错误'}`);
		}
	}

	// 监听账本选择，更新成员列表
	$: if (expandedLedger) {
		const ledger = ledgers.find((l) => l.id === expandedLedger);
		if (ledger) {
			activeLedgerName = ledger.name;
		}
		pb.collection('ledger_members')
			.getFullList({
				filter: `ledger = "${expandedLedger}"`,
				expand: 'user'
			})
			.then((res) => {
				members = res.map((m) => m.expand.user);
				if (!beneficiary && $currentUser) beneficiary = $currentUser.id;
			})
			.catch((err) => {
				console.error('Failed to load members:', err);
				toasts.error(`加载成员失败: ${err.message || '未知错误'}`);
			});
	}

	function toggleLedger(ledgerId: string) {
		if (expandedLedger === ledgerId) {
			expandedLedger = '';
			activeLedgerName = '';
		} else {
			expandedLedger = ledgerId;
		}
	}

	async function createLedger() {
		if (!newLedgerName) {
			showOperationMessage(
				{ type: 'warning', text: '请输入账本名称', scope: 'createLedger' },
				3000
			);
			return;
		}
		try {
			await pb.collection('ledgers').create({ name: newLedgerName, owner: $currentUser.id });
			newLedgerName = '';
			loadLedgers();
			showOperationMessage(
				{ type: 'success', text: '账本创建成功', scope: 'createLedger' },
				1500,
				() => (window as any).create_ledger_modal.close()
			);
		} catch (err: any) {
			console.error('Failed to create ledger:', err);
			showOperationMessage(
				{
					type: 'error',
					text: `创建账本失败: ${err.message || '未知错误'}`,
					scope: 'createLedger'
				},
				3000
			);
		}
	}

	async function addTransaction() {
		if (!expandedLedger || !amount) {
			showOperationMessage({ type: 'warning', text: '请输入金额', scope: 'addTransaction' }, 3000);
			return;
		}
		try {
			amount = parseFloat(amount).toFixed(2);
			await pb.collection('transactions').create({
				ledger: expandedLedger,
				payer: $currentUser.id,
				amount: Math.round(Number(amount) * 100),
				type: splitType,
				beneficiary: splitType === 'SINGLE' ? beneficiary : null,
				note,
				date: new Date()
			});

			showOperationMessage(
				{ type: 'success', text: '记账成功', scope: 'addTransaction' },
				1500,
				() => {
					(window as any).add_modal.close();
					amount = '';
					note = '';
					expandedLedger = '';
					activeLedgerName = '';
				}
			);
		} catch (err: any) {
			console.error('Failed to add transaction:', err);
			showOperationMessage(
				{ type: 'error', text: `记账失败: ${err.message || '未知错误'}`, scope: 'addTransaction' },
				3000
			);
		}
	}
</script>

<div class="space-y-4">
	<div class="gap-3 grid">
		{#each ledgers as ledger}
			<div
				class="card border transition-all duration-300 select-none {expandedLedger === ledger.id
					? 'border-primary ring-primary ring-2'
					: 'border-base hover:ring-2'}"
			>
				<!-- 账本头部 -->
				<button
					class="card-body p-4 w-full flex-row items-center justify-between text-left"
					on:click={() => toggleLedger(ledger.id)}
				>
					<div>
						<h3 class="font-bold text-xl">{ledger.name}</h3>
						<p class="tracking-widest text-xs uppercase opacity-40">{ledger.id}</p>
					</div>
					<div class="gap-2 flex items-center">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-5 w-5 transition-transform duration-200 {expandedLedger === ledger.id
								? 'rotate-180'
								: ''}"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M19 9l-7 7-7-7"
							/>
						</svg>
					</div>
				</button>

				<!-- 展开内容 -->
				{#if expandedLedger === ledger.id}
					<div class="border-base-300 p-4 space-y-4 border-t">
						<div class="flex items-center justify-between">
							<span class="text-sm opacity-60">快速记账</span>
							<button class="btn btn-primary btn-sm shadow-lg" on:click={openAddTransactionModal}>
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
										d="M12 4v16m8-8H4"
									/>
								</svg>
								记一笔
							</button>
						</div>

						<!-- 成员列表 -->
						<div class="space-y-2">
							<span class="text-xs font-medium opacity-60">成员 ({members.length})</span>
							<div class="gap-2 flex items-center">
								<div class="avatar-group -space-x-4">
									{#each members as member}
										<div class="avatar">
											<div class="w-8 h-8 border-primary rounded-full border-2">
												<img
													src={member.avatar
														? pb.files.getURL(member, member.avatar)
														: `https://api.dicebear.com/7.x/bottts/svg?seed=${member.id}`}
													alt="avatar"
												/>
											</div>
										</div>
									{/each}
								</div>
								<button
									class="btn btn-xs btn-primary btn-outline"
									on:click={() => openGenerateInviteModal(ledger.id, ledger.name)}
									title="邀请成员"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										class="h-4 w-4"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										><path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M12 4v16m8-8H4"
										/></svg
									>
								</button>
							</div>
						</div>
					</div>
				{/if}
			</div>
		{:else}
			<div class="text-center py-20 opacity-30">
				<p>还没有账本，点击创建一个</p>
			</div>
		{/each}
		<button class="btn btn-sm btn-ghost" on:click={openCreateLedgerModal} title="新建账本">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="h-4 w-4"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
				><path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="3"
					d="M12 4v16m8-8H4"
				/></svg
			>
		</button>
	</div>
</div>

<dialog
	id="create_ledger_modal"
	class="modal modal-bottom sm:modal-middle"
	on:close={onCreateLedgerModalClose}
>
	<div class="modal-box border">
		<h3 class="text-lg font-bold mb-2">新建账本</h3>

		{#if operationMessage && operationMessage.scope === 'createLedger'}
			<div role="alert" class="alert mb-4 {operationMessageAlertClass(operationMessage.type)}">
				<span>{operationMessage.text}</span>
			</div>
		{/if}

		<div class="py-4">
			<input
				type="text"
				bind:value={newLedgerName}
				placeholder="账本名称，如：冰岛行、上海合租"
				class="input w-full"
				disabled={isCreateLedgerSuccess}
			/>
		</div>
		<div class="modal-action">
			<button
				class="btn btn-primary btn-block"
				on:click={createLedger}
				disabled={isCreateLedgerSuccess}>确认创建</button
			>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button disabled={isCreateLedgerSuccess}>关闭</button>
	</form>
</dialog>

<dialog
	id="add_modal"
	class="modal modal-bottom sm:modal-middle"
	on:close={onAddTransactionModalClose}
>
	<div class="modal-box border">
		<h3 class="text-lg font-bold mb-2">为「{activeLedgerName}」记一笔</h3>

		{#if operationMessage && operationMessage.scope === 'addTransaction'}
			<div role="alert" class="alert mb-4 {operationMessageAlertClass(operationMessage.type)}">
				<span>{operationMessage.text}</span>
			</div>
		{/if}

		<div
			class="space-y-4"
			class:mt-4={!(operationMessage && operationMessage.scope === 'addTransaction')}
		>
			<input
				type="number"
				bind:value={amount}
				placeholder="金额"
				class="input input-bordered w-full"
				disabled={isAddTransactionSuccess}
			/>

			<div class="tabs tabs-boxed bg-base-200">
				<button
					class="tab flex-1 {splitType === 'AA' ? 'tab-active' : ''}"
					on:click={() => (splitType = 'AA')}
					disabled={isAddTransactionSuccess}>AA 均摊</button
				>
				<button
					class="tab flex-1 {splitType === 'SINGLE' ? 'tab-active' : ''}"
					on:click={() => (splitType = 'SINGLE')}
					disabled={isAddTransactionSuccess}>单人承担</button
				>
			</div>

			{#if splitType === 'SINGLE'}
				<div class="bg-base-200 border-primary/10 rounded-xl p-3 border">
					<label class="label pt-0"><span class="label-text-alt font-bold">由谁承担？</span></label>
					<select
						class="select select-sm select-ghost w-full"
						bind:value={beneficiary}
						disabled={isAddTransactionSuccess}
					>
						{#each members as m}
							<option value={m.id}
								>{m.name || m.email} {m.id === $currentUser.id ? '(自己)' : ''}</option
							>
						{/each}
					</select>
				</div>
			{/if}

			<input
				type="text"
				bind:value={note}
				placeholder="写点备注..."
				class="input input-bordered w-full"
				disabled={isAddTransactionSuccess}
			/>
		</div>
		<div class="modal-action">
			<button
				class="btn btn-primary btn-block shadow-lg"
				on:click={addTransaction}
				disabled={isAddTransactionSuccess}>记一笔</button
			>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button disabled={isAddTransactionSuccess}>关闭</button>
	</form>
</dialog>

{#each ledgers as ledger}
	<GenerateInviteModal
		ledgerId={ledger.id}
		ledgerName={ledger.name}
		on:register={handleRegisterInviteModal}
	/>
{/each}

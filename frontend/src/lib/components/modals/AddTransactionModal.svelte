<script lang="ts">
	import { addTransaction } from '$lib/api/ledger';
	import { currentUser } from '$lib/pb';
	import OperationMessage from '$lib/components/OperationMessage.svelte';
	import type { OperationMessageType } from '$lib/stores/operation-message';

	interface Member {
		id: string;
		name?: string;
		email: string;
	}

	interface Props {
		ledgerId: string;
		ledgerName: string;
		members: Member[];
		onadded?: () => void;
	}

	let { ledgerId, ledgerName, members, onadded }: Props = $props();

	let amount = $state('');
	let note = $state('');
	let splitType = $state('AA');
	let beneficiary = $state('');
	let operationMessage = $state<{ type: OperationMessageType; text: string } | null>(null);
	let operationMessageTimer: ReturnType<typeof setTimeout> | null = null;
	let isSuccess = $state(false);

	$effect(() => {
		if (members.length > 0 && !beneficiary && $currentUser) {
			beneficiary = $currentUser.id;
		}
	});

	function clearOperationMessage() {
		if (operationMessageTimer) {
			clearTimeout(operationMessageTimer);
			operationMessageTimer = null;
		}
		operationMessage = null;
	}

	function showOperationMessage(
		message: { type: OperationMessageType; text: string },
		durationMs: number,
		onTimeout?: () => void
	) {
		clearOperationMessage();
		operationMessage = message;
		isSuccess = message.type === 'success';
		operationMessageTimer = setTimeout(() => {
			onTimeout?.();
			operationMessage = null;
			operationMessageTimer = null;
			isSuccess = false;
		}, durationMs);
	}

	function onClose() {
		if (!isSuccess) {
			clearOperationMessage();
		}
	}

	async function handleAddTransaction() {
		if (!amount) {
			showOperationMessage({ type: 'warning', text: '请输入金额' }, 3000);
			return;
		}
		try {
			const parsedAmount = parseFloat(amount).toFixed(2);
			await addTransaction({
				ledger: ledgerId,
				payer: $currentUser.id,
				amount: Math.round(Number(parsedAmount) * 100),
				type: splitType,
				beneficiary: splitType === 'SINGLE' ? beneficiary : null,
				note,
				date: new Date()
			});

			showOperationMessage({ type: 'success', text: '记账成功' }, 1500, () => {
				(window as any).add_modal.close();
				amount = '';
				note = '';
				onadded?.();
			});
		} catch (err: any) {
			console.error('Failed to add transaction:', err);
			showOperationMessage({ type: 'error', text: `记账失败: ${err.message || '未知错误'}` }, 3000);
		}
	}
</script>

<dialog id="add_modal" class="modal modal-bottom sm:modal-middle" onclose={onClose}>
	<div class="modal-box border">
		<h3 class="text-lg font-bold mb-2">为「{ledgerName}」记一笔</h3>

		{#if operationMessage}
			<OperationMessage type={operationMessage.type} message={operationMessage.text} />
		{/if}

		<div class="space-y-4" class:mt-4={!operationMessage}>
			<input
				type="number"
				bind:value={amount}
				placeholder="金额"
				class="input input-bordered w-full"
				disabled={isSuccess}
			/>

			<div class="tabs tabs-boxed bg-base-200">
				<button
					class="tab flex-1 {splitType === 'AA' ? 'tab-active' : ''}"
					onclick={() => (splitType = 'AA')}
					disabled={isSuccess}>AA 均摊</button
				>
				<button
					class="tab flex-1 {splitType === 'SINGLE' ? 'tab-active' : ''}"
					onclick={() => (splitType = 'SINGLE')}
					disabled={isSuccess}>单人承担</button
				>
			</div>

			{#if splitType === 'SINGLE'}
				<div class="bg-base-200 border-primary/10 rounded-xl p-3 border">
					<label class="label pt-0"><span class="label-text-alt font-bold">由谁承担？</span></label>
					<select
						class="select select-sm select-ghost w-full"
						bind:value={beneficiary}
						disabled={isSuccess}
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
				disabled={isSuccess}
			/>
		</div>
		<div class="modal-action">
			<button
				class="btn btn-primary btn-block shadow-lg"
				onclick={handleAddTransaction}
				disabled={isSuccess}>记一笔</button
			>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button disabled={isSuccess}>关闭</button>
	</form>
</dialog>

<script lang="ts">
	import { createLedger } from '$lib/api/ledger';
	import { currentUser } from '$lib/pb';
	import { toasts } from '$lib/toast';
	import OperationMessage from '$lib/components/OperationMessage.svelte';
	import type { OperationMessageType } from '$lib/stores/operation-message';

	interface Props {
		oncreated?: (ledger: any) => void;
	}

	let { oncreated }: Props = $props();

	let newLedgerName = $state('');
	let operationMessage = $state<{ type: OperationMessageType; text: string } | null>(null);
	let operationMessageTimer: ReturnType<typeof setTimeout> | null = null;
	let isSuccess = $state(false);

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

	async function handleCreateLedger() {
		if (!newLedgerName) {
			showOperationMessage({ type: 'warning', text: '请输入账本名称' }, 3000);
			return;
		}
		try {
			const ledger = await createLedger(newLedgerName, $currentUser.id);
			newLedgerName = '';
			showOperationMessage({ type: 'success', text: '账本创建成功' }, 1500, () => {
				(window as any).create_ledger_modal.close();
				oncreated?.(ledger);
			});
		} catch (err: any) {
			console.error('Failed to create ledger:', err);
			showOperationMessage(
				{
					type: 'error',
					text: `创建账本失败: ${err.message || '未知错误'}`
				},
				3000
			);
		}
	}
</script>

<dialog id="create_ledger_modal" class="modal modal-bottom sm:modal-middle" onclose={onClose}>
	<div class="modal-box border">
		<h3 class="text-lg font-bold mb-2">新建账本</h3>

		{#if operationMessage}
			<OperationMessage type={operationMessage.type} message={operationMessage.text} />
		{/if}

		<div class="py-4">
			<input
				type="text"
				bind:value={newLedgerName}
				placeholder="账本名称，如：冰岛行、上海合租"
				class="input w-full"
				disabled={isSuccess}
			/>
		</div>
		<div class="modal-action">
			<button class="btn btn-primary btn-block" onclick={handleCreateLedger} disabled={isSuccess}
				>确认创建</button
			>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button disabled={isSuccess}>关闭</button>
	</form>
</dialog>

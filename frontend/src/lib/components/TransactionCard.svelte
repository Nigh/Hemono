<script lang="ts">
	import { pb } from '$lib/pb';

	interface Transaction {
		id: string;
		amount: number;
		type: string;
		direction?: string;
		note?: string;
		date: string;
		expand?: { payer?: { id: string; name?: string; email: string; avatar?: string } };
	}

	interface Props {
		transaction: Transaction;
		ondelete?: (id: string) => void;
	}

	let { transaction, ondelete }: Props = $props();

	const payer = $derived(transaction.expand?.payer);
	const amountYuan = $derived((transaction.amount / 100).toFixed(2));
	const dateStr = $derived(new Date(transaction.date).toLocaleDateString('zh-CN'));
	const isIncome = $derived(transaction.direction === 'INCOME');
	const canDelete = $derived(payer?.id === pb.authStore.record?.id);

	function handleDelete() {
		if (confirm('确定要删除这笔记录吗？')) {
			ondelete?.(transaction.id);
		}
	}
</script>

<div class="card border-base-300 bg-base-100 border">
	<div class="card-body p-3 gap-2">
		<div class="flex items-center justify-between">
			<div class="gap-2 min-w-0 flex items-center">
				{#if payer}
					<div class="avatar">
						<div class="w-6 h-6 rounded-full">
							<img
								src={payer.avatar
									? pb.files.getURL(payer, payer.avatar)
									: `https://api.dicebear.com/9.x/pixel-art/svg?seed=${payer.id}`}
								alt="avatar"
							/>
						</div>
					</div>
					<span class="text-sm truncate">{payer.name || payer.email}</span>
				{/if}
				{#if transaction.note}
					<span class="text-xs truncate opacity-40">· {transaction.note}</span>
				{/if}
			</div>

			<div class="gap-2 flex items-center">
				<span class="text-base font-bold tabular-nums {isIncome ? 'text-success' : 'text-error'}">
					{isIncome ? '+' : '-'}¥{amountYuan}
				</span>
				{#if canDelete}
					<button
						class="btn btn-ghost btn-xs btn-circle text-error"
						onclick={handleDelete}
						title="删除"
					>
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
								d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
							/>
						</svg>
					</button>
				{/if}
			</div>
		</div>

		<div class="gap-2 text-xs flex items-center opacity-40">
			<span>{dateStr}</span>
			<span class="badge badge-xs {isIncome ? 'badge-success' : 'badge-error'}">
				{isIncome ? '收入' : '支出'}
			</span>
			<span class="badge badge-xs">{transaction.type === 'AA' ? 'AA均摊' : '单人承担'}</span>
		</div>
	</div>
</div>

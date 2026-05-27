<script lang="ts">
	import { pb } from '$lib/pb';
	import type { MemberStat, LedgerStat } from '$lib/api/stats';

	interface Props {
		stats: LedgerStat | null;
		isLoading: boolean;
		error: string | null;
		onretry?: () => void;
		ownerId?: string;
	}

	let { stats, isLoading, error, onretry, ownerId }: Props = $props();

	function formatAmount(cents: number): string {
		return (cents / 100).toFixed(2);
	}

	function getBalanceClass(balance: number): string {
		if (balance > 0) return 'text-success';
		if (balance < 0) return 'text-error';
		return 'text-gray-500';
	}

	function getBalanceIcon(balance: number): string {
		if (balance > 0) return '✓ 应收';
		if (balance < 0) return '✗ 应付';
		return '平衡';
	}
</script>

{#if isLoading}
	<div class="gap-4 flex flex-col">
		<div class="gap-4 flex">
			<div class="skeleton h-12 rounded-lg w-1/2"></div>
			<div class="skeleton h-12 rounded-lg w-1/2"></div>
		</div>
		<div class="space-y-2">
			{#each Array(3) as _}
				<div class="skeleton h-16 rounded-lg w-full"></div>
			{/each}
		</div>
	</div>
{:else if error}
	<div class="alert alert-error">
		<svg
			xmlns="http://www.w3.org/2000/svg"
			class="h-6 w-6 shrink-0 stroke-current"
			fill="none"
			viewBox="0 0 24 24"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
			/>
		</svg>
		<span>{error}</span>
		{#if onretry}
			<button class="btn btn-sm btn-ghost" onclick={onretry}>重试</button>
		{/if}
	</div>
{:else if stats}
	<div class="gap-4 flex flex-col">
		<!-- 统计摘要 -->
		<div class="gap-4 sm:grid-cols-2 grid">
			<div class="rounded-lg bg-base-200 p-4">
				<div class="text-xs font-medium opacity-60">总支出</div>
				<div class="text-xl font-bold text-error">¥{formatAmount(stats.totalExpense)}</div>
			</div>
			<div class="rounded-lg bg-base-200 p-4">
				<div class="text-xs font-medium opacity-60">总收入</div>
				<div class="text-xl font-bold text-success">¥{formatAmount(stats.totalIncome)}</div>
			</div>
			<div class="rounded-lg bg-base-200 p-4">
				<div class="text-xs font-medium opacity-60">月支出</div>
				<div class="text-xl font-bold text-error">¥{formatAmount(stats.monthlyExpense)}</div>
			</div>
			<div class="rounded-lg bg-base-200 p-4">
				<div class="text-xs font-medium opacity-60">月收入</div>
				<div class="text-xl font-bold text-success">¥{formatAmount(stats.monthlyIncome)}</div>
			</div>
		</div>

		<!-- 成员结余列表 -->
		<div class="rounded-lg bg-base-200 p-4">
			<div class="mb-3 text-xs font-medium opacity-60">成员结余</div>
			<div class="space-y-2">
				{#each stats.memberStats as member}
					<div class="gap-3 rounded-lg bg-base-100 p-3 flex items-center">
						<div class="avatar">
							<div
								class="w-10 h-10 rounded-full {member.userId === ownerId
									? 'ring-warning ring-offset-base-100 ring-2 ring-offset-1'
									: 'border-primary border-2'}"
							>
								<img
									src={member.avatar
										? pb.files.getURL(
												{
													id: member.userId,
													collectionId: '',
													collectionName: 'users'
												},
												member.avatar
											)
										: `https://api.dicebear.com/7.x/bottts/svg?seed=${member.userId}`}
									alt="avatar"
								/>
							</div>
							{#if member.userId === ownerId}
								<div
									class="-top-0.5 -right-0.5 bg-base-100 shadow-sm absolute z-10 rounded-full p-px"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 24 24"
										fill="currentColor"
										class="w-4 h-4 text-warning drop-shadow"
									>
										<path d="M2.5 18.5l2-10 5 4 2.5-6 2.5 6 5-4 2 10z" />
									</svg>
								</div>
							{/if}
						</div>
						<div class="flex flex-1 flex-col">
							<div class="font-medium">{member.name || member.email}</div>
							<div class="text-xs opacity-60">
								支出: ¥{formatAmount(member.totalExpense)}
								{#if member.totalIncome > 0}
									| 收入: ¥{formatAmount(member.totalIncome)}
								{/if}
							</div>
						</div>
						<div class="text-right">
							<div class="font-bold {getBalanceClass(member.balance)}">
								{#if member.balance > 0}
									¥{formatAmount(member.balance)}
								{:else if member.balance < 0}
									-¥{formatAmount(Math.abs(member.balance))}
								{:else}
									¥0.00
								{/if}
							</div>
							<div class="text-xs {getBalanceClass(member.balance)}">
								{getBalanceIcon(member.balance)}
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
{/if}

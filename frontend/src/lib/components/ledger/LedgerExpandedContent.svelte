<script lang="ts">
	import { onMount } from 'svelte';
	import { pb } from '$lib/pb';
	import { fetchLedgerStats } from '$lib/api/stats';
	import type { LedgerStat } from '$lib/api/stats';
	import LedgerStats from './LedgerStats.svelte';

	interface Member {
		id: string;
		name?: string;
		email: string;
		avatar?: string;
	}

	interface Props {
		ledger: {
			id: string;
			name: string;
		};
		members: Member[];
		onaddtransaction?: () => void;
		ongenerateinvite?: () => void;
	}

	let { ledger, members, onaddtransaction, ongenerateinvite }: Props = $props();

	let stats: LedgerStat | null = $state(null);
	let isLoadingStats = $state(false);
	let statsError: string | null = $state(null);

	export async function loadStats() {
		isLoadingStats = true;
		statsError = null;
		try {
			stats = await fetchLedgerStats(ledger.id);
		} catch (error: any) {
			statsError = error.message || '加载统计失败';
			console.error('Failed to load stats:', error);
		} finally {
			isLoadingStats = false;
		}
	}

	onMount(() => {
		loadStats();
	});
</script>

<div class="border-base-300 p-4 space-y-4 border-t">
	<!-- 统计信息 -->
	<LedgerStats {stats} isLoading={isLoadingStats} error={statsError} onretry={loadStats} />

	<div class="flex items-center justify-between">
		<span class="text-sm opacity-60">快速记账</span>
		<button class="btn btn-primary btn-sm shadow-lg" onclick={onaddtransaction}>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="h-4 w-4"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
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
									: `https://api.dicebear.com/9.x/pixel-art/svg?seed=${member.id}`}
								alt="avatar"
							/>
						</div>
					</div>
				{/each}
			</div>
			<button
				class="btn btn-xs btn-primary btn-outline"
				onclick={ongenerateinvite}
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

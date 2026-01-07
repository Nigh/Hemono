<script lang="ts">
	import { pb, currentUser } from '$lib/pb';
	import { onMount } from 'svelte';

	let ledgers: any[] = [];
	let members: any[] = [];

	// 表单变量
	let newLedgerName = '';
	let selectedLedger = '';
	let amount = '';
	let note = '';
	let splitType = 'AA';
	let beneficiary = '';

	onMount(() => {
		loadLedgers();
	});

	async function loadLedgers() {
		ledgers = await pb.collection('ledgers').getFullList({ sort: '-created' });
	}

	// 监听账本选择，更新成员列表
	$: if (selectedLedger) {
		pb.collection('ledger_members')
			.getFullList({
				filter: `ledger = "${selectedLedger}"`,
				expand: 'user'
			})
			.then((res) => {
				members = res.map((m) => m.expand.user);
				if (!beneficiary && $currentUser) beneficiary = $currentUser.id;
			});
	}

	async function createLedger() {
		if (!newLedgerName) return;
		await pb.collection('ledgers').create({ name: newLedgerName, owner: $currentUser.id });
		newLedgerName = '';
		loadLedgers();
		window.create_ledger_modal.close();
	}

	async function addTransaction() {
		if (!selectedLedger || !amount) return;
		await pb.collection('transactions').create({
			ledger: selectedLedger,
			payer: $currentUser.id,
			amount: parseFloat(amount),
			split_type: splitType,
			beneficiary: splitType === 'SINGLE' ? beneficiary : null,
			note,
			date: new Date()
		});
		// 重置
		amount = '';
		note = '';
		window.add_modal.close();
	}
</script>

<div class="space-y-4">
	<div class="flex items-end justify-between px-1">
		<div>
			<h1 class="text-2xl font-black">我的账本</h1>
			<p class="text-xs opacity-50">管理并邀请好友加入</p>
		</div>
		<button
			class="btn btn-sm btn-circle btn-primary"
			on:click={() => window.create_ledger_modal.showModal()}
			title="新建账本"
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
					stroke-width="3"
					d="M12 4v16m8-8H4"
				/></svg
			>
		</button>
	</div>

	<div class="grid gap-3">
		{#each ledgers as ledger}
			<div
				class="card bg-base-100 border-base-300 border shadow-sm transition-transform active:scale-95"
			>
				<div class="card-body flex-row items-center justify-between p-4">
					<div>
						<h3 class="font-bold">{ledger.name}</h3>
						<p class="text-[10px] tracking-widest uppercase opacity-40">{ledger.id}</p>
					</div>
					<button class="btn btn-ghost btn-sm">详情</button>
				</div>
			</div>
		{:else}
			<div class="text-center py-20 opacity-30">
				<p>还没有账本，点击右上角创建一个</p>
			</div>
		{/each}
	</div>
</div>

<dialog id="create_ledger_modal" class="modal modal-bottom sm:modal-middle">
	<div class="modal-box">
		<h3 class="text-lg font-bold">新建账本</h3>
		<div class="py-4">
			<input
				type="text"
				bind:value={newLedgerName}
				placeholder="账本名称，如：日本行、宿舍公费"
				class="input input-bordered w-full"
			/>
		</div>
		<div class="modal-action">
			<button class="btn btn-primary btn-block" on:click={createLedger}>确认创建</button>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop"><button>关闭</button></form>
</dialog>

<dialog id="add_modal" class="modal modal-bottom sm:modal-middle">
	<div class="modal-box">
		<h3 class="mb-6 text-lg font-bold">新增支出记录</h3>
		<div class="space-y-4">
			<select class="select select-bordered w-full" bind:value={selectedLedger}>
				<option value="" disabled>选择账本</option>
				{#each ledgers as l}<option value={l.id}>{l.name}</option>{/each}
			</select>

			<div class="join w-full">
				<button class="btn join-item no-animation bg-base-200 border-base-300">¥</button>
				<input
					type="number"
					bind:value={amount}
					placeholder="金额"
					class="input input-bordered join-item w-full"
				/>
			</div>

			<div class="tabs tabs-boxed bg-base-200">
				<button
					class="tab flex-1 {splitType === 'AA' ? 'tab-active' : ''}"
					on:click={() => (splitType = 'AA')}>AA 均摊</button
				>
				<button
					class="tab flex-1 {splitType === 'SINGLE' ? 'tab-active' : ''}"
					on:click={() => (splitType = 'SINGLE')}>单人承担</button
				>
			</div>

			{#if splitType === 'SINGLE'}
				<div class="bg-base-200 border-primary/10 rounded-xl border p-3">
					<label class="label pt-0"
						><span class="label-text-alt font-bold">由谁支付/承担？</span></label
					>
					<select class="select select-sm select-ghost w-full" bind:value={beneficiary}>
						{#each members as m}
							<option value={m.id}
								>{m.username || m.email} {m.id === $currentUser.id ? '(自己)' : ''}</option
							>
						{/each}
					</select>
				</div>
			{/if}

			<input
				type="text"
				bind:value={note}
				placeholder="写点什么备注..."
				class="input input-bordered w-full"
			/>
		</div>
		<div class="modal-action">
			<button class="btn btn-primary btn-block shadow-lg" on:click={addTransaction}>保存账单</button
			>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop"><button>关闭</button></form>
</dialog>

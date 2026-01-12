<script lang="ts">
	import { pb, currentUser } from '$lib/pb';
	import { toasts } from '$lib/toast';
	import { onMount } from 'svelte';

	// Props
	export let ledgerId: string;
	export let ledgerName: string;

	// 状态管理
	let isLoading = false;
	let generatedCode = '';
	let expiresAt = '';
	let maxUses = 99;
	let usedCount = 0;
	let invitationId = '';
	let shouldCheckInvitation = false;

	// 监听 shouldCheckInvitation 变化
	$: if (shouldCheckInvitation) {
		checkExistingInvitation();
		shouldCheckInvitation = false;
	}

	// 关闭模态框
	function closeModal() {
		reset();
		const modal = document.getElementById(`generate_invite_${ledgerId}`) as HTMLDialogElement;
		if (modal) modal.close();
	}

	// 打开模态框时检查是否已存在有效邀请码
	async function checkExistingInvitation() {
		if (!ledgerId) return;

		isLoading = true;

		try {
			const response = await pb.send(`/api/invitations/by-ledger/${ledgerId}`, {
				method: 'GET'
			});

			const result = response as any;
			if (result.exists) {
				// 已存在有效邀请码，显示邀请码信息
				generatedCode = result.code;
				expiresAt = result.expiresAt;
				maxUses = result.maxUses;
				usedCount = result.usedCount;
				invitationId = result.id;
			} else {
				// 不存在有效邀请码，重置为初始状态
				generatedCode = '';
				expiresAt = '';
				maxUses = 99;
				usedCount = 0;
				invitationId = '';
			}
		} catch (err: any) {
			console.error('检查邀请码失败:', err);
			// 失败时显示生成表单
			generatedCode = '';
			expiresAt = '';
			maxUses = 99;
			usedCount = 0;
			invitationId = '';
		} finally {
			isLoading = false;
		}
	}

	// 生成邀请码
	async function generateInvitation() {
		if (!ledgerId) return;

		isLoading = true;

		try {
			const response = await pb.send('/api/invitations/generate', {
				method: 'POST',
				body: {
					ledger_id: ledgerId,
					max_uses: maxUses
				}
			});

			const result = response as any;
			generatedCode = result.code;
			expiresAt = result.expires_at;
			maxUses = result.max_uses;
			usedCount = result.used_count;
			toasts.success(`邀请码生成成功，可使用 ${maxUses} 次`);
		} catch (err: any) {
			console.error('生成邀请码失败:', err);
			// 如果已存在有效邀请码，显示该邀请码
			if (err.data?.code === 409 && err.data?.data) {
				const data = err.data.data;
				generatedCode = data.code;
				expiresAt = data.expiresAt;
				maxUses = data.maxUses;
				usedCount = data.usedCount;
				invitationId = data.id;
				toasts.info('该账本已存在有效邀请码');
			} else {
				toasts.error(err.data?.message || '生成邀请码失败');
			}
		} finally {
			isLoading = false;
		}
	}

	// 删除邀请码
	async function deleteInvitation() {
		if (!invitationId) return;

		isLoading = true;

		try {
			await pb.send(`/api/invitations/${invitationId}`, {
				method: 'DELETE'
			});

			toasts.success('邀请码已删除');
			// 重置为初始状态
			generatedCode = '';
			expiresAt = '';
			maxUses = 99;
			usedCount = 0;
			invitationId = '';
		} catch (err: any) {
			console.error('删除邀请码失败:', err);
			toasts.error(err.data?.message || '删除邀请码失败');
		} finally {
			isLoading = false;
		}
	}

	// 复制邀请码
	function copyCode() {
		if (!generatedCode) return;

		navigator.clipboard
			.writeText(generatedCode)
			.then(() => {
				toasts.success('邀请码已复制到剪贴板');
			})
			.catch(() => {
				toasts.error('复制失败');
			});
	}

	// 刷新邀请码状态
	async function refreshInvitationStatus() {
		if (!invitationId) return;

		try {
			const response = await pb.send(`/api/invitations/by-ledger/${ledgerId}`, {
				method: 'GET'
			});

			const result = response as any;
			if (result.exists) {
				usedCount = result.usedCount;
			} else {
				// 邀请码已失效
				generatedCode = '';
				expiresAt = '';
				maxUses = 99;
				usedCount = 0;
				invitationId = '';
				toasts.info('邀请码已过期或已用完');
			}
		} catch (err) {
			console.error('刷新邀请码状态失败:', err);
		}
	}

	// 关闭并重置
	function closeAndReset() {
		generatedCode = '';
		expiresAt = '';
		maxUses = 99;
		usedCount = 0;
		invitationId = '';
		closeModal();
	}

	// 重置
	function reset() {
		generatedCode = '';
		expiresAt = '';
		maxUses = 99;
		usedCount = 0;
		invitationId = '';
		isLoading = false;
	}
	function handleDialogClick(e) {
		const dialog = e.currentTarget as HTMLDialogElement;
		if (e.target === dialog) {
			closeAndReset();
		}
	}
</script>

<dialog
	id="generate_invite_{ledgerId}"
	class="modal modal-bottom sm:modal-middle"
	onclose={closeAndReset}
	onclick={handleDialogClick}
	onshow={() => (shouldCheckInvitation = true)}
>
	<div class="modal-box border">
		<div class="mb-4 flex items-center justify-between">
			<h3 class="text-lg font-bold">邀请成员加入「{ledgerName}」</h3>
			<button class="btn btn-ghost btn-sm" onclick={closeAndReset}>
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
		</div>

		{#if isLoading && !generatedCode}
			<div class="py-8 flex justify-center">
				<span class="loading loading-spinner loading-lg"></span>
			</div>
		{:else if generatedCode}
			<div class="space-y-4">
				<div class="text-center">
					<div class="bg-success/10 border-success/20 rounded-lg p-4 border">
						<h4 class="font-bold text-success mb-2">邀请码</h4>
						<div class="font-mono text-2xl font-bold tracking-wider bg-base-100 rounded p-3 border">
							{generatedCode}
						</div>
					</div>
				</div>

				<div class="bg-base-200 rounded-lg p-3">
					<div class="text-sm mb-2 flex items-center justify-between">
						<span class="text-base-content/60">有效期至：</span>
						<span class="font-medium text-warning">{expiresAt}</span>
					</div>
					<div class="text-sm flex items-center justify-between">
						<span class="text-base-content/60">已使用次数：</span>
						<span class="font-medium text-primary">{usedCount} / {maxUses} 次</span>
					</div>
				</div>

				<div class="bg-info/10 border-info/20 rounded-lg p-3 border">
					<p class="text-sm text-info">
						<strong>分享给朋友：</strong>让他们在头像菜单中选择"加入账本"，输入此邀请码即可加入。
					</p>
				</div>

				<div class="modal-action">
					<button class="btn btn-error btn-outline" onclick={deleteInvitation} disabled={isLoading}>
						{#if isLoading}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						删除邀请码
					</button>
					<button class="btn btn-ghost" onclick={refreshInvitationStatus} disabled={isLoading}>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-4 w-4 mr-1"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
							/>
						</svg>
						刷新
					</button>
					<button class="btn btn-primary" onclick={copyCode} disabled={isLoading}>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-4 w-4 mr-2"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
							/>
						</svg>
						复制邀请码
					</button>
				</div>
			</div>
		{:else}
			<div class="space-y-4">
				<p class="text-base-content/70">生成一个邀请码分享给朋友，他们可以使用此邀请码加入账本。</p>

				<div class="bg-base-200 rounded-lg p-4">
					<h4 class="font-semibold mb-2">邀请码设置：</h4>

					<!-- 使用次数设置 -->
					<div class="mb-3">
						<label for="max-uses" class="label">
							<span class="label-text font-medium">最大使用次数 (1-99)</span>
						</label>
						<input
							id="max-uses"
							type="number"
							min="1"
							max="99"
							bind:value={maxUses}
							class="input input-bordered w-24"
							oninput={() => {
								if (maxUses < 1) maxUses = 1;
								if (maxUses > 99) maxUses = 99;
							}}
						/>
					</div>

					<ul class="text-sm text-base-content/60 space-y-1">
						<li>• 邀请码有效期 24 小时</li>
						<li>• 邀请码可使用 {maxUses} 次</li>
						<li>• 邀请码格式：ABC-123456</li>
					</ul>
				</div>

				<div class="modal-action">
					<button class="btn btn-ghost" onclick={closeAndReset} disabled={isLoading}>取消</button>
					<button class="btn btn-primary" onclick={generateInvitation} disabled={isLoading}>
						{#if isLoading}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						生成邀请码
					</button>
				</div>
			</div>
		{/if}
	</div>
	<form method="dialog" class="modal-backdrop">
		<button onclick={closeAndReset}>关闭</button>
	</form>
</dialog>

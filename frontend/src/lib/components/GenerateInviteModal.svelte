<script lang="ts">
	import { pb, currentUser } from '$lib/pb';
	import { toasts } from '$lib/toast';

	// Props
	export let ledgerId: string;
	export let ledgerName: string;

	// 状态管理
	let isLoading = false;
	let generatedCode = '';
	let expiresAt = '';

	// 关闭模态框
	function closeModal() {
		reset();
		const modal = document.getElementById(`generate_invite_${ledgerId}`) as HTMLDialogElement;
		if (modal) modal.close();
	}

	// 生成邀请码
	async function generateInvitation() {
		if (!ledgerId) return;

		isLoading = true;

		try {
			const response = await pb.send('/api/invitations/generate', {
				method: 'POST',
				body: { ledger_id: ledgerId }
			});

			const result = response as any;
			generatedCode = result.code;
			expiresAt = result.expires_at;
			toasts.success('邀请码生成成功');
		} catch (err: any) {
			console.error('生成邀请码失败:', err);
			toasts.error(err.data?.message || '生成邀请码失败');
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

	// 关闭并重置
	function closeAndReset() {
		generatedCode = '';
		expiresAt = '';
		closeModal();
	}
</script>

<dialog id="generate_invite_{ledgerId}" class="modal modal-bottom sm:modal-middle">
	<div class="modal-box border">
		<div class="mb-4 flex items-center justify-between">
			<h3 class="text-lg font-bold">邀请成员加入「{ledgerName}」</h3>
			<button class="btn btn-ghost btn-sm" on:click={closeAndReset}>
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

		{#if !generatedCode}
			<div class="space-y-4">
				<p class="text-base-content/70">生成一个邀请码分享给朋友，他们可以使用此邀请码加入账本。</p>

				<div class="bg-base-200 rounded-lg p-4">
					<h4 class="font-semibold mb-2">邀请码说明：</h4>
					<ul class="text-sm text-base-content/60 space-y-1">
						<li>• 邀请码有效期 24 小时</li>
						<li>• 每个邀请码只能使用一次</li>
						<li>• 邀请码格式：ABC-123456</li>
					</ul>
				</div>

				<div class="modal-action">
					<button class="btn btn-ghost" on:click={closeAndReset} disabled={isLoading}>取消</button>
					<button class="btn btn-primary" on:click={generateInvitation} disabled={isLoading}>
						{#if isLoading}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						生成邀请码
					</button>
				</div>
			</div>
		{:else}
			<div class="space-y-4">
				<div class="text-center">
					<div class="bg-success/10 border-success/20 rounded-lg p-4 border">
						<h4 class="font-bold text-success mb-2">邀请码已生成</h4>
						<div class="font-mono text-2xl font-bold tracking-wider bg-base-100 rounded p-3 border">
							{generatedCode}
						</div>
					</div>
				</div>

				<div class="bg-base-200 rounded-lg p-3">
					<div class="text-sm flex items-center justify-between">
						<span class="text-base-content/60">有效期至：</span>
						<span class="font-medium text-warning">{expiresAt}</span>
					</div>
				</div>

				<div class="bg-info/10 border-info/20 rounded-lg p-3 border">
					<p class="text-sm text-info">
						<strong>分享给朋友：</strong>让他们在头像菜单中选择"加入账本"，输入此邀请码即可加入。
					</p>
				</div>

				<div class="modal-action">
					<button class="btn btn-ghost" on:click={closeAndReset}>关闭</button>
					<button class="btn btn-primary" on:click={copyCode}>
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
		{/if}
	</div>
	<form method="dialog" class="modal-backdrop">
		<button on:click={closeAndReset}>关闭</button>
	</form>
</dialog>

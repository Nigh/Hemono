<script lang="ts">
	import { pb, currentUser } from '$lib/pb';
	import { toasts } from '$lib/toast';

	// 状态管理
	let inputCode = '';
	let ledgerInfo: {
		ledgerId: string;
		ledgerName: string;
		createdBy: string;
		expiresAt: string;
	} | null = null;
	let isLoading = false;
	let step: 'input' | 'confirm' = 'input';
	let error = '';

	// 关闭模态框
	function closeModal() {
		reset();
		const modal = document.getElementById('join_ledger_modal') as HTMLDialogElement;
		if (modal) modal.close();
	}

	// 重置状态
	function reset() {
		inputCode = '';
		ledgerInfo = null;
		isLoading = false;
		step = 'input';
		error = '';
	}

	// 验证邀请码格式
	function isValidFormat(code: string): boolean {
		const formatRegex = /^[A-Z]{3}-\d{6}$/;
		return formatRegex.test(code);
	}

	// 自动格式化邀请码
	function formatCode(code: string): string {
		// 移除所有非字母数字字符
		const cleaned = code.replace(/[^A-Za-z0-9]/g, '').toUpperCase();

		// 限制长度
		const limited = cleaned.slice(0, 9); // 最多9个字符（3字母+6数字）

		// 如果长度超过3，插入连字符
		if (limited.length > 3) {
			return limited.slice(0, 3) + '-' + limited.slice(3);
		}

		return limited;
	}

	// 处理输入变化
	function handleInputChange(event: Event) {
		const target = event.target as HTMLInputElement;
		const formatted = formatCode(target.value);
		target.value = formatted;
		inputCode = formatted;
		error = '';
	}

	// 验证并预览邀请码
	async function validateAndPreview() {
		if (!inputCode) {
			error = '请输入邀请码';
			return;
		}

		if (!isValidFormat(inputCode)) {
			error = '邀请码格式不正确（格式：ABC-123456）';
			return;
		}

		isLoading = true;
		error = '';

		try {
			const response = await pb.send(`/api/invitations/by-code/${inputCode}`, {
				method: 'GET'
			});

			ledgerInfo = response as any;
			step = 'confirm';
		} catch (err: any) {
			console.error('验证邀请码失败:', err);
			if (err.status === 404) {
				error = '邀请码不存在，请检查是否输入正确';
			} else if (err.status === 410) {
				error = '邀请码已过期，请联系邀请方重新生成';
			} else if (err.status === 409) {
				error = '邀请码已被使用';
			} else {
				error = err.data?.message || '验证邀请码失败';
			}
		} finally {
			isLoading = false;
		}
	}

	// 确认加入账本
	async function confirmJoin() {
		if (!ledgerInfo) return;

		isLoading = true;
		error = '';

		try {
			const response = await pb.send('/api/invitations/join', {
				method: 'POST',
				body: { code: inputCode }
			});

			const result = response as any;
			toasts.success(`成功加入账本「${result.ledgerName}」`);

			// 延迟关闭，让用户看到成功消息
			setTimeout(() => {
				closeModal();
				// 刷新页面或触发重新加载账本列表
				window.location.reload();
			}, 1000);
		} catch (err: any) {
			console.error('加入账本失败:', err);
			if (err.status === 409) {
				error = err.data?.message || '你已经是该账本的成员';
			} else if (err.status === 410) {
				error = '邀请码已过期，请联系邀请方重新生成';
			} else {
				error = err.data?.message || '加入账本失败';
			}
		} finally {
			isLoading = false;
		}
	}

	// 返回上一步
	function goBack() {
		step = 'input';
		error = '';
		ledgerInfo = null;
	}
</script>

<dialog id="join_ledger_modal" class="modal modal-bottom sm:modal-middle">
	<div class="modal-box border">
		<div class="mb-4 flex items-center justify-between">
			<h3 class="text-lg font-bold">
				{#if step === 'input'}
					加入账本
				{:else}
					确认加入
				{/if}
			</h3>
			<button class="btn btn-ghost btn-sm" on:click={closeModal} title="关闭模态框">
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

		{#if error}
			<div role="alert" class="alert alert-error mb-4">
				<span>{error}</span>
			</div>
		{/if}

		{#if step === 'input'}
			<div class="space-y-4">
				<div>
					<label class="label">
						<span class="label-text font-medium">请输入邀请码</span>
					</label>
					<input
						type="text"
						placeholder="ABC-123456"
						class="input input-bordered font-mono text-lg tracking-wider w-full text-center"
						class:input-error={error && !isValidFormat(inputCode)}
						value={inputCode}
						on:input={handleInputChange}
						on:keydown={(e) => e.key === 'Enter' && validateAndPreview()}
						disabled={isLoading}
						maxlength="10"
					/>
					<label class="label">
						<span class="label-text-alt text-base-content/60">
							朋友会分享给你一个形如 ABC-123456 的邀请码
						</span>
					</label>
				</div>

				{#if inputCode && !isValidFormat(inputCode)}
					<div class="text-xs text-error">
						邀请码格式不正确，应为 3 位大写字母 + 连字符 + 6 位数字
					</div>
				{/if}

				<div class="modal-action">
					<button class="btn btn-ghost" on:click={closeModal} disabled={isLoading}>取消</button>
					<button
						class="btn btn-primary"
						on:click={validateAndPreview}
						disabled={isLoading || !isValidFormat(inputCode)}
					>
						{#if isLoading}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						加入
					</button>
				</div>
			</div>
		{:else if step === 'confirm' && ledgerInfo}
			<div class="space-y-4">
				<div class="bg-base-200 rounded-xl p-4">
					<h4 class="font-bold text-lg mb-2">邀请你加入账本</h4>
					<div class="space-y-2">
						<div class="flex justify-between">
							<span class="text-base-content/60">账本名称：</span>
							<span class="font-medium">{ledgerInfo.ledgerName}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-base-content/60">邀请方：</span>
							<span class="font-medium">{ledgerInfo.createdBy}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-base-content/60">有效期至：</span>
							<span class="font-medium text-warning">{ledgerInfo.expiresAt}</span>
						</div>
					</div>
				</div>

				<div class="modal-action">
					<button class="btn btn-ghost" on:click={goBack} disabled={isLoading}>返回</button>
					<button class="btn btn-primary" on:click={confirmJoin} disabled={isLoading}>
						{#if isLoading}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						确认加入
					</button>
				</div>
			</div>
		{/if}
	</div>
	<form method="dialog" class="modal-backdrop">
		<button on:click={closeModal}>关闭</button>
	</form>
</dialog>

<script lang="ts">
	import {
		fetchTokens,
		createToken,
		deleteToken,
		type ApiToken,
		type CreatedToken
	} from '$lib/api/tokens';
	import { toasts } from '$lib/toast';

	let tokens: ApiToken[] = $state([]);
	let isLoading = $state(true);
	let isCreating = $state(false);
	let newTokenName = $state('');
	let createdToken: CreatedToken | null = $state(null);
	let showCreateDialog = $state(false);
	let showTokenDialog = $state(false);
	let deleteTarget: ApiToken | null = $state(null);
	let showDeleteDialog = $state(false);

	let createDialog: HTMLDialogElement;
	let tokenDialog: HTMLDialogElement;
	let deleteDialog: HTMLDialogElement;

	async function loadTokens() {
		isLoading = true;
		try {
			tokens = await fetchTokens();
		} catch (err: any) {
			toasts.error('加载 Token 列表失败');
		} finally {
			isLoading = false;
		}
	}

	function openCreateDialog() {
		newTokenName = '';
		createDialog.showModal();
	}

	async function handleCreate() {
		if (!newTokenName.trim()) {
			toasts.warning('请输入 Token 名称');
			return;
		}
		isCreating = true;
		try {
			createdToken = await createToken(newTokenName.trim());
			createDialog.close();
			tokenDialog.showModal();
			await loadTokens();
		} catch (err: any) {
			if (err.status === 409) {
				toasts.error('最多创建 3 个 Token');
			} else {
				toasts.error(err.data?.message || '创建 Token 失败');
			}
		} finally {
			isCreating = false;
		}
	}

	function confirmDelete(token: ApiToken) {
		deleteTarget = token;
		showDeleteDialog = true;
		deleteDialog.showModal();
	}

	async function handleDelete() {
		if (!deleteTarget) return;
		try {
			await deleteToken(deleteTarget.id);
			toasts.success('Token 已删除');
			deleteDialog.close();
			await loadTokens();
		} catch (err: any) {
			toasts.error('删除 Token 失败');
		} finally {
			deleteTarget = null;
		}
	}

	async function copyToken() {
		if (!createdToken) return;
		try {
			await navigator.clipboard.writeText(createdToken.token);
			toasts.success('已复制到剪贴板');
		} catch {
			toasts.error('复制失败，请手动复制');
		}
	}

	$effect(() => {
		loadTokens();
	});
</script>

<svelte:head>
	<title>API Token 管理 - 荷物账本</title>
</svelte:head>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<h1 class="text-xl font-bold">API Token 管理</h1>
		<button class="btn btn-primary btn-sm" onclick={openCreateDialog} disabled={tokens.length >= 3}>
			创建 Token
		</button>
	</div>

	<p class="text-base-content/60 text-sm">
		通过 API Token 访问荷物账本的 REST API。每个账户最多创建 3 个 Token。
	</p>

	{#if isLoading}
		<div class="py-8 flex justify-center">
			<span class="loading loading-spinner loading-lg"></span>
		</div>
	{:else if tokens.length === 0}
		<div class="bg-base-200 rounded-xl p-8 text-center">
			<p class="text-base-content/60">暂无 Token</p>
			<p class="text-base-content/40 mt-1 text-sm">点击上方按钮创建你的第一个 API Token</p>
		</div>
	{:else}
		<div class="space-y-2">
			{#each tokens as token (token.id)}
				<div
					class="bg-base-100 border-base-300 rounded-xl p-4 flex items-center justify-between border"
				>
					<div class="min-w-0 flex-1">
						<div class="font-medium">{token.name}</div>
						<div class="text-base-content/60 font-mono text-sm">
							{token.token_prefix}...
						</div>
						<div class="text-base-content/40 text-xs">
							创建于 {new Date(token.created).toLocaleDateString('zh-CN')}
						</div>
					</div>
					<button class="btn btn-ghost btn-sm text-error" onclick={() => confirmDelete(token)}>
						删除
					</button>
				</div>
			{/each}
		</div>
	{/if}
</div>

<dialog
	id="create_token_dialog"
	class="modal modal-bottom sm:modal-middle"
	bind:this={createDialog}
>
	<div class="modal-box border">
		<h3 class="text-lg font-bold">创建 Token</h3>
		<div class="mt-4 space-y-4">
			<div>
				<label class="label" for="token-name">
					<span class="label-text">Token 名称</span>
				</label>
				<input
					id="token-name"
					type="text"
					class="input input-bordered w-full"
					placeholder="例如：我的 CLI 工具"
					bind:value={newTokenName}
					onkeydown={(e) => e.key === 'Enter' && handleCreate()}
					disabled={isCreating}
					maxlength={100}
				/>
			</div>
		</div>
		<div class="modal-action">
			<button class="btn btn-ghost" onclick={() => createDialog.close()} disabled={isCreating}>
				取消
			</button>
			<button
				class="btn btn-primary"
				onclick={handleCreate}
				disabled={isCreating || !newTokenName.trim()}
			>
				{#if isCreating}
					<span class="loading loading-spinner loading-sm"></span>
				{/if}
				创建
			</button>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>关闭</button>
	</form>
</dialog>

<dialog
	id="token_display_dialog"
	class="modal modal-bottom sm:modal-middle"
	bind:this={tokenDialog}
>
	<div class="modal-box border">
		<h3 class="text-lg font-bold text-success">Token 创建成功</h3>
		<div class="mt-4 space-y-4">
			<div role="alert" class="alert alert-warning">
				<span>请立即复制 Token，关闭后将无法再次查看。</span>
			</div>
			{#if createdToken}
				<div class="bg-base-200 rounded-xl p-4">
					<div class="text-base-content/60 mb-1 text-sm">名称</div>
					<div class="font-medium">{createdToken.name}</div>
				</div>
				<div class="bg-base-200 rounded-xl p-4">
					<div class="text-base-content/60 mb-1 text-sm">Token</div>
					<div class="bg-base-300 rounded-lg p-3 font-mono text-sm break-all">
						{createdToken.token}
					</div>
				</div>
			{/if}
		</div>
		<div class="modal-action">
			<button class="btn btn-primary" onclick={copyToken}> 复制 Token </button>
			<button class="btn btn-ghost" onclick={() => tokenDialog.close()}> 关闭 </button>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>关闭</button>
	</form>
</dialog>

<dialog
	id="delete_token_dialog"
	class="modal modal-bottom sm:modal-middle"
	bind:this={deleteDialog}
>
	<div class="modal-box border">
		<h3 class="text-lg font-bold text-error">确认删除</h3>
		<div class="mt-4">
			{#if deleteTarget}
				<p>确定要删除 Token <span class="font-mono font-bold">「{deleteTarget.name}」</span>吗？</p>
				<p class="text-base-content/60 mt-2 text-sm">
					此操作不可撤销，使用该 Token 的应用将立即失去访问权限。
				</p>
			{/if}
		</div>
		<div class="modal-action">
			<button class="btn btn-ghost" onclick={() => deleteDialog.close()}>取消</button>
			<button class="btn btn-error" onclick={handleDelete}>删除</button>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>关闭</button>
	</form>
</dialog>

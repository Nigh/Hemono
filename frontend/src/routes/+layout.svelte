<script lang="ts">
	import '../app.css';
	import { pb, currentUser } from '$lib/pb';
	import { onMount, type Snippet } from 'svelte';
	import { restoreAuth } from '$lib/api/auth';
	import { resolve } from '$app/paths';
	import Toast from '$lib/components/Toast.svelte';
	import JoinLedgerModal from '$lib/components/JoinLedgerModal.svelte';
	import { toasts } from '$lib/toast';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();
	let authReady = $state(false);
	let authError = $state(false);
	let restoring: Promise<void> | undefined;

	function checkSession() {
		if (restoring) return restoring;
		if (!pb.authStore.isValid) authReady = false;
		authError = false;
		restoring = restoreAuth()
			.catch(() => {
				authError = !authReady;
			})
			.finally(() => {
				authReady = true;
				restoring = undefined;
			});
		return restoring;
	}

	onMount(() => {
		checkSession();
	});

	// 退出登录
	function logout() {
		pb.cancelAllRequests();
		pb.authStore.clear();
		toasts.success('已退出登录');
	}

	// 登录逻辑
	async function loginWithGitHub() {
		try {
			await pb.collection('users').authWithOAuth2({ provider: 'github' });
			toasts.success('登录成功');
		} catch (err) {
			console.error('Login failed:', err);
			toasts.error(`登录失败: ${err instanceof Error ? err.message : '未知错误'}`);
		}
	}

	// 打开加入账本模态框
	function openJoinLedgerModal() {
		const modal = document.getElementById('join_ledger_modal') as HTMLDialogElement;
		if (modal) modal.showModal();
	}
</script>

<svelte:window onfocus={checkSession} />

<div class="bg-base-200 min-h-screen">
	<Toast />
	<header class="navbar bg-base-100 border-base-300 top-0 px-4 sm:px-6 sticky z-50 border-b">
		<div class="flex-1">
			<a href={resolve('/')} class="text-primary text-xl font-bold">荷物账本</a>
		</div>
		<div class="px-2 flex-none select-none">
			{#if authReady && !authError && $currentUser}
				<div class="dropdown dropdown-end">
					<button type="button" aria-label="打开账户菜单" class="avatar btn btn-ghost btn-circle">
						<div
							class="hover:bg-primary ring-primary w-10 h-10 border-primary cursor-pointer rounded-full border-2 hover:ring-2"
						>
							<img
								src={$currentUser.avatar
									? pb.files.getURL($currentUser, $currentUser.avatar)
									: `https://api.dicebear.com/7.x/bottts/svg?seed=${$currentUser.id}`}
								alt="avatar"
							/>
						</div>
					</button>
					<ul
						tabindex="-1"
						class="menu menu-xs w-48 bg-base-100 dropdown-content border-base-300 rounded-box mt-3 p-2 shadow z-1 border"
					>
						<li class="menu-title mb-2 border-b">
							<span class="tooltip" data-tip={$currentUser.email}>{$currentUser.name}</span>
						</li>
						<li><button onclick={openJoinLedgerModal}>加入账本</button></li>
						<li><button onclick={logout} class="text-error">退出登录</button></li>
					</ul>
				</div>
			{:else if authReady && !authError}
				<button class="btn btn-sm btn-primary" onclick={loginWithGitHub}>GitHub 登录</button>
			{/if}
		</div>
	</header>

	<main class="max-w-2xl px-4 py-8 sm:px-6 sm:py-12 mx-auto w-full">
		{#if !authReady}
			<div class="py-20 flex justify-center" role="status" aria-label="正在验证登录状态">
				<span class="loading loading-spinner loading-md"></span>
			</div>
		{:else if authError}
			<div class="alert alert-error" role="alert">
				<span>无法验证登录状态，请检查网络后重试。</span>
				<button class="btn btn-sm" onclick={checkSession}>重试</button>
			</div>
		{:else if $currentUser}
			{@render children()}
		{:else}
			<div class="flex min-h-[65vh] flex-col items-center justify-center text-center">
				<p class="text-primary mb-4 text-sm font-medium">一起记录，轻松分担</p>
				<h1 class="mb-4 text-3xl font-bold leading-snug sm:text-4xl">欢迎使用荷物账本</h1>
				<p class="text-text-muted mb-8 leading-relaxed">记录每一笔，让分享更容易</p>
				<button class="btn btn-primary" onclick={loginWithGitHub}> 使用 GitHub 快速开始 </button>
			</div>
		{/if}
	</main>

	{#if authReady && !authError && $currentUser}
		<JoinLedgerModal />
	{/if}
</div>

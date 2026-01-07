<script lang="ts">
	import '../app.css';
	import { pb, currentUser } from '$lib/pb';
	import { onMount } from 'svelte';

	// 退出登录
	function logout() {
		pb.authStore.clear();
	}

	// 登录逻辑
	async function loginWithGitHub() {
		try {
			await pb.collection('users').authWithOAuth2({ provider: 'github' });
		} catch (err) {
			console.error('Login failed:', err);
		}
	}
</script>

<div class="bg-base-200 min-h-screen pb-20">
	<div class="navbar bg-base-100 sticky top-0 z-50 shadow-sm">
		<div class="flex-1">
			<span class="text-primary px-4 text-xl font-black tracking-tighter">荷物账本</span>
		</div>
		<div class="flex-none px-2">
			{#if $currentUser}
				<div class="dropdown dropdown-end">
					<div
						tabindex="0"
						role="button"
						class="btn btn-ghost btn-circle avatar border-primary/20 border"
					>
						<div class="w-10 rounded-full">
							<img
								src={$currentUser.avatar
									? pb.files.getURL($currentUser, $currentUser.avatar)
									: `https://api.dicebear.com/7.x/bottts/svg?seed=${$currentUser.id}`}
								alt="avatar"
							/>
						</div>
					</div>
					<ul
						tabindex="-1"
						class="menu menu-sm dropdown-content bg-base-100 rounded-box z-1 mt-3 w-52 p-2 shadow"
					>
						<li><button on:click={logout} class="text-error">退出登录</button></li>
					</ul>
				</div>
			{:else}
				<button class="btn btn-sm btn-primary" on:click={loginWithGitHub}>GitHub 登录</button>
			{/if}
		</div>
	</div>

	<main class="mx-auto max-w-md p-4">
		{#if $currentUser}
			<slot />
		{:else}
			<div class="flex h-[70vh] flex-col items-center justify-center text-center">
				<h2 class="mb-2 text-2xl font-bold">欢迎使用荷物账本</h2>
				<p class="text-base-content/60 mb-8">记录每一笔，让分享更容易</p>
				<button class="btn btn-primary btn-wide shadow-lg" on:click={loginWithGitHub}>
					使用 GitHub 快速开始
				</button>
			</div>
		{/if}
	</main>

	{#if $currentUser}
		<div class="dock border-base-300 z-50 border-t">
			<button class="active">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-5 w-5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					><path
						d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
					/></svg
				>
				<span class="btm-nav-label text-xs">首页</span>
			</button>
			<button on:click={() => window.add_modal.showModal()} title="添加账单">
				<div
					class="bg-primary text-primary-content border-base-200 -mt-10 rounded-full border-4 p-3 shadow-xl"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-6 w-6"
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
				</div>
			</button>
			<button>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-5 w-5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					><path
						d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
					/></svg
				>
				<span class="btm-nav-label text-xs">报表</span>
			</button>
		</div>
	{/if}
</div>

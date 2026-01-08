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

<div class="bg-base-200 min-h-screen">
	<div class="navbar bg-base-100 top-0 shadow-sm sticky z-50">
		<div class="flex-1">
			<span class="text-primary px-4 text-xl font-black tracking-tighter">荷物账本</span>
		</div>
		<div class="px-2 flex-none">
			{#if $currentUser}
				<div class="dropdown dropdown-end">
					<div
						tabindex="0"
						role="button"
						class="avatar"
					>
						<div class="cursor-pointer hover:bg-primary hover:ring-2 ring-primary w-10 h-10 border-primary border-2 rounded-full">
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
						class="menu menu-xs dropdown-content border border-base rounded-box mt-3 p-2 shadow z-1"
					>
						<li class="menu-title"><span>{$currentUser.username || $currentUser.email}</span></li>
						<li><button on:click={logout} class="text-error">退出登录</button></li>
					</ul>
				</div>
			{:else}
				<button class="btn btn-sm btn-primary" on:click={loginWithGitHub}>GitHub 登录</button>
			{/if}
		</div>
	</div>

	<main class="max-w-md p-4 mx-auto">
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
</div>

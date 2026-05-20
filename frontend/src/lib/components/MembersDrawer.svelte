<script lang="ts">
	import { pb } from '$lib/pb';

	interface Member {
		id: string;
		name?: string;
		email: string;
		avatar?: string;
	}

	interface Props {
		members: Member[];
		open: boolean;
		onclose: () => void;
	}

	let { members, open, onclose }: Props = $props();
</script>

<div class="inset-0 fixed z-50 {open ? 'pointer-events-auto' : 'pointer-events-none'}">
	<button
		class="inset-0 bg-black/40 fixed transition-opacity duration-300 {open
			? 'opacity-100'
			: 'opacity-0'}"
		onclick={onclose}
		aria-label="关闭"
	></button>
	<div
		class="top-0 right-0 w-72 bg-base-100 border-base-300 shadow-xl p-4 space-y-4 fixed h-full overflow-y-auto border-l transition-transform duration-300 {open
			? 'translate-x-0'
			: 'translate-x-full'}"
	>
		<div class="flex items-center justify-between">
			<h3 class="font-bold text-lg">成员列表</h3>
			<button class="btn btn-ghost btn-sm btn-circle" onclick={onclose} aria-label="关闭成员列表">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-5 w-5"
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

		<ul class="menu menu-sm w-full">
			{#each members as member}
				<li>
					<div class="gap-3 flex items-center">
						<div class="avatar">
							<div class="w-8 h-8 rounded-full">
								<img
									src={member.avatar
										? pb.files.getURL(member, member.avatar)
										: `https://api.dicebear.com/9.x/pixel-art/svg?seed=${member.id}`}
									alt="avatar"
								/>
							</div>
						</div>
						<div class="min-w-0 flex-1">
							<p class="font-medium truncate">{member.name || member.email}</p>
							<p class="text-xs truncate opacity-50">{member.email}</p>
						</div>
					</div>
				</li>
			{:else}
				<li class="text-center opacity-50 py-4">暂无成员</li>
			{/each}
		</ul>
	</div>
</div>

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
		ownerId?: string;
	}

	let { members, open, onclose, ownerId }: Props = $props();
	let dialog: HTMLDialogElement;

	$effect(() => {
		if (open) dialog.showModal();
		else dialog.close();
	});
</script>

<dialog bind:this={dialog} class="modal" aria-labelledby="members-title" {onclose}>
	<div
		class="right-0 top-0 w-72 space-y-4 border-base-300 bg-base-100 p-4 fixed h-full max-w-full overflow-y-auto border-l"
	>
		<div class="flex items-center justify-between">
			<h2 id="members-title" class="text-xl font-semibold">成员列表</h2>
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
							<div
								class="w-8 h-8 rounded-full {member.id === ownerId
									? 'ring-warning ring-offset-base-100 ring-2 ring-offset-1'
									: ''}"
							>
								<img
									src={member.avatar
										? pb.files.getURL(member, member.avatar)
										: `https://api.dicebear.com/9.x/pixel-art/svg?seed=${member.id}`}
									alt="avatar"
								/>
							</div>
							{#if member.id === ownerId}
								<div
									class="-top-0.5 -right-0.5 bg-base-100 shadow-sm absolute z-10 rounded-full p-px"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 24 24"
										fill="currentColor"
										class="w-3.5 h-3.5 text-warning drop-shadow"
									>
										<path d="M2.5 18.5l2-10 5 4 2.5-6 2.5 6 5-4 2 10z" />
									</svg>
								</div>
							{/if}
						</div>
						<div class="min-w-0 flex-1">
							<p class="font-medium truncate">{member.name || member.email}</p>
							<p class="text-xs text-text-muted truncate">{member.email}</p>
						</div>
					</div>
				</li>
			{:else}
				<li class="text-center text-text-muted py-4">暂无成员</li>
			{/each}
		</ul>
	</div>
	<form method="dialog" class="modal-backdrop"><button>关闭</button></form>
</dialog>

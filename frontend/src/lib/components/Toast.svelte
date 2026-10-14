<script lang="ts">
	import { toasts } from '../toast';
	import { flip } from 'svelte/animate';
	import { fade, fly } from 'svelte/transition';
</script>

<div class="toast toast-top toast-center z-100 max-w-full" aria-live="polite">
	{#each $toasts as toast (toast.id)}
		<div
			animate:flip={{ duration: 300 }}
			in:fly={{ y: -20, duration: 300 }}
			out:fade={{ duration: 200 }}
			class="alert {toast.type === 'success' ? 'alert-success' : ''} {toast.type === 'error'
				? 'alert-error'
				: ''} {toast.type === 'warning' ? 'alert-warning' : ''} {toast.type === 'info'
				? 'alert-info'
				: ''} "
		>
			<span>{toast.message}</span>
			<button
				class="btn btn-ghost btn-xs"
				onclick={() => toasts.remove(toast.id)}
				aria-label="关闭通知">✕</button
			>
		</div>
	{/each}
</div>

<script lang="ts">
	interface Props {
		ledger: {
			id: string;
			name: string;
		};
		isExpanded: boolean;
		ontoggle?: () => void;
	}

	let { ledger, isExpanded, ontoggle }: Props = $props();
</script>

<button
	class="card-body gap-4 rounded-box p-5 hover:bg-interactive-hover w-full flex-row items-center justify-between text-left"
	onclick={ontoggle}
	aria-expanded={isExpanded}
>
	<div class="min-w-0">
		<h2 class="text-xl font-semibold break-words">{ledger.name}</h2>
		<p class="mt-1 font-mono text-xs text-text-muted">{ledger.id}</p>
	</div>
	<div class="gap-2 flex items-center">
		<svg
			xmlns="http://www.w3.org/2000/svg"
			class="h-5 w-5 transition-transform duration-200 {isExpanded ? 'rotate-180' : ''}"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
		>
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
		</svg>
	</div>
</button>

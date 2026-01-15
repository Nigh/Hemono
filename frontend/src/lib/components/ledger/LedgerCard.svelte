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
	class="card-body p-4 w-full flex-row items-center justify-between text-left"
	onclick={ontoggle}
>
	<div>
		<h3 class="font-bold text-xl">{ledger.name}</h3>
		<p class="tracking-widest text-xs uppercase opacity-40">{ledger.id}</p>
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

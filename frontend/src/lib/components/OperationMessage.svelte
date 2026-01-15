<script lang="ts">
	import type { OperationMessageType } from '$lib/stores/operation-message';

	interface Props {
		type: OperationMessageType;
		message: string;
	}

	let { type, message }: Props = $props();

	function getAlertClass(type: OperationMessageType): string {
		if (type === 'success') return 'alert-success';
		if (type === 'error') return 'alert-error';
		return 'alert-warning';
	}
</script>

<div role="alert" class="alert mb-4 {getAlertClass(type)}">
	<span>{message}</span>
</div>

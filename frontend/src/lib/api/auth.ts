import { ClientResponseError } from 'pocketbase';
import { pb } from '$lib/pb';

export async function restoreAuth() {
	if (!pb.authStore.isValid) {
		pb.authStore.clear();
		return;
	}

	const token = pb.authStore.token;
	try {
		await pb.collection('users').authRefresh();
	} catch (error) {
		if (error instanceof ClientResponseError && [401, 403].includes(error.status)) {
			if (pb.authStore.token === token) pb.authStore.clear();
			return;
		}
		throw error;
	}
}

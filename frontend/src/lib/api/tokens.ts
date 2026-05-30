import { pb } from '$lib/pb';

export interface ApiToken {
	id: string;
	name: string;
	token_prefix: string;
	created: string;
}

export interface CreatedToken extends ApiToken {
	token: string;
}

export async function fetchTokens(): Promise<ApiToken[]> {
	return (await pb.send('/api/tokens', {
		method: 'GET',
		requestKey: null
	})) as ApiToken[];
}

export async function createToken(name: string): Promise<CreatedToken> {
	return (await pb.send('/api/tokens', {
		method: 'POST',
		body: { name },
		requestKey: null
	})) as CreatedToken;
}

export async function deleteToken(id: string): Promise<void> {
	await pb.send(`/api/tokens/${id}`, {
		method: 'DELETE',
		requestKey: null
	});
}

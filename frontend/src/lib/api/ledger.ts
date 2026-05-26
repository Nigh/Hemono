import { pb } from '$lib/pb';

export async function fetchLedgers() {
	return await pb.collection('ledgers').getFullList({ sort: '-created' });
}

export async function createLedger(name: string, ownerId: string) {
	return await pb.collection('ledgers').create({ name, owner: ownerId });
}

export async function fetchLedger(id: string) {
	return await pb.collection('ledgers').getOne(id);
}

export async function fetchLedgerMembers(ledgerId: string) {
	const res = await pb.collection('ledger_members').getFullList({
		filter: `ledger = "${ledgerId}"`,
		expand: 'user',
		requestKey: null
	});
	return res.filter((m) => m.expand?.user).map((m) => m.expand!.user);
}

export interface TransactionData {
	ledger: string;
	payer: string;
	amount: number;
	type: string;
	direction: string;
	beneficiary: string | null;
	note: string;
	date: Date;
}

export async function addTransaction(data: TransactionData) {
	return await pb.collection('transactions').create(data);
}

export async function fetchTransactions(ledgerId: string) {
	return await pb.collection('transactions').getFullList({
		filter: `ledger = "${ledgerId}"`,
		expand: 'payer',
		sort: '-date,-created'
	});
}

export async function deleteTransaction(id: string) {
	return await pb.collection('transactions').delete(id);
}

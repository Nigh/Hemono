import { pb } from '$lib/pb';

export async function fetchLedgers() {
	return await pb.collection('ledgers').getFullList({ sort: '-created' });
}

export async function createLedger(name: string, ownerId: string) {
	return await pb.collection('ledgers').create({ name, owner: ownerId });
}

export async function fetchLedgerMembers(ledgerId: string) {
	const res = await pb.collection('ledger_members').getFullList({
		filter: `ledger = "${ledgerId}"`,
		expand: 'user'
	});
	return res.map((m) => m.expand.user);
}

export interface TransactionData {
	ledger: string;
	payer: string;
	amount: number;
	type: string;
	beneficiary: string | null;
	note: string;
	date: Date;
}

export async function addTransaction(data: TransactionData) {
	return await pb.collection('transactions').create(data);
}

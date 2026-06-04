import { type RecordModel, type ListResult } from 'pocketbase';
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

export interface FetchTransactionsParams {
	ledgerId: string;
	page?: number;
	perPage?: number;
	keyword?: string;
}

export async function fetchTransactions(
	params: FetchTransactionsParams
): Promise<ListResult<RecordModel>>;
export async function fetchTransactions(ledgerId: string): Promise<RecordModel[]>;
export async function fetchTransactions(
	paramsOrLedgerId: FetchTransactionsParams | string
): Promise<RecordModel[] | ListResult<RecordModel>> {
	if (typeof paramsOrLedgerId === 'string') {
		return await pb.collection('transactions').getFullList({
			filter: `ledger = "${paramsOrLedgerId}"`,
			expand: 'payer',
			sort: '-date,-created'
		});
	}

	const { ledgerId, page = 1, perPage = 20, keyword = '' } = paramsOrLedgerId;
	const filters: string[] = [`ledger = "${ledgerId}"`];

	if (keyword.trim()) {
		const kw = keyword.trim();
		const escaped = kw.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
		const subFilters: string[] = [
			`note ~ "${escaped}"`,
			`payer.name ~ "${escaped}"`,
			`payer.email ~ "${escaped}"`
		];
		const parsed = parseFloat(kw);
		if (!isNaN(parsed) && parsed > 0) {
			const dotIndex = kw.indexOf('.');
			const decimals = dotIndex >= 0 ? kw.length - dotIndex - 1 : 0;
			const cents = Math.round(parsed * 100);

			if (decimals >= 2) {
				subFilters.push(`amount = ${cents}`);
			} else if (decimals === 1) {
				subFilters.push(`(amount >= ${cents} && amount < ${cents + 10})`);
			} else {
				const base = Math.round(parsed);
				const rangeFilters: string[] = [];
				let scale = 1;
				while (base * scale < 1e8) {
					const lo = base * scale * 100;
					const hi = (base + 1) * scale * 100;
					rangeFilters.push(`(amount >= ${lo} && amount < ${hi})`);
					scale *= 10;
				}
				if (rangeFilters.length > 0) {
					subFilters.push(`(${rangeFilters.join(' || ')})`);
				}
			}
		}
		filters.push(`(${subFilters.join(' || ')})`);
	}

	return await pb.collection('transactions').getList(page, perPage, {
		filter: filters.join(' && '),
		expand: 'payer',
		sort: '-date,-created'
	});
}

export async function deleteTransaction(id: string) {
	return await pb.collection('transactions').delete(id);
}

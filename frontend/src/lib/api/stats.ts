import { pb } from '$lib/pb';

export interface MemberStat {
	userId: string;
	name: string;
	avatar: string;
	totalExpense: number; // 分
	totalBenefit: number; // 分
	balance: number; // 分，支出 - 受益
	percentage: number; // 百分比
}

export interface LedgerStat {
	totalExpense: number; // 分
	totalBenefit: number; // 分
	memberStats: MemberStat[];
}

export async function fetchLedgerStats(ledgerId: string, month?: string): Promise<LedgerStat> {
	let url = `/api/ledgers/${ledgerId}/stats`;
	if (month) {
		url += `?month=${encodeURIComponent(month)}`;
	}

	const response = await pb.send(url, {
		method: 'GET'
	});

	return response as LedgerStat;
}

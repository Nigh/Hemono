import { pb } from '$lib/pb';

export interface MemberStat {
	userId: string;
	name: string;
	email: string;
	avatar: string;
	totalExpense: number; // 分
	totalBenefit: number; // 分
	totalIncome: number; // 分
	incomeShare: number; // 分
	balance: number; // 分，支出 - 收入 - 受益 + 收入分配
	percentage: number; // 百分比
}

export interface LedgerStat {
	totalExpense: number; // 分
	totalBenefit: number; // 分
	totalIncome: number; // 分
	totalIncomeShare: number; // 分
	monthlyExpense: number; // 分
	monthlyIncome: number; // 分
	last7DaysExpense: number; // 分
	last7DaysIncome: number; // 分
	memberStats: MemberStat[];
}

export async function fetchLedgerStats(ledgerId: string, month?: string): Promise<LedgerStat> {
	let url = `/api/ledgers/${ledgerId}/stats`;
	if (month) {
		url += `?month=${encodeURIComponent(month)}`;
	}

	const response = await pb.send(url, {
		method: 'GET',
		requestKey: null
	});

	return response as LedgerStat;
}

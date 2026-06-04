export interface TransactionLike {
	note?: string | null;
	amount: number;
	expand?: {
		payer?: {
			name?: string | null;
			email?: string | null;
		} | null;
	} | null;
}

export function filterTransactions<T extends TransactionLike>(
	transactions: T[],
	keyword: string
): T[] {
	const kw = keyword.trim().toLowerCase();
	if (!kw) return transactions;
	return transactions.filter((tx) => {
		const note = tx.note?.toLowerCase() ?? '';
		const payerName = tx.expand?.payer?.name?.toLowerCase() ?? '';
		const payerEmail = tx.expand?.payer?.email?.toLowerCase() ?? '';
		const amountStr = (tx.amount / 100).toFixed(2);
		return (
			note.includes(kw) ||
			payerName.includes(kw) ||
			payerEmail.includes(kw) ||
			amountStr.includes(kw)
		);
	});
}

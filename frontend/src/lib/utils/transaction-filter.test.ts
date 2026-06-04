import { describe, it, expect } from 'vitest';
import { filterTransactions, type TransactionLike } from './transaction-filter';

function tx(
	id: string,
	amount: number,
	note: string,
	payerName: string,
	payerEmail?: string
): TransactionLike & { id: string } {
	return {
		id,
		amount,
		note,
		expand: {
			payer: { name: payerName, email: payerEmail ?? `${payerName.toLowerCase()}@example.com` }
		}
	};
}

const PAYERS = [
	{ name: '张三', email: 'zhangsan@example.com' },
	{ name: '李四', email: 'lisi@example.com' },
	{ name: '王五', email: 'wangwu@example.com' },
	{ name: '赵六', email: 'zhaoliu@example.com' },
	{ name: 'Alice', email: 'alice@example.com' },
	{ name: 'Bob', email: 'bob@example.com' },
	{ name: 'Charlie', email: 'charlie@example.com' }
];

const NOTES = [
	'午餐',
	'晚餐',
	'打车',
	'电影票',
	'超市购物',
	'水电费',
	'房租',
	'咖啡',
	'零食',
	'快递费',
	'健身房',
	'书本费',
	'聚餐',
	'旅游门票',
	'机票',
	'酒店',
	'加油',
	'停车费',
	'电话费',
	'网费'
];

function generateTransactions(count: number): (TransactionLike & { id: string })[] {
	const result: (TransactionLike & { id: string })[] = [];
	for (let i = 0; i < count; i++) {
		const payer = PAYERS[i % PAYERS.length];
		const note = NOTES[i % NOTES.length];
		const amount = ((i * 137 + 500) % 100000) + 100;
		result.push(tx(`tx${String(i + 1).padStart(3, '0')}`, amount, note, payer.name, payer.email));
	}
	return result;
}

const TRANSACTIONS = generateTransactions(60);

describe('filterTransactions', () => {
	it('should return all transactions when keyword is empty', () => {
		expect(filterTransactions(TRANSACTIONS, '')).toHaveLength(60);
		expect(filterTransactions(TRANSACTIONS, '   ')).toHaveLength(60);
	});

	it('should filter by note (备注)', () => {
		const result = filterTransactions(TRANSACTIONS, '午餐');
		expect(result.length).toBeGreaterThan(0);
		for (const tx of result) {
			expect(tx.note).toContain('午餐');
		}
	});

	it('should filter by payer name (付款人名称)', () => {
		const result = filterTransactions(TRANSACTIONS, '张三');
		expect(result.length).toBeGreaterThan(0);
		for (const tx of result) {
			expect(tx.expand?.payer?.name).toContain('张三');
		}
	});

	it('should filter by payer email (付款人邮箱)', () => {
		const result = filterTransactions(TRANSACTIONS, 'alice@example.com');
		expect(result.length).toBeGreaterThan(0);
		for (const tx of result) {
			expect(tx.expand?.payer?.email).toContain('alice@example.com');
		}
	});

	it('should filter by partial email domain', () => {
		const result = filterTransactions(TRANSACTIONS, '@example.com');
		expect(result).toHaveLength(60);
	});

	it('should filter by amount (金额)', () => {
		const firstTx = TRANSACTIONS[0];
		const amountStr = (firstTx.amount / 100).toFixed(2);
		const result = filterTransactions(TRANSACTIONS, amountStr);
		expect(result.length).toBeGreaterThanOrEqual(1);
		expect(result[0].amount).toBe(firstTx.amount);
	});

	it('should be case-insensitive for English keywords', () => {
		const lower = filterTransactions(TRANSACTIONS, 'alice');
		const upper = filterTransactions(TRANSACTIONS, 'ALICE');
		const mixed = filterTransactions(TRANSACTIONS, 'Alice');
		expect(lower.length).toBe(upper.length);
		expect(lower.length).toBe(mixed.length);
		expect(lower.length).toBeGreaterThan(0);
	});

	it('should return empty array for non-matching keyword', () => {
		const result = filterTransactions(TRANSACTIONS, '不存在的关键词xyz999');
		expect(result).toHaveLength(0);
	});

	it('should handle transactions with null note', () => {
		const data: TransactionLike[] = [
			{ amount: 1000, note: null, expand: { payer: { name: 'Test', email: 't@e.com' } } },
			{ amount: 2000, note: '有备注', expand: { payer: { name: 'Test', email: 't@e.com' } } }
		];
		const result = filterTransactions(data, '有备注');
		expect(result).toHaveLength(1);
		expect(result[0].note).toBe('有备注');
	});

	it('should handle transactions with null expand/payer', () => {
		const data: TransactionLike[] = [
			{ amount: 1000, note: '测试', expand: null },
			{ amount: 2000, note: '测试', expand: { payer: null } },
			{ amount: 3000, note: '测试', expand: { payer: { name: null, email: null } } }
		];
		expect(filterTransactions(data, '测试')).toHaveLength(3);
		expect(filterTransactions(data, 'anything')).toHaveLength(0);
	});

	it('should match partial note text', () => {
		const result = filterTransactions(TRANSACTIONS, '餐');
		const expected = TRANSACTIONS.filter((t) => t.note?.includes('餐'));
		expect(result.length).toBe(expected.length);
		expect(result.length).toBeGreaterThan(0);
	});

	it('should match partial amount string', () => {
		const allAmounts = TRANSACTIONS.map((t) => (t.amount / 100).toFixed(2));
		const target = allAmounts[0].slice(0, 3);
		const result = filterTransactions(TRANSACTIONS, target);
		expect(result.length).toBeGreaterThanOrEqual(1);
	});

	it('should correctly filter 60 transactions (超过两页) by payer Bob', () => {
		const result = filterTransactions(TRANSACTIONS, 'Bob');
		const bobIndex = PAYERS.findIndex((p) => p.name === 'Bob');
		const expected = Math.floor(60 / PAYERS.length) + (bobIndex < 60 % PAYERS.length ? 1 : 0);
		expect(result.length).toBe(expected);
		for (const tx of result) {
			expect(tx.expand?.payer?.name).toBe('Bob');
		}
	});

	it('should correctly filter 60 transactions by 超市购物', () => {
		const result = filterTransactions(TRANSACTIONS, '超市购物');
		const expected = 60 / 20;
		expect(result.length).toBe(expected);
		for (const tx of result) {
			expect(tx.note).toBe('超市购物');
		}
	});

	it('should handle single character keyword across all fields', () => {
		const result = filterTransactions(TRANSACTIONS, 'a');
		expect(result.length).toBeGreaterThan(0);
	});

	it('should find transactions with specific amount like 12.50', () => {
		const data: TransactionLike[] = [
			{ amount: 1250, note: '咖啡', expand: { payer: { name: '张三', email: 'z@s.com' } } },
			{ amount: 2500, note: '午餐', expand: { payer: { name: '李四', email: 'l@s.com' } } },
			{ amount: 12500, note: '房租', expand: { payer: { name: '王五', email: 'w@s.com' } } }
		];
		const result = filterTransactions(data, '12.50');
		expect(result).toHaveLength(1);
		expect(result[0].amount).toBe(1250);
	});

	it('should handle very large dataset of 200 transactions', () => {
		const big = generateTransactions(200);
		expect(filterTransactions(big, '')).toHaveLength(200);
		const aliceIndex = PAYERS.findIndex((p) => p.name === 'Alice');
		const expectedAlice =
			Math.floor(200 / PAYERS.length) + (aliceIndex < 200 % PAYERS.length ? 1 : 0);
		expect(filterTransactions(big, 'Alice').length).toBe(expectedAlice);
		expect(filterTransactions(big, '午餐').length).toBe(200 / 20);
		expect(filterTransactions(big, 'nonexistent_keyword_abc')).toHaveLength(0);
	});
});

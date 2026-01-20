export type TransactionType = 'INCOME' | 'EXPENSE';

export interface TransactionSearchCriteria {
	q?: string;
	from?: string;
	until?: string;
	budget_id?: number;
	wallet_id?: number;
	type?: TransactionType;
	page?: number;
	pageSize?: number;
}

export interface Transaction {
	id: number;
	userId: number;
	date: string;
	budgetId: number;
	walletId: number;
	description: string;
	amountInCents: number;
	type: TransactionType;
	isPending: boolean;
	isDebt?: boolean | null;
	tags?: string[];
}

export interface TransactionDTO {
	id: number;
	date: string;
	description: string;
	amountInCents: number;
	type: TransactionType;
	budgetName: string;
	walletName: string;
	isPending: boolean;
}

export interface PaginatedTransactions {
	transactions: TransactionDTO[];
	total: number;
	page: number;
	pageSize: number;
}

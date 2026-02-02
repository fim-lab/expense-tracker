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

export interface Wallet {
	id: number;
	userId: number;
	name: string;
	balanceCents: number;
	canDelete: boolean;
	isEditing?: boolean;
	newName?: string;
}

export interface Budget {
	id: number;
	userId: number;
	name: string;
	limitCents: number;
	balanceCents: number;
	canDelete: boolean;
	isEditing?: boolean;
	newName?: string;
	newLimitEuros?: number;
}

export interface TransactionSearchCriteria {
	q?: string;
	from?: string;
	until?: string;
	budget_id?: number;
	wallet_id?: number;
	type?: 'INCOME' | 'EXPENSE';
	page?: number;
	pageSize?: number;
}

import type { Transaction, TransactionSearchCriteria } from "./types";

export const formatCurrency = (cents: number) => {
	return (cents / 100).toLocaleString('de-DE', {
		style: 'currency',
		currency: 'EUR'
	});
};

export async function searchTransactions(
	criteria: TransactionSearchCriteria
): Promise<Transaction[]> {
	const params = new URLSearchParams();

	if (criteria.q) {
		params.append("q", criteria.q);
	}
	if (criteria.from) {
		params.append("from", criteria.from);
	}
	if (criteria.until) {
		params.append("until", criteria.until);
	}
	if (criteria.budget_id) {
		params.append("budget_id", String(criteria.budget_id));
	}
	if (criteria.wallet_id) {
		params.append("wallet_id", String(criteria.wallet_id));
	}
	if (criteria.type) {
		params.append("type", criteria.type);
	}
	if (criteria.page) {
		params.append("page", String(criteria.page));
	}
	if (criteria.pageSize) {
		params.append("pageSize", String(criteria.pageSize));
	}

	const response = await fetch(`/api/transactions/search?${params.toString()}`);
	if (!response.ok) {
		throw new Error(`HTTP error! status: ${response.status}`);
	}
	const result = await response.json();
	return result.data || [];
}

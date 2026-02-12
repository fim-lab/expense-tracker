import type { PaginatedTransactions, TransactionSearchCriteria } from './types';

import { goto } from '$app/navigation';
import { page } from '$app/state';

export function updateParams(updates: Partial<TransactionSearchCriteria>, resetPage = true) {
	const params = new URLSearchParams(page.url.searchParams);

	for (const [key, value] of Object.entries(updates)) {
		if (value === undefined || value === null || value === '') {
			params.delete(key);
		} else {
			params.set(key, String(value));
		}
	}

	if (resetPage) {
		params.set('page', '1');
	}

	goto(`?${params.toString()}`, { keepFocus: true, replaceState: true });
}

export const formatCurrency = (cents: number) => {
	return (cents / 100).toLocaleString('de-DE', {
		style: 'currency',
		currency: 'EUR'
	});
};

export function getShortMonthYearString(date: Date): string {
	return date.toLocaleDateString('de-DE', { year: '2-digit', month: 'long' });
}

export function getLastMonthShortMonthYearString(date: Date): string {
	if (date.getMonth() == 0) {
		date.setMonth(11);
		date.setFullYear(date.getFullYear() - 1);
		return getShortMonthYearString(date);
	}
	date.setMonth(date.getMonth() - 1);
	return getShortMonthYearString(date);
}

export async function searchTransactions(
	criteria: TransactionSearchCriteria
): Promise<PaginatedTransactions> {
	const params = new URLSearchParams();

	if (criteria.q) {
		params.append('q', criteria.q);
	}
	if (criteria.from) {
		params.append('from', criteria.from);
	}
	if (criteria.until) {
		params.append('until', criteria.until);
	}
	if (criteria.budget_id) {
		params.append('budget_id', String(criteria.budget_id));
	}
	if (criteria.wallet_id) {
		params.append('wallet_id', String(criteria.wallet_id));
	}
	if (criteria.type) {
		params.append('type', criteria.type);
	}
	if (criteria.page) {
		params.append('page', String(criteria.page));
	}
	if (criteria.pageSize) {
		params.append('pageSize', String(criteria.pageSize));
	}

	const response = await fetch(`/api/transactions/search?${params.toString()}`);
	if (!response.ok) {
		throw new Error(`HTTP error! status: ${response.status}`);
	}
	const result = await response.json();
	return result.data || [];
}

export function debounce<T extends (...args: any[]) => void>(func: T, delay: number) {
	let timeout: ReturnType<typeof setTimeout>;

	return function (this: ThisParameterType<T>, ...args: Parameters<T>) {
		clearTimeout(timeout);
		timeout = setTimeout(() => func.apply(this, args), delay);
	};
}

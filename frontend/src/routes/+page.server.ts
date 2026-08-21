import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, url, cookies }) => {
	const cookieHeader = cookies
		.getAll()
		.map((c) => `${c.name}=${c.value}`)
		.join('; ');

	const authedApiFetch = async (path: string) => {
		const res = await fetch(`/api${path}`, {
			headers: { Cookie: cookieHeader }
		});

		if (res.status === 401) {
			throw redirect(302, '/login');
		}

		if (!res.ok) return null;
		return res.json();
	};

	const wallets = (await authedApiFetch('/wallets')) || [];
	const depots = (await authedApiFetch('/depots')) || [];
	const budgets = (await authedApiFetch('/budgets')) || [];
	const transactions = (await authedApiFetch(`/transactions/search?${url.searchParams}`)) ?? {
		data: [],
		total: 0,
		sumInCents: 0,
		page: 0,
		pageSize: 0
	};
	const debtSummary = (await authedApiFetch('/transactions/search?debt=true&pageSize=1')) ?? {
		total: 0,
		sumInCents: 0
	};

	return {
		transactions: transactions.transactions,
		total: transactions.total,
		sumInCents: transactions.sumInCents,
		page: transactions.page,
		pageSize: transactions.pageSize,
		wallets,
		depots,
		budgets,
		debtTotal: debtSummary.total,
		debtSumInCents: debtSummary.sumInCents
	};
};

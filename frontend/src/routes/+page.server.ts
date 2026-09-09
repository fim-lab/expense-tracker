import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, url, cookies }) => {
	const cookieHeader = cookies
		.getAll()
		.map((c) => `${c.name}=${c.value}`)
		.join('; ');

	const apiFetch = async (path: string) => {
		const res = await fetch(`/api${path}`, {
			headers: { Cookie: cookieHeader }
		});

		if (!res.ok) return null;
		return res.json();
	};

	const walletsRes = await fetch('/api/wallets', { headers: { Cookie: cookieHeader } });
	if (walletsRes.status === 401) {
		throw redirect(302, '/login');
	}
	const wallets = walletsRes.ok ? await walletsRes.json() : [];

	const depots = (await apiFetch('/depots')) || [];
	const budgets = (await apiFetch('/budgets')) || [];
	const budgetGroups = (await apiFetch('/budget-groups')) || [];
	const transactions = (await apiFetch(`/transactions/search?${url.searchParams}`)) ?? {
		transactions: [],
		total: 0,
		sumInCents: 0,
		page: 1,
		pageSize: 8
	};
	const debtSummary = (await apiFetch('/transactions/search?debt=true&pageSize=1')) ?? {
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
		budgetGroups,
		debtTotal: debtSummary.total,
		debtSumInCents: debtSummary.sumInCents * -1 // -1, because if you lend someone money (expense/negative), that does effectively mean, you own more money than is in your pocket (so add it/positive)
	};
};

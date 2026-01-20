import { error, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, cookies, params }) => {
	const API_URL = '/api';
	const cookieHeader = cookies.getAll().map((c) => `${c.name}=${c.value}`).join('; ');

	const [walletsRes, budgetsRes, transactionRes] = await Promise.all([
		fetch(`${API_URL}/wallets`, { headers: { Cookie: cookieHeader } }),
		fetch(`${API_URL}/budgets`, { headers: { Cookie: cookieHeader } }),
		fetch(`${API_URL}/transactions/${params.id}`, { headers: { Cookie: cookieHeader } })
	]);

	if (walletsRes.status === 401 || budgetsRes.status === 401 || transactionRes.status === 401) {
		throw error(401, 'Unauthorized');
	}

	if (!transactionRes.ok) {
		throw error(transactionRes.status, 'Failed to load transaction');
	}

	return {
		wallets: walletsRes.ok ? await walletsRes.json() : [],
		budgets: budgetsRes.ok ? await budgetsRes.json() : [],
		transaction: await transactionRes.json()
	};
};

export const actions: Actions = {
	default: async ({ request, fetch, cookies, params }) => {
		const API_URL = '/api';
		const cookieHeader = cookies.getAll().map((c) => `${c.name}=${c.value}`).join('; ');
		const data = await request.formData();

		const rfc3339Date = new Date(data.get('date') as string).toISOString();

		const payload = {
			description: data.get('description'),
			amountInCents: Math.round(Number(data.get('amount')) * 100),
			date: rfc3339Date,
			walletId: Number(data.get('walletId')),
			budgetId: Number(data.get('budgetId')),
			type: data.get('type')
		};

		const response = await fetch(`${API_URL}/transactions/${params.id}`, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json',
				Cookie: cookieHeader
			},
			body: JSON.stringify(payload)
		});

		if (!response.ok) {
			const res = await response.json();
			return {
				error: res.message
			};
		}

        const redirectParam = new URL(request.url).searchParams.get('redirect') || '/';
		const redirectUrl = decodeURIComponent(redirectParam);
		throw redirect(303, `/${redirectUrl}`);
	}
};

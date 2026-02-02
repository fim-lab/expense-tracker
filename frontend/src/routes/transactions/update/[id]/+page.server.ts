import { error, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, cookies, params }) => {
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
	const budgets = (await authedApiFetch('/budgets')) || [];
	const transaction = (await authedApiFetch(`/transactions/${params.id}`)) || {};

	if (!transaction.ok) {
		throw error(transaction.status, 'Failed to load transaction');
	}

	return {
		wallets,
		budgets,
		transaction
	};
};

export const actions: Actions = {
	default: async ({ request, fetch, cookies, params }) => {
		const cookieHeader = cookies
			.getAll()
			.map((c) => `${c.name}=${c.value}`)
			.join('; ');
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

		const response = await fetch(`/api/transactions/${params.id}`, {
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

		const redirectParam = new URL(request.url).searchParams.get('redirect');
		if (!redirectParam) {
			throw redirect(303, '/');
		}
		const redirectUrl = decodeURIComponent(redirectParam);
		throw redirect(303, `/${redirectUrl}`);
	}
};

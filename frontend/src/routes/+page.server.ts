import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

const LIMIT = 5;

export const load: PageServerLoad = async ({ fetch, url, cookies }) => {
    const page = Number(url.searchParams.get('page') ?? '1');
    const offset = (page - 1) * LIMIT;

    const API_URL = 'http://localhost:8080/api';
    const cookieHeader = cookies.getAll().map(c => `${c.name}=${c.value}`).join('; ');

    const authedFetch = async (path: string) => {
        const res = await fetch(`${API_URL}${path}`, {
            headers: { Cookie: cookieHeader }
        });

        if (res.status === 401) {
            throw redirect(302, '/login');
        }

        if (!res.ok) return null;
        return res.json();
    };

    const wallets = await authedFetch('/wallets') || [];
    const budgets = await authedFetch('/budgets') || [];
    const transactions = await authedFetch(
    `/transactions?limit=${LIMIT}&offset=${offset}`
  ) || [];

    return {
        transactions: transactions.data,
        total: transactions.total,
        page,
        limit: LIMIT,
        wallets,
        budgets
    };
};
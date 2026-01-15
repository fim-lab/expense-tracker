import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
  const API_URL = 'http://localhost:8080/api';
  const cookieHeader = cookies.getAll().map(c => `${c.name}=${c.value}`).join('; ');

  // Fetch wallets and budgets in parallel
  const [walletsRes, budgetsRes] = await Promise.all([
    fetch(`${API_URL}/wallets`, { headers: { Cookie: cookieHeader } }),
    fetch(`${API_URL}/budgets`, { headers: { Cookie: cookieHeader } })
  ]);

  if (walletsRes.status === 401) throw error(401, 'Unauthorized');

  return {
    wallets: walletsRes.ok ? await walletsRes.json() : [],
    budgets: budgetsRes.ok ? await budgetsRes.json() : []
  };
};
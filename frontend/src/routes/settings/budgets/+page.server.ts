import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
    const res = await fetch('/api/budgets');
    if (res.ok) {
        const budgets = await res.json();
        return { budgets };
    }
    return { budgets: [] };
};

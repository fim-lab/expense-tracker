import type { PageServerLoad } from './$types';
import type { User } from '$lib/types';

export const load: PageServerLoad = async ({ fetch }) => {
    const budgetRes = await fetch('/api/budgets');
    let budgets = [];
    if (budgetRes.ok) {
        budgets = await budgetRes.json();
    }

    const userRes = await fetch('/api/users/me');
    let user: User | null = null;
    if (userRes.ok) {
        user = await userRes.json();
    }
    
    return { budgets, user };
};
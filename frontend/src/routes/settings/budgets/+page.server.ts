import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
    const budgetRes = await fetch('/api/budgets');
    let budgets = [];
    if (budgetRes.ok) {
        budgets = await budgetRes.json();
    }

    const budgetGroupRes = await fetch('/api/budget-groups');
    let budgetGroups = [];
    if (budgetGroupRes.ok) {
        budgetGroups = await budgetGroupRes.json();
    }

    return { budgets, budgetGroups };
};
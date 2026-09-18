export interface BudgetProgress {
	rawPercentage: number;
	isOverBudget: boolean;
	isOverLowerLimit: boolean;
	displayPercentage: number;
	withinLimitWidth: number;
	overflowWidth: number;
	lowerLimitWidth: number;
}

export function computeBudgetProgress(balanceCents: number, limitCents: number): BudgetProgress {
	const rawPercentage = limitCents <= 0 ? 0 : (balanceCents / limitCents) * 100;
	const isOverBudget = rawPercentage > 100;
	const isOverLowerLimit = rawPercentage > 10;
	const displayPercentage = Math.max(0, Math.min(100, rawPercentage));
	const withinLimitWidth = isOverBudget ? 10000 / rawPercentage : displayPercentage;
	const overflowWidth = isOverBudget ? 100 - withinLimitWidth : 0;
	const lowerLimitWidth = Math.min(10, displayPercentage);
	return {
		rawPercentage,
		isOverBudget,
		isOverLowerLimit,
		displayPercentage,
		withinLimitWidth,
		overflowWidth,
		lowerLimitWidth
	};
}

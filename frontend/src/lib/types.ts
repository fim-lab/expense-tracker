export type TransactionType = 'INCOME' | 'EXPENSE';

export interface TransactionSearchCriteria {
	q?: string;
	from?: string;
	until?: string;
	budget_id?: number;
	budget_group_id?: number;
	wallet_id?: number;
	type?: TransactionType;
	debt?: boolean;
	page?: number;
	pageSize?: number;
}

export interface Transaction {
	id: number;
	userId: number;
	date: string;
	budgetId: number;
	walletId: number;
	description: string;
	amountInCents: number;
	type: TransactionType;
	isPending: boolean;
	isDebt?: boolean | null;
	tags?: string[];
}

export interface TransactionDTO {
	id: number;
	date: string;
	description: string;
	amountInCents: number;
	type: TransactionType;
	budgetName: string;
	walletName: string;
	isPending: boolean;
	isDebt: boolean;
}

export interface PaginatedTransactions {
	transactions: TransactionDTO[];
	total: number;
	sumInCents: number;
	page: number;
	pageSize: number;
}

export interface Wallet {
	id: number;
	userId: number;
	name: string;
	balanceCents: number;
	canDelete: boolean;
	isEditing?: boolean;
	newName?: string;
}

export interface User {
	id: number;
}

export interface Budget {
	id: number;
	userId: number;
	name: string;
	limitCents: number;
	balanceCents: number;
	canDelete: boolean;
	groupId?: number | null;
	isEditing?: boolean;
	newName?: string;
	newLimitEuros?: number;
	newGroupId?: number;
}

export interface BudgetGroup {
	id: number;
	userId: number;
	name: string;
	isEditing?: boolean;
	newName?: string;
}

export interface Depot {
	id: number;
	userId?: number;
	walletId: number;
	budgetId: number;
	name: string;
	investedInCents?: number;
	currentValueInCents?: number;
	isEditing?: boolean;
	newName?: string;
	newWalletId?: number;
	newBudgetId?: number;
}

export interface TransactionTemplate {
	id: number;
	userId: number;
	day: number; // Day of the month (1-31)
	budgetId: number | null;
	groupId?: number | null;
	walletId: number;
	description: string;
	amountInCents: number;
	type: TransactionType;
	tags?: string[];
	budgetName?: string;
	walletName?: string;
	newGroupId?: number;
}

export interface TemplateGroup {
	id: number;
	userId: number;
	name: string;
	isEditing?: boolean;
	newName?: string;
}

export type TradeType = 'BUY' | 'SELL';

export interface Trade {
	id: number;
	depotId: number;
	walletTransactionId: number | null;
	stockId?: number;
	wkn: string;
	type: TradeType;
	quantity: number;
	totalInCents: number;
	feesInCents: number;
	taxesInCents: number;
	timestamp: string;
}

export interface TradeDTO extends Trade {
	costBasisInCents: number;
	proceedsInCents: number;
	realizedGainInCents: number;
	canDelete: boolean;
}

export interface Lot {
	tradeId: number;
	depotId: number;
	stockId: number;
	dateOfPurchase: string;
	quantity: number;
	remaining: number;
	totalInCents: number;
	remainingCostInCents: number;
}

export interface Position {
	depotId: number;
	stockId: number;
	wkn: string;
	ticker?: string;
	quantity: number;
	investedInCents: number;
	avgPriceInCents: number;
	currentPriceInCents?: number;
	currentValueInCents?: number;
	unrealizedGainInCents?: number;
	lots: Lot[];
}

export interface Portfolio {
	depotId: number;
	positions: Position[];
	investedInCents: number;
	realizedGainInCents: number;
	currentValueInCents?: number;
	unrealizedGainInCents?: number;
}

export interface Stock {
	id: number;
	wkn: string;
	ticker: string;
	priceInCents: number;
	lastFetched: string | null;
	isEditing?: boolean;
	newWkn?: string;
	newTicker?: string;
	newPriceEuros?: number;
}

/**
 * Format a currency amount with proper locale formatting
 * @param amount - Amount in cents (will be divided by 100)
 * @param currency - Currency code (defaults to USD)
 * @returns Formatted currency string
 */
export function formatCurrency(amount: number, currency: string = 'USD'): string {
	return new Intl.NumberFormat('en-US', {
		style: 'currency',
		currency: currency.toUpperCase()
	}).format(amount / 100);
}

/**
 * Format a currency amount without dividing by 100 (for amounts already in dollars)
 * @param amount - Amount in dollars
 * @param currency - Currency code (defaults to USD)
 * @returns Formatted currency string
 */
export function formatCurrencyRaw(amount: number, currency: string = 'USD'): string {
	return new Intl.NumberFormat('en-US', {
		style: 'currency',
		currency: currency.toUpperCase()
	}).format(amount);
}

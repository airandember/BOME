/**
 * Date utility functions for consistent date handling across the application
 */

/**
 * Converts an ISO date string to HTML date input format (yyyy-MM-dd)
 * @param isoDate - ISO date string or null
 * @returns Formatted date string for HTML date input or empty string
 */
export function isoToDateInput(isoDate: string | null): string {
	if (!isoDate) return '';
	
	try {
		// Handle timezone-aware dates by extracting just the date part
		const dateMatch = isoDate.match(/^(\d{4}-\d{2}-\d{2})/);
		if (dateMatch) {
			return dateMatch[1];
		}
		
		// Fallback to Date object parsing
		const date = new Date(isoDate);
		if (isNaN(date.getTime())) return '';
		
		return date.toISOString().split('T')[0];
	} catch (error) {
		console.error('Error converting ISO date to date input format:', error);
		return '';
	}
}

/**
 * Converts an HTML date input value to ISO date string
 * @param dateInput - Date string in yyyy-MM-dd format
 * @returns ISO date string or null
 */
export function dateInputToIso(dateInput: string | null): string | null {
	if (!dateInput) return null;
	
	try {
		const date = new Date(dateInput);
		if (isNaN(date.getTime())) return null;
		
		return date.toISOString();
	} catch (error) {
		console.error('Error converting date input to ISO format:', error);
		return null;
	}
}

/**
 * Formats a date for display in a user-friendly format
 * Handles timezone-aware dates properly to prevent date shifting
 * @param isoDate - ISO date string or null
 * @param options - Intl.DateTimeFormatOptions for formatting
 * @returns Formatted date string
 */
export function formatDateForDisplay(
	isoDate: string | null, 
	options: Intl.DateTimeFormatOptions = { 
		year: 'numeric', 
		month: 'short', 
		day: 'numeric' 
	}
): string {
	if (!isoDate) return '';
	
	try {
		// Extract the date part from timezone-aware strings
		// Handle formats like "2025-07-31 00:00:00-06" or "2025-08-01 00:00:00"
		const dateMatch = isoDate.match(/^(\d{4}-\d{2}-\d{2})/);
		if (dateMatch) {
			const dateOnly = dateMatch[1];
			const [year, month, day] = dateOnly.split('-').map(Number);
			
			// Create date in local timezone to avoid shifting
			const date = new Date(year, month - 1, day);
			return date.toLocaleDateString('en-US', options);
		}
		
		// Fallback to standard Date parsing
		const date = new Date(isoDate);
		if (isNaN(date.getTime())) return '';
		
		return date.toLocaleDateString('en-US', options);
	} catch (error) {
		console.error('Error formatting date for display:', error);
		return '';
	}
}

/**
 * Gets today's date in HTML date input format
 * @returns Today's date in yyyy-MM-dd format
 */
export function getTodayDateInput(): string {
	return new Date().toISOString().split('T')[0];
}

/**
 * Gets a date X days from today in HTML date input format
 * @param daysFromToday - Number of days to add (can be negative)
 * @returns Date in yyyy-MM-dd format
 */
export function getDateInputFromToday(daysFromToday: number): string {
	const date = new Date();
	date.setDate(date.getDate() + daysFromToday);
	return date.toISOString().split('T')[0];
}

/**
 * Validates if a date string is in the correct format for HTML date input
 * @param dateString - Date string to validate
 * @returns True if valid, false otherwise
 */
export function isValidDateInput(dateString: string): boolean {
	if (!dateString) return false;
	
	const regex = /^\d{4}-\d{2}-\d{2}$/;
	if (!regex.test(dateString)) return false;
	
	const date = new Date(dateString);
	return !isNaN(date.getTime());
} 
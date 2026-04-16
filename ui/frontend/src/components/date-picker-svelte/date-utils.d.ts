import type { FormatToken } from './parse';
export declare function isLeapYear(year: number): boolean;
export declare function getMonthLength(year: number, month: number): number;
export declare function toText(date: Date | null, formatTokens: FormatToken[]): string;
export declare type CalendarDay = {
    year: number;
    month: number;
    number: number;
};
export declare function getMonthDays(year: number, month: number): CalendarDay[];
export declare function getCalendarDays(value: Date, weekStartsOn: number): CalendarDay[];

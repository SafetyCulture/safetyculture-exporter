export declare type Locale = {
    weekdays?: string[];
    months?: string[];
    weekStartsOn?: number;
};
declare type InnerLocale = {
    weekdays: string[];
    months: string[];
    weekStartsOn: number;
};
export declare function getLocaleDefaults(): InnerLocale;
export declare function getInnerLocale(locale?: Locale): InnerLocale;
declare type DateFnsLocale = {
    options?: {
        weekStartsOn?: 0 | 1 | 2 | 3 | 4 | 5 | 6;
    };
    localize?: {
        month: (n: number, options?: {
            width?: string;
        }) => string;
        day: (i: number, options?: {
            width?: string;
        }) => string;
    };
};
/** Create a Locale from a date-fns locale */
export declare function localeFromDateFnsLocale(dateFnsLocale: DateFnsLocale): InnerLocale;
export {};

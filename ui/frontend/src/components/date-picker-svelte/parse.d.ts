declare type RuleToken = {
    id: string;
    toString: (d: Date) => string;
};
export declare type FormatToken = string | RuleToken;
declare type ParseResult = {
    date: Date | null;
    missingPunctuation: string;
};
/** Parse a string according to the supplied format tokens. Returns a date if successful, and the missing punctuation if there is any that should be after the string */
export declare function parse(str: string, tokens: FormatToken[], baseDate: Date | null): ParseResult;
export declare function createFormat(s: string): FormatToken[];
export {};

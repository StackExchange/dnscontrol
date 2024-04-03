/**
 * @dnscontrol-auto-doc-comment language-reference/top-level-functions/FETCH.md
 */
declare function FETCH(
    url: string,
    init?: {
        method?:
            | 'GET'
            | 'POST'
            | 'PUT'
            | 'PATCH'
            | 'DELETE'
            | 'HEAD'
            | 'OPTIONS';
        headers?: { [key: string]: string | string[] };
        // Ignored by the underlying code
        // redirect: 'follow' | 'error' | 'manual';
        body?: string;
    }
): Promise<FetchResponse>;

interface FetchResponse {
    readonly bodyUsed: boolean;
    readonly headers: ResponseHeaders;
    readonly ok: boolean;
    readonly status: number;
    readonly statusText: string;
    readonly type: string;

    text(): Promise<string>;
    json(): Promise<any>;
}

interface ResponseHeaders {
    get(name: string): string | undefined;
    getAll(name: string): string[];
    has(name: string): boolean;

    append(name: string, value: string): void;
    delete(name: string): void;
    set(name: string, value: string): void;
}

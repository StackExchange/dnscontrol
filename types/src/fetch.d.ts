/**
 * `FETCH` is a wrapper for the [Fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API). This allows dynamically setting DNS records based on an external data source, e.g. the API of your cloud provider.
 *
 * Compared to `fetch` from Fetch API, `FETCH` will call [PANIC](https://dnscontrol.org/js#PANIC) to terminate the execution of the script, and therefore DNSControl, if a network error occurs.
 *
 * Otherwise the syntax of `FETCH` is the same as `fetch`.
 *
 * `FETCH` is not enabled by default. Please read the warnings below.
 *
 * > WARNING:
 * >
 * > 1. Relying on external sources adds a point of failure. If the external source doesn't work, your script won't either. Please make sure you are aware of the consequences.
 * > 2. Make sure DNSControl only uses verified configuration if you want to use `FETCH`. For example, an attacker can send Pull Requests to your config repo, and have your CI test malicious configurations and make arbitrary HTTP requests. Therefore, `FETCH` must be explicitly enabled with flag `--allow-fetch` on DNSControl invocation.
 *
 * ```js
 * var REG_NONE = NewRegistrar('none');
 * var DNS_BIND = NewDnsProvider('bind');
 *
 * D('example.com', REG_NONE, DnsProvider(DNS_BIND), [
 *   A('@', '1.2.3.4'),
 * ]);
 *
 * FETCH('https://example.com', {
 *   // All three options below are optional
 *   headers: {"X-Authentication": "barfoo"},
 *   method: "POST",
 *   body: "Hello World",
 * }).then(function(r) {
 *   return r.text();
 * }).then(function(t) {
 *   // Example of generating record based on response
 *   D_EXTEND('example.com', [
 *     TXT('@', t.slice(0, 100)),
 *   ]);
 * });
 * ```
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

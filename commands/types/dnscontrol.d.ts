// WARNING: These type definitions are experimental and subject to change in future releases.

interface Domain {
    name: string;
    subdomain: string;
    registrar: unknown;
    meta: Record<string, unknown>;
    records: DNSRecord[];
    dnsProviders: Record<string, unknown>;
    defaultTTL: number;
    nameservers: unknown[];
    ignored_names: unknown[];
    ignored_targets: unknown[];
    [key: string]: unknown;
}

interface DNSRecord {
    type: string;
    meta: Record<string, unknown>;
    ttl: number;
}

type DomainModifier =
    | ((domain: Domain) => void)
    | Partial<Domain>
    | DomainModifier[];

type RecordModifier =
    | ((record: DNSRecord) => void)
    | Partial<DNSRecord['meta']>;

type Duration =
    | `${number}${'s' | 'm' | 'h' | 'd' | 'w' | 'n' | 'y' | ''}`
    | number /* seconds */;


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


declare function require(name: `${string}.json`): any;
declare function require(name: string): true;

/**
 * Issuer critical flag. CA that does not understand this tag will refuse to issue certificate for this domain.
 *
 * CAA record is supported only by BIND, Google Cloud DNS, Amazon Route 53 and OVH. Some certificate authorities may not support this record until the mandatory date of September 2017.
 */
declare const CAA_CRITICAL: RecordModifier;

/**
 * This disables a safety check intended to prevent:
 * 1. Two owners toggling a record between two settings.
 * 2. The other owner wiping all records at this label, which won't
 * be noticed until the next time dnscontrol is run.
 * See https://github.com/StackExchange/dnscontrol/issues/1106
 */
declare const IGNORE_NAME_DISABLE_SAFETY_CHECK: RecordModifier;

// Cloudflare aliases:

/** Proxy disabled. */
declare const CF_PROXY_OFF: RecordModifier;
/** Proxy enabled. */
declare const CF_PROXY_ON: RecordModifier;
/** Proxy+Railgun enabled. */
declare const CF_PROXY_FULL: RecordModifier;

/** Proxy default off for entire domain (the default) */
declare const CF_PROXY_DEFAULT_OFF: DomainModifier;
/** Proxy default on for entire domain */
declare const CF_PROXY_DEFAULT_ON: DomainModifier;
/** UniversalSSL off for entire domain */
declare const CF_UNIVERSALSSL_OFF: DomainModifier;
/** UniversalSSL on for entire domain */
declare const CF_UNIVERSALSSL_ON: DomainModifier;

/**
 * Set default values for CLI variables. See: https://dnscontrol.org/cli-variables
 */
declare function CLI_DEFAULTS(vars: Record<string, unknown>): void;

/**
 * `END` permits the last item to include a comma.
 *
 * ```js
 * D("foo.com", ...
 *    A(...),
 *    A(...),
 *    A(...),
 * END)
 * ```
 */
declare const END: DomainModifier & RecordModifier;

/**
 * Permit labels like `"foo.bar.com.bar.com"` (normally an error)
 *
 * ```js
 * D("bar.com", ...
 *     A("foo.bar.com", "10.1.1.1", DISABLE_REPEATED_DOMAIN_CHECK),
 * )
 * ```
 */
declare const DISABLE_REPEATED_DOMAIN_CHECK: RecordModifier;


/**
 * A adds an A record To a domain. The name should be the relative label for the record. Use `@` for the domain apex.
 * 
 * The address should be an ip address, either a string, or a numeric value obtained via [IP](../global/IP.md).
 * 
 * Modifiers can be any number of [record modifiers](https://docs.dnscontrol.org/language-reference/record-modifiers) or JSON objects, which will be merged into the record's metadata.
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider("R53"),
 *   A("@", "1.2.3.4"),
 *   A("foo", "2.3.4.5"),
 *   A("test.foo", IP("1.2.3.4"), TTL(5000)),
 *   A("*", "1.2.3.4", {foo: 42})
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/a
 */
declare function A(name: string, address: string | number, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * AAAA adds an AAAA record To a domain. The name should be the relative label for the record. Use `@` for the domain apex.
 * 
 * The address should be an IPv6 address as a string.
 * 
 * Modifiers can be any number of [record modifiers](https://docs.dnscontrol.org/language-reference/record-modifiers) or JSON objects, which will be merged into the record's metadata.
 * 
 * ```javascript
 * var addrV6 = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
 * 
 * D("example.com", REGISTRAR, DnsProvider("R53"),
 *   AAAA("@", addrV6),
 *   AAAA("foo", addrV6),
 *   AAAA("test.foo", addrV6, TTL(5000)),
 *   AAAA("*", addrV6, {foo: 42})
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/aaaa
 */
declare function AAAA(name: string, address: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * AKAMAICDN is a proprietary record type that is used to configure [Zone Apex Mapping](https://blogs.akamai.com/2019/08/fast-dns-zone-apex-mapping-dnssec.html).
 * The AKAMAICDN target must be preconfigured in the Akamai network.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/service-provider-specific/akamai-edge-dns/akamaicdn
 */
declare function AKAMAICDN(name: string, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * ALIAS is a virtual record type that points a record at another record. It is analogous to a CNAME, but is usually resolved at request-time and served as an A record. Unlike CNAMEs, ALIAS records can be used at the zone apex (`@`)
 * 
 * Different providers handle ALIAS records differently, and many do not support it at all. Attempting to use ALIAS records with a DNS provider type that does not support them will result in an error.
 * 
 * The name should be the relative label for the domain.
 * 
 * Target should be a string representing the target. If it is a single label we will assume it is a relative name on the current domain. If it contains *any* dots, it should be a fully qualified domain name, ending with a `.`.
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider("CLOUDFLARE"),
 *   ALIAS("@", "google.com."), // example.com -> google.com
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/alias
 */
declare function ALIAS(name: string, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * AUTODNSSEC_OFF tells the provider to disable AutoDNSSEC. It takes no
 * parameters.
 * 
 * See `AUTODNSSEC_ON` for further details.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/autodnssec_off
 */
declare const AUTODNSSEC_OFF: DomainModifier;

/**
 * AUTODNSSEC_ON tells the provider to enable AutoDNSSEC.
 * 
 * AUTODNSSEC_OFF tells the provider to disable AutoDNSSEC.
 * 
 * AutoDNSSEC is a feature where a DNS provider can automatically manage
 * DNSSEC for a domain. Not all providers support this.
 * 
 * At this time, AUTODNSSEC_ON takes no parameters.  There is no ability
 * to tune what the DNS provider sets, no algorithm choice.  We simply
 * ask that they follow their defaults when enabling a no-fuss DNSSEC
 * data model.
 * 
 * NOTE: No parenthesis should follow these keywords.  That is, the
 * correct syntax is `AUTODNSSEC_ON` not `AUTODNSSEC_ON()`
 * 
 * ```javascript
 * D("example.com", .... ,
 *   AUTODNSSEC_ON,  // Enable AutoDNSSEC.
 *   A("@", "10.1.1.1")
 * );
 * 
 * D("insecure.com", .... ,
 *   AUTODNSSEC_OFF,  // Disable AutoDNSSEC.
 *   A("@", "10.2.2.2")
 * );
 * ```
 * 
 * If neither `AUTODNSSEC_ON` or `AUTODNSSEC_OFF` is specified for a
 * domain no changes will be requested.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/autodnssec_on
 */
declare const AUTODNSSEC_ON: DomainModifier;

/**
 * AZURE_ALIAS is a Azure specific virtual record type that points a record at either another record or an Azure entity.
 * It is analogous to a CNAME, but is usually resolved at request-time and served as an A record.
 * Unlike CNAMEs, ALIAS records can be used at the zone apex (`@`)
 * 
 * Unlike the regular ALIAS directive, AZURE_ALIAS is only supported on AZURE.
 * Attempting to use AZURE_ALIAS on another provider than Azure will result in an error.
 * 
 * The name should be the relative label for the domain.
 * 
 * The type can be any of the following:
 * * A
 * * AAAA
 * * CNAME
 * 
 * Target should be the Azure Id representing the target. It starts `/subscription/`. The resource id can be found in https://resources.azure.com/.
 * 
 * The Target can :
 * 
 * * Point to a public IP resource from a DNS `A/AAAA` record set.
 * You can create an A/AAAA record set and make it an alias record set to point to a public IP resource (standard or basic).
 * The DNS record set changes automatically if the public IP address changes or is deleted.
 * Dangling DNS records that point to incorrect IP addresses are avoided.
 * There is a current limit of 20 alias records sets per resource.
 * * Point to a Traffic Manager profile from a DNS `A/AAAA/CNAME` record set.
 * You can create an A/AAAA or CNAME record set and use alias records to point it to a Traffic Manager profile.
 * It's especially useful when you need to route traffic at a zone apex, as traditional CNAME records aren't supported for a zone apex.
 * For example, say your Traffic Manager profile is myprofile.trafficmanager.net and your business DNS zone is contoso.com.
 * You can create an alias record set of type A/AAAA for contoso.com (the zone apex) and point to myprofile.trafficmanager.net.
 * * Point to an Azure Content Delivery Network (CDN) endpoint.
 * This is useful when you create static websites using Azure storage and Azure CDN.
 * * Point to another DNS record set within the same zone.
 * Alias records can reference other record sets of the same type.
 * For example, a DNS CNAME record set can be an alias to another CNAME record set.
 * This arrangement is useful if you want some record sets to be aliases and some non-aliases.
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider("AZURE_DNS"),
 *   AZURE_ALIAS("foo", "A", "/subscriptions/726f8cd6-6459-4db4-8e6d-2cd2716904e2/resourceGroups/test/providers/Microsoft.Network/trafficManagerProfiles/testpp2"), // record for traffic manager
 *   AZURE_ALIAS("foo", "CNAME", "/subscriptions/726f8cd6-6459-4db4-8e6d-2cd2716904e2/resourceGroups/test/providers/Microsoft.Network/dnszones/example.com/A/quux."), // record in the same zone
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/service-provider-specific/azure-dns/azure_alias
 */
declare function AZURE_ALIAS(name: string, type: "A" | "AAAA" | "CNAME", target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `CAA()` adds a CAA record to a domain. The name should be the relative label for the record. Use `@` for the domain apex.
 * 
 * Tag can be one of
 * 1. `"issue"`
 * 2. `"issuewild"`
 * 3. `"iodef"`
 * 
 * Value is a string. The format of the contents is different depending on the tag. DNSControl will handle any escaping or quoting required, similar to TXT records. For example use `CAA("@", "issue", "letsencrypt.org")` rather than `CAA("@", "issue", "\"letsencrypt.org\"")`.
 * 
 * Flags are controlled by modifier:
 * - `CAA_CRITICAL`: Issuer critical flag. CA that does not understand this tag will refuse to issue certificate for this domain.
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider("GCLOUD"),
 *   // Allow letsencrypt to issue certificate for this domain
 *   CAA("@", "issue", "letsencrypt.org"),
 *   // Allow no CA to issue wildcard certificate for this domain
 *   CAA("@", "issuewild", ";"),
 *   // Report all violation to test@example.com. If CA does not support
 *   // this record then refuse to issue any certificate
 *   CAA("@", "iodef", "mailto:test@example.com", CAA_CRITICAL)
 * );
 * ```
 * 
 * DNSControl contains a [`CAA_BUILDER`](../record/CAA_BUILDER.md) which can be used to simply create `CAA()` records for your domains. Instead of creating each CAA record individually, you can simply configure your report mail address, the authorized certificate authorities and the builder cares about the rest.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/caa
 */
declare function CAA(name: string, tag: "issue" | "issuewild" | "iodef", value: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `CF_REDIRECT` uses Cloudflare-specific features ("Forwarding URL" Page Rules) to
 * generate a HTTP 301 permanent redirect.
 * 
 * If _any_ `CF_REDIRECT` or [`CF_TEMP_REDIRECT`](CF_TEMP_REDIRECT.md) functions are used then
 * `dnscontrol` will manage _all_ "Forwarding URL" type Page Rules for the domain.
 * Page Rule types other than "Forwarding URL” will be left alone.
 * 
 * WARNING: Cloudflare does not currently fully document the Page Rules API and
 * this interface is not extensively tested. Take precautions such as making
 * backups and manually verifying `dnscontrol preview` output before running
 * `dnscontrol push`. This is especially true when mixing Page Rules that are
 * managed by DNSControl and those that aren't.
 * 
 * HTTP 301 redirects are cached by browsers forever, usually ignoring any TTLs or
 * other cache invalidation techniques. It should be used with great care. We
 * suggest using a `CF_TEMP_REDIRECT` initially, then changing to a `CF_REDIRECT`
 * only after sufficient time has elapsed to prove this is what you really want.
 * 
 * This example redirects the bare (aka apex, or naked) domain to www:
 * 
 * ```javascript
 * D("foo.com", .... ,
 *   CF_REDIRECT("mydomain.com/*", "https://www.mydomain.com/$1"),
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/service-provider-specific/cloudflare-dns/cf_redirect
 */
declare function CF_REDIRECT(source: string, destination: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `CF_TEMP_REDIRECT` uses Cloudflare-specific features ("Forwarding URL" Page
 * Rules) to generate a HTTP 302 temporary redirect.
 * 
 * If _any_ [`CF_REDIRECT`](CF_REDIRECT.md) or `CF_TEMP_REDIRECT` functions are used then
 * `dnscontrol` will manage _all_ "Forwarding URL" type Page Rules for the domain.
 * Page Rule types other than "Forwarding URL” will be left alone.
 * 
 * WARNING: Cloudflare does not currently fully document the Page Rules API and
 * this interface is not extensively tested. Take precautions such as making
 * backups and manually verifying `dnscontrol preview` output before running
 * `dnscontrol push`. This is especially true when mixing Page Rules that are
 * managed by DNSControl and those that aren't.
 * 
 * ```javascript
 * D("foo.com", .... ,
 *   CF_TEMP_REDIRECT("example.mydomain.com/*", "https://otherplace.yourdomain.com/$1"),
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/service-provider-specific/cloudflare-dns/cf_temp_redirect
 */
declare function CF_TEMP_REDIRECT(source: string, destination: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `CF_WORKER_ROUTE` uses the [Cloudflare Workers](https://developers.cloudflare.com/workers/)
 * API to manage [worker routes](https://developers.cloudflare.com/workers/platform/routes)
 * for a given domain.
 * 
 * If _any_ `CF_WORKER_ROUTE` function is used then `dnscontrol` will manage _all_
 * Worker Routes for the domain. To be clear: this means it will delete existing routes that
 * were created outside of DNSControl.
 * 
 * WARNING: This interface is not extensively tested. Take precautions such as making
 * backups and manually verifying `dnscontrol preview` output before running
 * `dnscontrol push`.
 * 
 * This example assigns the patterns `api.foo.com/*` and `foo.com/api/*` to a `my-worker` script:
 * 
 * ```javascript
 * D("foo.com", .... ,
 *     CF_WORKER_ROUTE("api.foo.com/*", "my-worker"),
 *     CF_WORKER_ROUTE("foo.com/api/*", "my-worker"),
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/service-provider-specific/cloudflare-dns/cf_worker_route
 */
declare function CF_WORKER_ROUTE(pattern: string, script: string): DomainModifier;

/**
 * Documentation needed.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/service-provider-specific/cloudns/cloudns_wr
 */
declare function CLOUDNS_WR(name: string, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * CNAME adds a CNAME record to the domain. The name should be the relative label for the domain.
 * Using `@` or `*` for CNAME records is not recommended, as different providers support them differently.
 * 
 * Target should be a string representing the CNAME target. If it is a single label we will assume it is a relative name on the current domain. If it contains *any* dots, it should be a fully qualified domain name, ending with a `.`.
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider("R53"),
 *   CNAME("foo", "google.com."), // foo.example.com -> google.com
 *   CNAME("abc", "@"), // abc.example.com -> example.com
 *   CNAME("def", "test"), // def.example.com -> test.example.com
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/cname
 */
declare function CNAME(name: string, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * DS adds a DS record to the domain.
 * 
 * Key Tag should be a number.
 * 
 * Algorithm should be a number.
 * 
 * Digest Type must be a number.
 * 
 * Digest must be a string.
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider(R53),
 *   DS("example.com", 2371, 13, 2, "ABCDEF")
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/ds
 */
declare function DS(name: string, keytag: number, algorithm: number, digesttype: number, digest: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * DefaultTTL sets the TTL for all subsequent records following it in a domain that do not explicitly set one with [`TTL`](../record/TTL.md). If neither `DefaultTTL` or `TTL` exist for a record,
 * the record will inherit the DNSControl global internal default of 300 seconds. See also [`DEFAULTS`](../global/DEFAULTS.md) to override the internal defaults.
 * 
 * NS records are currently a special case, and do not inherit from `DefaultTTL`. See [`NAMESERVER_TTL`](../domain/NAMESERVER_TTL.md) to set a default TTL for all NS records.
 * 
 * ```javascript
 * D('example.com', REGISTRAR, DnsProvider('R53'),
 *   DefaultTTL("4h"),
 *   A('@','1.2.3.4'), // uses default
 *   A('foo', '2.3.4.5', TTL(600)) // overrides default
 * );
 * ```
 * 
 * The DefaultTTL duration is the same format as [`TTL`](../record/TTL.md), an integer number of seconds
 * or a string with a unit such as `'4d'`.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/defaultttl
 */
declare function DefaultTTL(ttl: Duration): DomainModifier;

/**
 * DnsProvider indicates that the specified provider should be used to manage
 * records for this domain. The name must match the name used with [NewDnsProvider](../global/NewDnsProvider.md).
 * 
 * The nsCount parameter determines how the nameservers will be managed from this provider.
 * 
 * Leaving the parameter out means "fetch and use all nameservers from this provider as authoritative". ie: `DnsProvider("name")`
 * 
 * Using `0` for nsCount means "do not fetch nameservers from this domain, or give them to the registrar".
 * 
 * Using a different number, ie: `DnsProvider("name",2)`, means "fetch all nameservers from this provider,
 * but limit it to this many.
 * 
 * See [this page](../../nameservers.md) for a detailed explanation of how DNSControl handles nameservers and NS records.
 * 
 * If a domain (`D()`) does not include any `DnsProvider()` functions,
 * the DNS records will not be modified. In fact, if you want to control
 * the Registrar for a domain but not the DNS records themselves, simply
 * do not include a `DnsProvider()` function for that `D()`.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/dnsprovider
 */
declare function DnsProvider(name: string, nsCount?: number): DomainModifier;

/**
 * Documentation needed.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/frame
 */
declare function FRAME(name: string, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * WARNING: The `IGNORE_*` family  of functions is risky to use. The code
 * is brittle and has subtle bugs. Use at your own risk. Do not use these
 * commands with `D_EXTEND()`.
 * 
 * `IGNORE_NAME` can be used to ignore some records present in zone.
 * Records of that name will be completely ignored. An optional `rTypes` may be specified as a comma separated list to only ignore records of the given type, e.g. `"A"`, `"A,CNAME"`, `"A, MX, CNAME"`. If `rTypes` is omitted or is `"*"` all record types matching the name will be ignored.
 * 
 * `IGNORE_NAME` is like `NO_PURGE` except it acts only on some specific records instead of the whole zone.
 * 
 * Technically `IGNORE_NAME` is a promise that DNSControl will not add, change, or delete records at a given label.  This permits another entity to "own" that label.
 * 
 * `IGNORE_NAME` is generally used in very specific situations:
 * 
 * * Some records are managed by some other system and DNSControl is only used to manage some records and/or keep them updated. For example a DNS `A` record that is managed by a dynamic DNS client, or by Kubernetes External DNS, but DNSControl is used to manage the rest of the zone. In this case we don't want DNSControl to try to delete the externally managed record.
 * * To work-around a pseudo record type that is not supported by DNSControl. For example some providers have a fake DNS record type called "URL" which creates a redirect. DNSControl normally deletes these records because it doesn't understand them. `IGNORE_NAME` will leave those records alone.
 * 
 * In this example, DNSControl will insert/update the "baz.example.com" record but will leave unchanged the "foo.example.com" and "bar.example.com" ones.
 * 
 * ```javascript
 * D("example.com",
 *   IGNORE_NAME("foo"), // ignore all record types for name foo
 *   IGNORE_NAME("baz", "*"), // ignore all record types for name baz
 *   IGNORE_NAME("bar", "A,MX"), // ignore only A and MX records for name bar
 *   CNAME("bar", "www"), // CNAME is not ignored
 *   A("baz", "1.2.3.4")
 * );
 * ```
 * 
 * `IGNORE_NAME` also supports glob patterns in the style of the [gobwas/glob](https://github.com/gobwas/glob) library. All of
 * the following patterns will work:
 * 
 * * `IGNORE_NAME("*.foo")` will ignore all records in the style of `bar.foo`, but will not ignore records using a double
 * subdomain, such as `foo.bar.foo`.
 * * `IGNORE_NAME("**.foo")` will ignore all subdomains of `foo`, including double subdomains.
 * * `IGNORE_NAME("?oo")` will ignore all records of three symbols ending in `oo`, for example `foo` and `zoo`. It will
 * not match `.`
 * * `IGNORE_NAME("[abc]oo")` will ignore records `aoo`, `boo` and `coo`. `IGNORE_NAME("[a-c]oo")` is equivalent.
 * * `IGNORE_NAME("[!abc]oo")` will ignore all three symbol records ending in `oo`, except for `aoo`, `boo`, `coo`. `IGNORE_NAME("[!a-c]oo")` is equivalent.
 * * `IGNORE_NAME("{bar,[fz]oo}")` will ignore `bar`, `foo` and `zoo`.
 * * `IGNORE_NAME("\\*.foo")` will ignore the literal record `*.foo`.
 * 
 * # Caveats
 * 
 * It is considered as an error to try to manage an ignored record.
 * Ignoring a label is a promise that DNSControl won't meddle with
 * anything at a particular label, therefore DNSControl prevents you from
 * adding records at a label that is `IGNORE_NAME`'ed.
 * 
 * Use `IGNORE_NAME("@")` to ignore at the domain's apex. Most providers
 * insert magic or unchangeable records at the domain's apex; usually `NS`
 * and `SOA` records.  DNSControl treats them specially.
 * 
 * # Errors
 * 
 * * `trying to update/add IGNORE_NAME'd record: foo CNAME`
 * 
 * This means you have both ignored `foo` and included a record (in this
 * case, a CNAME) to update it.  This is an error because `IGNORE_NAME`
 * is a promise not to modify records at a certain label so that others
 * may have free reign there.  Therefore, DNSControl prevents you from
 * modifying that label.
 * 
 * The `foo CNAME` at the end of the message indicates the label name
 * (`foo`) and the type of record (`CNAME`) that your dnsconfig.js file
 * is trying to insert.
 * 
 * You can override this error by adding the
 * `IGNORE_NAME_DISABLE_SAFETY_CHECK` flag to the record.
 * 
 *     TXT('vpn', "this thing", IGNORE_NAME_DISABLE_SAFETY_CHECK)
 * 
 * Disabling this safety check creates two risks:
 * 
 * 1. Two owners (DNSControl and some other entity) toggling a record between two settings.
 * 2. The other owner wiping all records at this label, which won't be noticed until the next time DNSControl is run.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/ignore_name
 */
declare function IGNORE_NAME(pattern: string, rTypes?: string): DomainModifier;

/**
 * WARNING: The `IGNORE_*` family  of functions is risky to use. The code
 * is brittle and has subtle bugs. Use at your own risk. Do not use these
 * commands with `D_EXTEND()` or use it at the domain apex.
 * 
 * IGNORE_TARGET can be used to ignore some records present in zone based on the record's target and type. IGNORE_TARGET currently only supports CNAME record types.
 * 
 * IGNORE_TARGET is like NO_PURGE except it acts only on some specific records instead of the whole zone.
 * 
 * IGNORE_TARGET is generally used in very specific situations:
 * 
 * * Some records are managed by some other system and DNSControl is only used to manage some records and/or keep them updated. For example a DNS record that is created by AWS Certificate Manager for validation, but DNSControl is used to manage the rest of the zone. In this case we don't want DNSControl to try to delete the externally managed record.
 * 
 * In this example, DNSControl will insert/update the "baz.example.com" record but will leave unchanged a CNAME to "foo.acm-validations.aws" record.
 * 
 * ```javascript
 * D("example.com",
 *   IGNORE_TARGET('**.acm-validations.aws.', 'CNAME'),
 *   A("baz", "1.2.3.4")
 * );
 * ```
 * 
 * IGNORE_TARGET also supports glob patterns in the style of the [gobwas/glob](https://github.com/gobwas/glob#example) library. Some example patterns:
 * 
 * * `IGNORE_TARGET("example.com", "CNAME")` will ignore all CNAME records with targets of exactly `example.com`.
 * * `IGNORE_TARGET("*.foo", "CNAME")` will ignore all CNAME records with targets in the style of `bar.foo`, but will not ignore records with targets using a double subdomain, such as `foo.bar.foo`.
 * * `IGNORE_TARGET("**.bar", "CNAME")` will ignore all CNAME records with target subdomains of `bar`, including double subdomains such as `www.foo.bar`.
 * * `IGNORE_TARGET("dev.*.foo", "CNAME")` will ignore all CNAME records with targets in the style of `dev.bar.foo`, but will not ignore records with targets using a double subdomain, such as `dev.foo.bar.foo`.
 * 
 * It is considered as an error to try to manage an ignored record.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/ignore_target
 */
declare function IGNORE_TARGET(pattern: string, rType: string): DomainModifier;

/**
 * Includes all records from a given domain
 * 
 * ```javascript
 * D("example.com!external", REGISTRAR, DnsProvider(R53),
 *   A("test", "8.8.8.8")
 * );
 * 
 * D("example.com!internal", REGISTRAR, DnsProvider(R53),
 *   INCLUDE("example.com!external"),
 *   A("home", "127.0.0.1")
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/include
 */
declare function INCLUDE(domain: string): DomainModifier;

/**
 * The parameter number types are as follows:
 * 
 * ```
 * name: string
 * target: string
 * deg1: uint32
 * min1: uint32
 * sec1: float32
 * deg2: uint32
 * min2: uint32
 * sec2: float32
 * altitude: uint32
 * size: float32
 * horizontal_precision: float32
 * vertical_precision: float32
 * ```
 * 
 * ## Description ##
 * 
 * Strictly follows [RFC 1876](https://datatracker.ietf.org/doc/html/rfc1876).
 * 
 * A LOC record holds a geographical position. In the zone file, it may look like:
 * 
 * ```text
 * ;
 * pipex.net.                    LOC   52 14 05 N 00 08 50 E 10m
 * ```
 * 
 * On the wire, it is in a binary format.
 * 
 * A use case for LOC is suggested in the RFC:
 * 
 * > Some uses for the LOC RR have already been suggested, including the
 *    USENET backbone flow maps, a "visual traceroute" application showing
 *    the geographical path of an IP packet, and network management
 *    applications that could use LOC RRs to generate a map of hosts and
 *    routers being managed.
 * 
 * There is the UK based [https://find.me.uk](https://find.me.uk/) whereby you can do:
 * 
 * ```sh
 * dig loc <uk-postcode>.find.me.uk
 * ```
 * 
 * There are some behaviours that you should be aware of, however:
 * 
 * > If omitted, minutes and seconds default to zero, size defaults to 1m,
 *    horizontal precision defaults to 10000m, and vertical precision
 *    defaults to 10m.  These defaults are chosen to represent typical
 *    ZIP/postal code area sizes, since it is often easy to find
 *    approximate geographical location by ZIP/postal code.
 * 
 * Alas, the world does not revolve around US ZIP codes, but here we are. Internally,
 * the LOC record type will supply defaults where values were absent on DNS import.
 * One must supply the `LOC()` js helper all parameters. If that seems like too
 * much work, see also helper functions:
 * 
 *  * [`LOC_BUILDER_DD({})`](../record/LOC_BUILDER_DD.md) - build a `LOC` by supplying only **d**ecimal **d**egrees.
 *  * [`LOC_BUILDER_DMS_STR({})`](../record/LOC_BUILDER_DMS_STR.md) - accepts DMS 33°51′31″S 151°12′51″E
 *  * [`LOC_BUILDER_DMM_STR({})`](../record/LOC_BUILDER_DMM_STR.md) - accepts DMM 25.24°S 153.15°E
 *  * [`LOC_BUILDER_STR({})`](../record/LOC_BUILDER_STR.md) - tries the cooordinate string in all `LOC_BUILDER_DM*_STR()` functions until one works
 * 
 * ## Format ##
 * 
 * The coordinate format for `LOC()` is: 
 * 
 * `degrees,minutes,seconds,[NnSs],deg,min,sec,[EeWw],altitude,size,horizontal_precision,vertical_precision`
 * 
 * ## Examples ##
 * 
 * ```javascript
 * D("foo.com","none"
 *   // LOC "subdomain", d1, m1, s1, "[NnSs]", d2, m2, s2, "[EeWw]", alt, siz, hp, vp)
 *   //42 21 54     N  71 06  18     W -24m 30m
 *   , LOC("@", 42, 21, 54,     "N", 71,  6, 18,     "W", -24,   30,    0,  0)
 *   //42 21 43.952 N  71 5   6.344  W -24m 1m 200m 10m
 *   , LOC("a", 42, 21, 43.952, "N", 71,  5,  6.344, "W", -24,    1,  200, 10)
 *   //52 14 05     N  00 08  50     E 10m
 *   , LOC("b", 52, 14,  5,     "N",  0,  8, 50,     "E",  10,    0,    0,  0)
 *   //32  7 19     S 116  2  25     E 10m
 *   , LOC("c", 32,  7, 19,     "S",116,  2, 25,     "E",  10,    0,    0,  0)
 *   //42 21 28.764 N  71 00  51.617 W -44m 2000m
 *   , LOC("d", 42, 21, 28.764, "N", 71,  0, 51.617, "W", -44, 2000,    0,  0)
 * );
 * 
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/loc
 */
declare function LOC(deg1: number, min1: number, sec1: number, deg2: number, min2: number, sec2: number, altitude: number, size: number, horizontal_precision: number, vertical_precision: number): DomainModifier;

/**
 * MX adds an MX record to the domain.
 * 
 * Priority should be a number.
 * 
 * Target should be a string representing the MX target. If it is a single label we will assume it is a relative name on the current domain. If it contains *any* dots, it should be a fully qualified domain name, ending with a `.`.
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider(R53),
 *   MX("@", 5, "mail"), // mx example.com -> mail.example.com
 *   MX("sub", 10, "mail.foo.com.")
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/mx
 */
declare function MX(name: string, priority: number, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `NAMESERVER()` instructs DNSControl to inform the domain's registrar where to find this zone.
 * For some registrars this will also add NS records to the zone itself.
 * 
 * This takes exactly one argument: the name of the nameserver. It must end with
 * a "." if it is a FQDN, just like all targets.
 * 
 * This is different than the [`NS()`](NS.md) function, which inserts NS records
 * in the current zone and accepts a label. [`NS()`](NS.md) is useful for downward
 * delegations. `NAMESERVER()` is for informing upstream delegations.
 * 
 * For more information, refer to [this page](../../nameservers.md).
 * 
 * ```javascript
 * D("example.com", REGISTRAR, .... ,
 *   DnsProvider(route53, 0),
 *   // Replace the nameservers:
 *   NAMESERVER("ns1.myserver.com."),
 *   NAMESERVER("ns2.myserver.com."),
 * );
 * 
 * D("example2.com", REGISTRAR, .... ,
 *   // Add these two additional nameservers to the existing list of nameservers.
 *   NAMESERVER("ns1.myserver.com."),
 *   NAMESERVER("ns2.myserver.com."),
 * );
 * ```
 * 
 * # The difference between NS() and NAMESERVER()
 * 
 * Nameservers are one of the least
 * understood parts of DNS, so a little extra explanation is required.
 * 
 * * [`NS()`](NS.md) lets you add an NS record to a zone, just like [`A()`](A.md) adds an A
 *   record to the zone. This is generally used to delegate a subzone.
 * 
 * * The `NAMESERVER()` directive speaks to the Registrar about how the parent should delegate the zone.
 * 
 * Since the parent zone could be completely unrelated to the current
 * zone, changes made by `NAMESERVER()` have to be done by an API call to
 * the registrar, who then figures out what to do. For example, if I
 * use `NAMESERVER()` in the zone `stackoverflow.com`, DNSControl talks to
 * the registrar who does the hard work of talking to the people that
 * control `.com`.  If the domain was `gmeet.io`, the registrar does
 * the right thing to talk to the people that control `.io`.
 * 
 * (A better name might have been `PARENTNAMESERVER()` but we didn't
 * think of that at the time.)
 * 
 * Each registrar handles delegations differently.  Most use
 * the `NAMESERVER()` targets to update the delegation, adding
 * `NS` records to the parent zone as required.
 * Some providers restrict the names to hosts they control.
 * Others may require you to add the `NS` records to the parent domain
 * manually.
 * 
 * # How to prevent changing the parent NS records?
 * 
 * If dnsconfig.js has zero `NAMESERVER()` commands for a domain, it will
 * use the API to remove all non-default nameservers.
 * 
 * If dnsconfig.js has 1 or more `NAMESERVER()` commands for a domain, it
 * will use the API to add those nameservers (unless, of course,
 * they already exist).
 * 
 * So how do you tell DNSControl not to make any changes at all?  Use the
 * special Registrar called "NONE". It makes no changes.
 * 
 * It looks like this:
 * 
 * ```javascript
 * var REG_THIRDPARTY = NewRegistrar('ThirdParty', 'NONE')
 * D("mydomain.com", REG_THIRDPARTY,
 *   ...
 * )
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/nameserver
 */
declare function NAMESERVER(name: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * NAMESERVER_TTL sets the TTL on the domain apex NS RRs defined by [`NAMESERVER`](NAMESERVER.md).
 * 
 * The value can be an integer or a string. See [`TTL`](../record/TTL.md) for examples.
 * 
 * ```javascript
 * D('example.com', REGISTRAR, DnsProvider('R53'),
 *   NAMESERVER_TTL('2d'),
 *   NAMESERVER('ns')
 * );
 * ```
 * 
 * Use `NAMESERVER_TTL('3600'),` or `NAMESERVER_TTL('1h'),` for a 1h default TTL for all subsequent `NS` entries:
 * 
 * ```javascript
 * D('example.com', REGISTRAR, DnsProvider('xyz'),
 *   DefaultTTL("4h"),
 *   NAMESERVER_TTL('3600'),
 *   NAMESERVER('ns1.provider.com.'), //inherits NAMESERVER_TTL
 *   NAMESERVER('ns2.provider.com.'), //inherits NAMESERVER_TTL
 *   A('@','1.2.3.4'), // inherits DefaultTTL
 *   A('foo', '2.3.4.5', TTL(600)) // overrides DefaultTTL for this record only
 * );
 * ```
 * 
 * To apply a default TTL to all other record types, see [`DefaultTTL`](../domain/DefaultTTL.md)
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/nameserver_ttl
 */
declare function NAMESERVER_TTL(ttl: Duration): DomainModifier;

/**
 * ## Introduction
 * 
 * NAPTR adds a NAPTR record to the domain. Various formats exist. NAPTR is a part of DDDS such as ENUM (defined by [RFC 6116](https://www.rfc-editor.org/rfc/rfc6116)), SIP ([RFC 3263](https://www.rfc-editor.org/rfc/rfc3263)), S-NAPTR ([RFC 3958](https://www.rfc-editor.org/rfc/rfc3958)) or U-NAPTR ([RFC 4848](https://www.rfc-editor.org/rfc/rfc4848)).
 * 
 * ## Parameters
 * 
 * ### `subdomain`
 * 
 * Subdomain of the domain (e.g. `example.com`) this entry represents. 
 * 
 * #### E164 
 * In the case of E164 (e.g. `3.2.1.5.5.5.0.0.8.1.e164.arpa.`) - where [`terminalflag`](#terminalflag) is `u` - the final digit of the zone it represents, or the zone apex record `@`. For example, the ARPA zone `3.2.1.5.5.5.0.0.8.1.e164.arpa.` represents the phone number block 001800555123*X* (or the synonymous +1800555123*X*), where *X* is the final digit of the phone number string, i.e. the [`subdomain`](#subdomain).
 * 
 * ### `order`
 * 
 * ordinal (1st, 2nd, 3rd, ...) 16 bit number (2^16 i.e. <= 65535) which determines lower entries are sent first (`1`), and  higher, last (`65535`).
 * 
 * ### `preference`
 * 
 * 16 bit number (2^16 i.e. <= 65535). At the DNS server, this entry is summed with other entries of identical [`order`](#order) value and normalised to a fraction of 100 percent, determining the likelihood that this record is returned by the DNS system. Effective for load balancing services.
 * 
 * ### `terminalflag`
 * (case insensitive)
 * 
 * One of [AaSsUuPp], where:
 *  * `a` (terminal lookup) means that the output of the [`target`](#target) rewrite will be a domain-name for which an [`A`](A.md) or [`AAAA`](AAAA.md) record should be queried
 *  * `p` Protocol specific
 *  * `s` (terminal lookup) indicates that [`target`](#target) points to a [`SRV`](SRV.md) record
 *  * `u` (terminal lookup) indicates that [`target`](#target) is a (SIP) URN or URI
 *  * "" (empty string) - a non-terminal condition defined by the ENUM application ([RFC 6116](https://www.rfc-editor.org/rfc/rfc6116)) to indicate that regexp is empty and the replace field contains the FQDN of another NAPTR RR
 * 
 * Mutually exclusive; more than one cannot be combined in the same record. Since there is no place for a port specification in the NAPTR record, when the `a` [`terminalflag`](#terminalflag) is used, the specified protocol must be running on its default port (Note that at least SIP URI forms allow ports to be specified).
 * 
 * Flags called 'terminal' halt the looping rewrite algorithm of DNS.
 * 
 * ### `service`
 * (case insensitive)
 * 
 * *`protocol+rs`* where *`protocol`* defines the protocol used by the DDDS application. *`rs`* is the resolution service. There may be 0 or more resolution services each separated by `+`. ENUM further defines this to be a type field and allows a subtype separated by a colon (`:`). 
 * 
 * For E164, typically one of `E2U+SIP` (or `E2U+sip`) or `E2U+email`. For SIP, typically `SIPS+D2T` for TCP/TLS `sips:` URIs, or TLS `sip:` URIs, or `SIP+D2T` for TCP based SIP, or `SIP+D2U` for UDP based SIP. Note that SCTP, WS and WSS are also available.
 * 
 * Valid [IANA registered services for ENUM](https://www.iana.org/assignments/enum-services/enum-services.xhtml#enum-services-1):
 * ```text
 * E2U+pres
 * E2U+voice:tel+sms:tel (compound form)
 * E2U+web:http
 * E2U+sms:mailto
 * E2U+sms:tel
 * E2U+sip
 * E2U+pstn
 * E2U+tel
 * ```
 * 
 * Valid [IANA registered SIP services](https://www.iana.org/assignments/sip-table/sip-table.xhtml#sip-table-1):
 * 
 * ```text
 * SIP+D2T
 * SIPS+D2T
 * SIP+D2U
 * SIP+D2S
 * SIPS+D2S
 * SIP+D2W
 * SIPS+D2W
 * ```
 * 
 * ### `regexp`
 * 
 * [Syntax: `delimit ere delimit substitution delimit flag`] an ERE or extended regular expression which captures any address string `.*` found between the line start `^` and finish `$` anchors (i.e. `!^.*$!`), and redirects it to the stated `sip:`, `sips:`, `tel:` or `mailto:` URI. Other URI forms may be possible. Other delimiter (`!`) forms are possible. The final `flag`, if any, shall be `i`, i.e. case **i**nsensitive.
 * 
 * Examples (taken from [Zytrax](https://www.zytrax.com/books/dns/ch8/naptr.html#regex-examples)):
 * ```text
 * # AUS = Application User String
 * # all examples use ! as the delimiter for consistency
 * # and simplicity
 * # AUS = +441115551234 in all cases
 * 
 * !(\\+441115551234)!tel:\\1!
 * # explicit check of all characters in string
 * # the +441115551234 because of () creates a group
 * # which is referenced by \1 in substitution
 * # result = tel:+441115551234
 * 
 * !^(\\+441115551234)$!tel:\\1!
 * # this is functionally identical to the expression
 * # above but uses ^ and $ to anchor both ends of
 * # the expression, there is no technical reason to do this
 * # within an ere and the RFCs are silent on the topic
 * # result = tel:+441115551234
 * 
 * !(.+)!tel:\\1!
 * # given the AUS of +441115551234
 * # the expression (.+) sets back ref 1 = +441115551234
 * # . = any character, + = 0 or more times
 * # result = tel:+441115551212
 * 
 * !\\+44111(.+)!sip:775\\1@some.example.com!
 * # given the AUS of +441115551234 provides partial replacement
 * # removes the 44111 part and substitutes 775
 * # result = sip:7755551234@some.example.com
 * 
 * !.*!sip:james@sip.example.com!
 * # reads and ignores AUS using .*
 * # and is called a simple replacement expression
 * # result = sip:james@sip.example.com
 * ```
 * 
 * U-NAPTR supported regexp fields must be of the form (from the RFC):
 * 
 * ```text
 * "!.*!<URI>!"
 * # the .* (any character 1 or more times)
 * # is fixed by the RFC and essentially ignores 
 * # the AUS data. The result will always be URI
 * ```
 * 
 * ### `target`
 * 
 * A (replacement) record for the target - format depends on [`terminalflag`](#terminalflag). 
 *  * A [`SRV`](SRV.md), if the [`terminalflag`](#terminalflag) is `s` (syntax: *`_Service._Proto.Name`*)
 *  * An [`A`](A.md) or [`AAAA`](AAAA.md) if the [`terminalflag`](#terminalflag) is `a`
 *  * URI if the [`terminalflag`](#terminalflag) is `u`
 * 
 * Not all examples are guaranteed to be standards compliant, or correct. 
 * 
 * ## Examples
 * 
 * ### Examples for e164 ARPA:
 * 
 * Individual e164 records
 * 
 * ```javascript
 * D("3.2.1.5.5.5.0.0.8.1.e164.arpa.", REGISTRAR, DnsProvider(R53),
 *   NAPTR('1',  10, 10, "u", "E2U+SIP", "!^.*$!sip:bob@example.com!", "."),
 *   NAPTR('2',  10, 10, "u", "E2U+SIP", "!^.*$!sip:alice@example.com!", "."),
 *   NAPTR('4',  10, 10, "u", "E2U+SIP", "!^.*$!sip:kate@example.com!", "."),
 *   NAPTR('5',  10, 10, "u", "E2U+SIP", "!^.*$!sip:steve@example.com!", "."),
 *   NAPTR('6',  10, 10, "u", "E2U+SIP", "!^.*$!sip:joe@example.com!", "."),
 *   NAPTR('7',  10, 10, "u", "E2U+SIP", "!^.*$!sip:jane@example.com!", "."),
 *   NAPTR('8',  10, 10, "u", "E2U+SIP", "!^.*$!sip:mike@example.com!", "."),
 *   NAPTR('9',  10, 10, "u", "E2U+SIP", "!^.*$!sip:linda@example.com!", "."),
 *   NAPTR('0',  10, 10, "u", "E2U+SIP", "!^.*$!sip:fax@example.com!", ".")
 * );
 * ```
 * 
 * Single e164 zone
 * ```javascript
 * D("4.3.2.1.5.5.5.0.0.8.1.e164.arpa.", REGISTRAR, DnsProvider(R53),
 *   NAPTR('@', 100, 50, "u", "E2U+SIP", "!^.*$!sip:customer-service@example.com!", "."),
 *   NAPTR('@', 101, 50, "u", "E2U+email", "!^.*$!mailto:information@example.com!", "."),
 *   NAPTR('@', 101, 50, "u", "smtp+E2U", "!^.*$!mailto:information@example.com!", ".")
 * );
 * ```
 * 
 * ### Examples for SIP:
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider(R53),
 *   NAPTR('@', 20, 50, "s", "SIPS+D2T", "", "_sips._tcp.example.com."),
 *   NAPTR('@', 20, 50, "s", "SIP+D2T", "", "_sip._tcp.example.com."),
 *   NAPTR('@', 30, 50, "s", "SIP+D2U", "", "_sip._udp.example.com."),
 *   NAPTR('help', 100, 50, "s", "SIP+D2U", "!^.*$!sip:customer-service@example.com!", "_sip._udp.example.com."),
 *   NAPTR('help', 101, 50, "s", "SIP+D2T", "!^.*$!sip:customer-service@example.com!", "_sip._tcp.example.com."),
 *   SRV('_sip._udp', 100, 0, 5060, 'sip.example.com.'),
 *   SRV('_sip._tcp', 100, 0, 5060, 'sip.example.com.'),
 *   SRV('_sips._tcp', 100, 0, 5061, 'sip.example.com.'),
 *   A('sip', '192.0.2.2'),
 *   AAAA('sip', '2001:db8::85a3'),
 *   // and so on
 * );
 * ```
 * 
 * ### Other RFC based examples:
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider(R53),
 *   NAPTR('@',100, 50, "a", "z3950+N2L+N2C", "", "cidserver.example.com."),
 *   NAPTR('@', 50, 50, "a", "rcds+N2C", "", "cidserver.example.com."),
 *   NAPTR('@', 30, 50, "s", "http+N2L+N2C+N2R", "", "www.example.com."),
 *   NAPTR('www',100,100, "s", "http+I2R", "", "_http._tcp.example.com."),
 *   NAPTR('www',100,100, "s", "ftp+I2R", "", "_ftp._tcp.example.com."),
 *   SRV('_z3950._tcp', 0, 0, 1000, 'z3950.beast.example.com.'),
 *   SRV('_http._tcp', 10, 0, 80, 'foo.example.com.'),
 *   // and so on
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/naptr
 */
declare function NAPTR(subdomain: string, order: number, preference: number, terminalflag: string, service: string, regexp: string, target: string): DomainModifier;

/**
 * `NO_PURGE` indicates that records should not be deleted from a domain.
 * Records will be added and updated, but not removed.
 * 
 * `NO_PURGE` is generally used in very specific situations:
 * 
 * * A domain is managed by some other system and DNSControl is only used to insert a few specific records and/or keep them updated. For example a DNS Zone that is managed by Active Directory, but DNSControl is used to update a few, specific, DNS records. In this case we want to specify the DNS records we are concerned with but not delete all the other records.  This is a risky use of `NO_PURGE` since, if `NO_PURGE` was removed (or buggy) there is a chance you could delete all the other records in the zone, which could be a disaster. That said, domains with some records updated using Dynamic DNS have no other choice.
 * * To work-around a pseudo record type that is not supported by DNSControl. For example some providers have a fake DNS record type called "URL" which creates a redirect. DNSControl normally deletes these records because it doesn't understand them. `NO_PURGE` will leave those records alone.
 * 
 * In this example DNSControl will insert "foo.example.com" into the
 * zone, but otherwise leave the zone alone.  Changes to "foo"'s IP
 * address will update the record. Removing the A("foo", ...) record
 * from DNSControl will leave the record in place.
 * 
 * ```javascript
 * D("example.com", .... , NO_PURGE,
 *   A("foo","1.2.3.4")
 * );
 * ```
 * 
 * The main caveat of `NO_PURGE` is that intentionally deleting records
 * becomes more difficult. Suppose a `NO_PURGE` zone has an record such
 * as A("ken", "1.2.3.4"). Removing the record from dnsconfig.js will
 * not delete "ken" from the domain. DNSControl has no way of knowing
 * the record was deleted from the file  The DNS record must be removed
 * manually.  Users of `NO_PURGE` are prone to finding themselves with
 * an accumulation of orphaned DNS records. That's easy to fix for a
 * small zone but can be a big mess for large zones.
 * 
 * Not all providers support `NO_PURGE`. For example the BIND provider
 * rewrites zone files from scratch each time, which precludes supporting
 * `NO_PURGE`.  DNSControl will exit with an error if `NO_PURGE` is used
 * on a driver that does not support it.
 * 
 * There is also [`PURGE`](PURGE.md) command for completeness. [`PURGE`](PURGE.md) is the
 * default, thus this command is a no-op.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/no_purge
 */
declare const NO_PURGE: DomainModifier;

/**
 * NS adds a NS record to the domain. The name should be the relative label for the domain.
 * 
 * The name may not be `@` (the bare domain), as that is controlled via [`NAMESERVER()`](NAMESERVER.md).
 * The difference between `NS()` and [`NAMESERVER()`](NAMESERVER.md) is explained in the [`NAMESERVER()` description](NAMESERVER.md).
 * 
 * Target should be a string representing the NS target. If it is a single label we will assume it is a relative name on the current domain. If it contains *any* dots, it should be a fully qualified domain name, ending with a `.`.
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider("R53"),
 *   NS("foo", "ns1.example2.com."), // Delegate ".foo.example.com" zone to another server.
 *   NS("foo", "ns2.example2.com."), // Delegate ".foo.example.com" zone to another server.
 *   A("ns1.example2.com", "10.10.10.10"), // Glue records
 *   A("ns2.example2.com", "10.10.10.20"), // Glue records
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/ns
 */
declare function NS(name: string, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * Documentation needed.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/service-provider-specific/ns1/ns1_urlfwd
 */
declare function NS1_URLFWD(name: string, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * This is a rough implementation of [RFC 7929](https://www.rfc-editor.org/rfc/rfc7929). Rough?
 * 
 * Rough: prior to the SHA256 hash step, a number of UTF8 normalization steps
 * are done, but not all.
 * 
 * `OPENPGPKEY({})` allows you to store a record of type `OPENPGPKEY`.
 * 
 * It currently takes two parameters in an object:
 * 
 *  * local - this is the local part of the email address (before the `@`) of the key
 * 	- everything after the `@` is discarded
 * 	- whitespace is removed
 * 	- various forms of quotation marks are removed
 * 	- several UTF normalization steps process the text to UTF8
 * 	- the 28 octet truncated SHA256 hash of the UTF8 is produced
 * 	- the hash is suffixed with `._openpgpkey` as specified in the RFC
 *  * digest - this is the base64 part of the key
 * 	- `-----* PGP PUBLIC KEY BLOCK-----` lines are discarded
 * 	- the CRC portion of any ASCII armored text block (radix64/base64) is discarded
 * 	- whitespace and linebreaks are removed
 * 
 *  
 * This is an (ed25519) Open PGP key:
 * ```text
 * -----BEGIN PGP PUBLIC KEY BLOCK-----
 * 
 * mDMEZCMu8xYJKwYBBAHaRw8BAQdAH4FTbN/H5SoMBl9Ez2cFQ1NuzymK894fq2ff
 * sYDvRkG0EWFsaWNlQGV4YW1wbGUuY29tiJYEExYKAD4CGwMFCwkIBwMFFQoJCAsF
 * FgIDAQACHgECF4AWIQRjw8oAQytQxDz5Q/Io7xpohfeBngUCZCMv5gUJAAk7ZgAK
 * CRAo7xpohfeBnlmVAP9k0slIpLwddCD1bZ9qVjqzNcS743OIDny7XuH6x02L2wEA
 * wxqAotO7/oUm0L4wyYR6hvGlhuGMSZXc9xMwZ1wVcA8=
 * =vHSO
 * -----END PGP PUBLIC KEY BLOCK-----
 * ```
 * 
 * The `digest` portion is the base64 portion without the trailing CRC portion
 * (the last base64 line starting `=` - in this case - `=vHSO`) at the end.
 * 
 * In effect:
 * 
 * ```text
 * mDMEZCMu8xYJKwYBBAHaRw8BAQdAH4FTbN/H5SoMBl9Ez2cFQ1NuzymK894fq2ff
 * sYDvRkG0EWFsaWNlQGV4YW1wbGUuY29tiJYEExYKAD4CGwMFCwkIBwMFFQoJCAsF
 * FgIDAQACHgECF4AWIQRjw8oAQytQxDz5Q/Io7xpohfeBngUCZCMv5gUJAAk7ZgAK
 * CRAo7xpohfeBnlmVAP9k0slIpLwddCD1bZ9qVjqzNcS743OIDny7XuH6x02L2wEA
 * wxqAotO7/oUm0L4wyYR6hvGlhuGMSZXc9xMwZ1wVcA8=
 * ```
 * 
 * Example:
 * 
 * ```javascript
 * D("example.com","none"
 *   // hugh@example.com -> c93f1e400f26708f98cb19d936620da35eec8f72e57f9eec01c1afd6._openpgpkey.example.com -> testing123 -> dGVzdGluZzEyMw==
 *   , OPENPGPKEY({local:"hugh@", digest:"dGVzdGluZzEyMw=="})
 *   // 麻衣子@example.com -> 2bb5bc4202aaecd48dcb54967c8e7f1b7574a436f04e0d15534b20e5._openpgpkey.example.com
 *   , OPENPGPKEY({local:"麻衣子@", digest:"\
 *   mDMEZCMxgRYJKwYBBAHaRw8BAQdA/fgtlQjGflt2MUMWhRZRnH5Hg+BY9sQTeePm\
 *   qqUs+lK0Fem6u+iho+WtkEBleGFtcGxlLmNvbYiWBBMWCgA+AhsDBQsJCAcDBRUK\
 *   CQgLBRYCAwEAAh4BAheAFiEEIWsEkWx5wygGCb61+tJ3q3m88E0FAmQjMbMFCQAJ\
 *   OqwACgkQ+tJ3q3m88E0z4gEAtowKJMPefyV5YCW8VubgXK7Fa+hjwXOPSsHnEnJw\
 *   9pUBAL+VZvNZv/VZvyGGMd31Yivqerzl6q+VIkZ6XffVb2AB\
 *   =sRIg"})
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/openpgpkey
 */
declare function OPENPGPKEY(opts: { local: string; digest: string; ...modifiers: RecordModifier[] }): DomainModifier;

/**
 * PTR adds a PTR record to the domain.
 * 
 * The name is normally a relative label for the domain, or a FQDN that ends with `.`.  If magic mode is enabled (see below) it can also be an IP address, which will be replaced by the proper string automatically, thus
 * saving the user from having to reverse the IP address manually.
 * 
 * Target should be a string representing the FQDN of a host.  Like all FQDNs in DNSControl, it must end with a `.`.
 * 
 * **Magic Mode:**
 * 
 * PTR records are complex and typos are common. Therefore DNSControl
 * enables features to save labor and
 * prevent typos.  This magic is only
 * enabled when the domain ends with `in-addr.arpa.` or `ipv6.arpa.`.
 * 
 * *Automatic IP-to-reverse:* If the name is a valid IP address, DNSControl will replace it with
 * a string that is appropriate for the domain. That is, if the domain
 * ends with `in-addr.arpa` (no `.`) and name is a valid IPv4 address, the name
 * will be replaced with the correct string to make a reverse lookup for that address.
 * IPv6 is properly handled too.
 * 
 * *Extra Validation:* DNSControl considers it an error to include a name that
 * is inappropriate for the domain.  For example
 * `PTR('1.2.3.4', 'f.co.')` is valid for the domain `D("3.2.1.in-addr.arpa',`
 *  but DNSControl will generate an error if the domain is `D("9.9.9.in-addr.arpa',`.
 * This is because `1.2.3.4` is contained in `1.2.3.0/24` but not `9.9.9.0/24`.
 * This validation works for IPv6, IPv4, and
 * RFC2317 "Classless in-addr.arpa delegation" domains.
 * 
 * *Automatic truncation:* DNSControl will automatically truncate FQDNs
 * as needed.
 * If the name is a FQDN ending with `.`, DNSControl will verify that the
 * name is contained within the CIDR block implied by domain.  For example
 * if name is `4.3.2.1.in-addr.arpa.` (note the trailing `.`)
 * and the domain is `2.1.in-addr.arpa` (no trailing `.`)
 * then the name will be replaced with `4.3`.  Note that the output
 * of `REV('1.2.3.4')` is `4.3.2.1.in-addr.arpa.`, which means the following
 * are all equivalent:
 * 
 * * `PTR(REV('1.2.3.4'), `
 * * `PTR('4.3.2.1.in-addr.arpa.'), `
 * * `PTR('4.3',`    // Assuming the domain is `2.1.in-addr.arpa`
 * 
 * All magic is RFC2317-aware. We use the first format listed in the
 * RFC for both [`REV()`](../global/REV.md) and `PTR()`. The format is
 * `FIRST/MASK.C.B.A.in-addr.arpa` where `FIRST` is the first IP address
 * of the zone, `MASK` is the netmask of the zone (25-31 inclusive),
 * and A, B, C are the first 3 octets of the IP address. For example
 * `172.20.18.130/27` is located in a zone named
 * `128/27.18.20.172.in-addr.arpa`
 * 
 * ```javascript
 * D(REV('1.2.3.0/24'), REGISTRAR, DnsProvider(BIND),
 *   PTR('1', 'foo.example.com.'),
 *   PTR('2', 'bar.example.com.'),
 *   PTR('3', 'baz.example.com.'),
 *   // If the first parameter is a valid IP address, DNSControl will generate the correct name:
 *   PTR('1.2.3.10', 'ten.example.com.'),    // '10'
 * );
 * ```
 * 
 * ```javascript
 * D(REV('9.9.9.128/25'), REGISTRAR, DnsProvider(BIND),
 *   PTR('9.9.9.129', 'first.example.com.'),
 * );
 * ```
 * 
 * ```javascript
 * D(REV('2001:db8:302::/48'), REGISTRAR, DnsProvider(BIND),
 *   PTR('1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0', 'foo.example.com.'),  // 2001:db8:302::1
 *   // If the first parameter is a valid IP address, DNSControl will generate the correct name:
 *   PTR('2001:db8:302::2', 'two.example.com.'),                          // '2.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0'
 *   PTR('2001:db8:302::3', 'three.example.com.'),                        // '3.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0'
 * );
 * ```
 * 
 * In the future we plan on adding a flag to [`A()`](A.md) which will insert
 * the correct PTR() record if the appropriate `.arpa` domain has been
 * defined.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/ptr
 */
declare function PTR(name: string, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `PURGE` is the default setting for all domains.  Therefore `PURGE` is
 * a no-op. It is included for completeness only.
 * 
 * A domain with a mixture of `NO_PURGE` and `PURGE` parameters will abide
 * by the last one.
 * 
 * These three examples all are equivalent.
 * 
 * `PURGE` is the default:
 * 
 * ```javascript
 * D("example.com", .... ,
 * );
 * ```
 * 
 * Purge is the default, but we set it anyway:
 * 
 * ```javascript
 * D("example.com", .... ,
 *   PURGE,
 * );
 * ```
 * 
 * Since the "last command wins", this is the same as `PURGE`:
 * 
 * ```javascript
 * D("example.com", .... ,
 *   PURGE,
 *   NO_PURGE,
 *   PURGE,
 *   NO_PURGE,
 *   PURGE,
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/purge
 */
declare const PURGE: DomainModifier;

/**
 * `R53_ALIAS` is a Route53 specific virtual record type that points a record at either another record or an AWS entity (like a Cloudfront distribution, an ELB, etc...). It is analogous to a `CNAME`, but is usually resolved at request-time and served as an `A` record. Unlike `CNAME` records, `ALIAS` records can be used at the zone apex (`@`)
 * 
 * Unlike the regular [`ALIAS`](ALIAS.md) directive, `R53_ALIAS` is only supported on Route53. Attempting to use `R53_ALIAS` on another provider than Route53 will result in an error.
 * 
 * The name should be the relative label for the domain.
 * 
 * Target should be a string representing the target. If it is a single label we will assume it is a relative name on the current domain. If it contains *any* dots, it should be a fully qualified domain name, ending with a `.`.
 * 
 * The Target can be any of:
 * 
 * * _CloudFront distribution_: in this case specify the domain name that CloudFront assigned when you created your distribution (note that your CloudFront distribution must include an alternate domain name that matches the record you're adding)
 * * _Elastic Beanstalk environment_: specify the `CNAME` attribute for the environment. The environment must have a regionalized domain name. To get the `CNAME`, you can use either the [AWS Console](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/customdomains.html), [AWS Elastic Beanstalk API](http://docs.aws.amazon.com/elasticbeanstalk/latest/api/API_DescribeEnvironments.html), or the [AWS CLI](http://docs.aws.amazon.com/cli/latest/reference/elasticbeanstalk/describe-environments.html).
 * * _ELB load balancer_: specify the DNS name that is associated with the load balancer. To get the DNS name you can use either the AWS Console (on the EC2 page, choose Load Balancers, select the right one, choose the description tab), [ELB API](http://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_DescribeLoadBalancers.html), the [AWS ELB CLI](http://docs.aws.amazon.com/cli/latest/reference/elb/describe-load-balancers.html), or the [AWS ELBv2 CLI](http://docs.aws.amazon.com/cli/latest/reference/elbv2/describe-load-balancers.html).
 * * _S3 bucket_ (configured as website): specify the domain name of the Amazon S3 website endpoint in which you configured the bucket (for instance s3-website-us-east-2.amazonaws.com). For the available values refer to the [Amazon S3 Website Endpoints](http://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region).
 * * _Another Route53 record_: specify the value of the name of another record in the same hosted zone.
 * 
 * For all the target type, excluding 'another record', you have to specify the `Zone ID` of the target. This is done by using the `R53_ZONE` record modifier.
 * 
 * The zone id can be found depending on the target type:
 * 
 * * _CloudFront distribution_: specify `Z2FDTNDATAQYW2`
 * * _Elastic Beanstalk environment_: specify the hosted zone ID for the region in which the environment has been created. Refer to the [List of regions and hosted Zone IDs](http://docs.aws.amazon.com/general/latest/gr/rande.html#elasticbeanstalk_region).
 * * _ELB load balancer_: specify the value of the hosted zone ID for the load balancer. You can find it in [the List of regions and hosted Zone IDs](http://docs.aws.amazon.com/general/latest/gr/rande.html#elb_region)
 * * _S3 bucket_ (configured as website): specify the hosted zone ID for the region that you created the bucket in. You can find it in [the List of regions and hosted Zone IDs](http://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region)
 * * _Another Route 53 record_: you can either specify the correct zone id or do not specify anything and DNSControl will figure out the right zone id. (Note: Route53 alias can't reference a record in a different zone).
 * 
 * ```javascript
 * D('example.com', REGISTRAR, DnsProvider('ROUTE53'),
 *   R53_ALIAS('foo', 'A', 'bar'),                              // record in same zone
 *   R53_ALIAS('foo', 'A', 'bar', R53_ZONE('Z35SXDOTRQ7X7K')),  // record in same zone, zone specified
 *   R53_ALIAS('foo', 'A', 'blahblah.elasticloadbalancing.us-west-1.amazonaws.com.', R53_ZONE('Z368ELLRRE2KJ0')),     // a classic ELB in us-west-1
 *   R53_ALIAS('foo', 'A', 'blahblah.elasticbeanstalk.us-west-2.amazonaws.com.', R53_ZONE('Z38NKT9BP95V3O')),     // an Elastic Beanstalk environment in us-west-2
 *   R53_ALIAS('foo', 'A', 'blahblah-bucket.s3-website-us-west-1.amazonaws.com.', R53_ZONE('Z2F56UZL2M1ACD')),     // a website S3 Bucket in us-west-1
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/service-provider-specific/amazon-route-53/r53_alias
 */
declare function R53_ALIAS(name: string, target: string, zone_idModifier: DomainModifier & RecordModifier): DomainModifier;

/**
 * `SOA` adds an `SOA` record to a domain. The name should be `@`.  ns and mbox are strings. The other fields are unsigned 32-bit ints.
 * 
 * ```javascript
 * D("example.com", REG_THIRDPARTY, DnsProvider("DNS_BIND"),
 *   SOA("@", "ns3.example.org.", "hostmaster@example.org", 3600, 600, 604800, 1440),
 * );
 * ```
 * 
 * If you accidentally include an `@` in the email field DNSControl will quietly
 * change it to a `.`. This way you can specify a human-readable email address
 * when you are making it easier for spammers how to find you.
 * 
 * ## Notes
 * * The serial number is managed automatically.  It isn't even a field in `SOA()`.
 * * Most providers automatically generate SOA records.  They will ignore any `SOA()` statements.
 * * The mbox field should not be set to a real email address unless you love spam and hate your privacy.
 * 
 * There is more info about `SOA` in the documentation for the [BIND provider](../../providers/bind.md).
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/soa
 */
declare function SOA(name: string, ns: string, mbox: string, refresh: number, retry: number, expire: number, minttl: number, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `SRV` adds a `SRV` record to a domain. The name should be the relative label for the record.
 * 
 * Priority, weight, and port are ints.
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider("GCLOUD"),
 *   // Create SRV records for a a SIP service:
 *   //               pr  w   port, target
 *   SRV('_sip._tcp', 10, 60, 5060, 'bigbox.example.tld.'),
 *   SRV('_sip._tcp', 10, 20, 5060, 'smallbox1.example.tld.'),
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/srv
 */
declare function SRV(name: string, priority: number, weight: number, port: number, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `SSHFP` contains a fingerprint of a SSH server which can be validated before SSH clients are establishing the connection.
 * 
 * **Algorithm** (type of the key)
 * 
 * | ID | Algorithm |
 * |----|-----------|
 * | 0  | reserved  |
 * | 1  | RSA       |
 * | 2  | DSA       |
 * | 3  | ECDSA     |
 * | 4  | ED25519   |
 * 
 * **Type** (fingerprint format)
 * 
 * | ID | Algorithm |
 * |----|-----------|
 * | 0  | reserved  |
 * | 1  | SHA-1     |
 * | 2  | SHA-256   |
 * 
 * `value` is the fingerprint as a string.
 * 
 * ```javascript
 * SSHFP('@', 1, 1, '00yourAmazingFingerprint00'),
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/sshfp
 */
declare function SSHFP(name: string, algorithm: 0 | 1 | 2 | 3 | 4, type: 0 | 1 | 2, value: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `TLSA` adds a `TLSA` record to a domain. The name should be the relative label for the record.
 * 
 * Usage, selector, and type are ints.
 * 
 * Certificate is a hex string.
 * 
 * ```javascript
 * D("example.com", REGISTRAR, DnsProvider("GCLOUD"),
 *   // Create TLSA record for certificate used on TCP port 443
 *   TLSA("_443._tcp", 3, 1, 1, "abcdef0"),
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/tlsa
 */
declare function TLSA(name: string, usage: number, selector: number, type: number, certificate: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `TXT` adds an `TXT` record To a domain. The name should be the relative
 * label for the record. Use `@` for the domain apex.
 * 
 * The contents is either a single or multiple strings.  To
 * specify multiple strings, specify them as an array.
 * 
 * Each string is a JavaScript string (quoted using single or double
 * quotes).  The (somewhat complex) quoting rules of the DNS protocol
 * will be done for you.
 * 
 * Modifiers can be any number of [record modifiers](https://docs.dnscontrol.org/language-reference/record-modifiers) or JSON objects, which will be merged into the record's metadata.
 * 
 * ```javascript
 *     D("example.com", REGISTRAR, ....,
 *       TXT('@', '598611146-3338560'),
 *       TXT('listserve', 'google-site-verification=12345'),
 *       TXT('multiple', ['one', 'two', 'three']),  // Multiple strings
 *       TXT('quoted', 'any "quotes" and escapes? ugh; no worries!'),
 *       TXT('_domainkey', 't=y; o=-;'), // Escapes are done for you automatically.
 *       TXT('long', 'X'.repeat(300)) // Long strings are split automatically.
 *     );
 * ```
 * 
 * NOTE: In the past, long strings had to be annotated with the keyword
 * `AUTOSPLIT`. This is no longer required. The keyword is now a no-op.
 * 
 * ### Long strings
 * 
 * Strings that are longer than 255 octets (bytes) will be quietly
 * split into 255-octets chunks or the provider may report an error
 * if it does not handle multiple strings.
 * 
 * ### TXT record edge cases
 * 
 * Most providers do not support the full possibilities of what a `TXT`
 * record can store.  DNSControl can not handle all the edge cases
 * and incompatibles that providers have introduced.  Instead, it
 * stores the string(s) that you provide and passes them to the provider
 * verbatim. The provider may opt to accept the data, fix it, or
 * reject it. This happens early in the processing, long before
 * the DNSControl talks to the provider's API.
 * 
 * The RFCs specify that a `TXT` record stores one or more strings,
 * each is up to 255 octets (bytes) long. We call these individual
 * strings *chunks*.  Each chunk may be zero to 255 octets long.
 * There is no limit to the number of chunks in a `TXT` record,
 * other than IP packet length restrictions.  The contents of each chunk
 * may be octets of value from 0x00 to 0xff.
 * 
 * In reality DNS Service Providers (DSPs) place many restrictions on `TXT`
 * records.
 * 
 * Some DSPs only support a single string of 255 octets or fewer.
 * Multiple strings, or any one string being longer than 255 octets will
 * result in an error. One provider limits the string to 254 octets,
 * which makes me think they're code has an off-by-one error.
 * 
 * Some DSPs only support one string, but it may be of any length.
 * Behind the scenes the provider splits it into 255-octet chunks
 * (except the last one, of course).
 * 
 * Some DSPs support multiple strings, but API requests must be 512-bytes
 * or fewer, and with quoting, escaping, and other encoding mishegoss
 * you can't be sure what will be permitted until you actually try it.
 * 
 * Regardless of the quantity and length of strings, some providers ban
 * double quotes, back-ticks, or other chars.
 * 
 * ### Testing the support of a provider
 * 
 * #### How can you tell if a provider will support a particular `TXT()` record?
 * 
 * Include the `TXT()` record in a [`D()`](../global/D.md) as usual, along
 * with the `DnsProvider()` for that provider.  Run `dnscontrol check` to
 * see if any errors are produced.  The check command does not talk to
 * the provider's API, thus permitting you to do this without having an
 * account at that provider.
 * 
 * #### What if the provider rejects a string that is supported?
 * 
 * Suppose I can create the TXT record using the DSP's web portal but
 * DNSControl rejects the string?
 * 
 * It is possible that the provider code in DNSControl rejects strings
 * that the DSP accepts.  This is because the test is done in code, not
 * by querying the provider's API.  It is possible that the code was
 * written to work around a bug (such as rejecting a string with a
 * back-tick) but now that bug has been fixed.
 * 
 * All such checks are in `providers/${providername}/auditrecords.go`.
 * You can try removing the check that you feel is in error and see if
 * the provider's API accepts the record.  You can do this by running the
 * integration tests, or by simply adding that record to an existing
 * `dnsconfig.js` and seeing if `dnscontrol push` is able to push that
 * record into production. (Be careful if you are testing this on a
 * domain used in production.)
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/txt
 */
declare function TXT(name: string, contents: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * Documentation needed.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/url
 */
declare function URL(name: string, target: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * Documentation needed.
 * 
 * @see https://docs.dnscontrol.org/language-reference/domain-modifiers/url301
 */
declare function URL301(name: string, ...modifiers: RecordModifier[]): DomainModifier;

/**
 * `D` adds a new Domain for DNSControl to manage. The first two arguments are required: the domain name (fully qualified `example.com` without a trailing dot), and the
 * name of the registrar (as previously declared with [NewRegistrar](NewRegistrar.md)). Any number of additional arguments may be included to add DNS Providers with [DNSProvider](NewDnsProvider.md),
 * add records with [A](../domain/A.md), [CNAME](../domain/CNAME.md), and so forth, or add metadata.
 * 
 * Modifier arguments are processed according to type as follows:
 * 
 * - A function argument will be called with the domain object as it's only argument. Most of the [built-in modifier functions](https://docs.dnscontrol.org/language-reference/domain-modifiers) return such functions.
 * - An object argument will be merged into the domain's metadata collection.
 * - An array argument will have all of it's members evaluated recursively. This allows you to combine multiple common records or modifiers into a variable that can
 *    be used like a macro in multiple domains.
 * 
 * ```javascript
 * var REGISTRAR = NewRegistrar("name.com");
 * var r53 = NewDnsProvider("R53");
 * 
 * // simple domain
 * D("example.com", REGISTRAR, DnsProvider(r53),
 *   A("@","1.2.3.4"),
 *   CNAME("test", "foo.example2.com.")
 * );
 * 
 * // "macro" for records that can be mixed into any zone
 * var GOOGLE_APPS_DOMAIN_MX = [
 *     MX('@', 1, 'aspmx.l.google.com.'),
 *     MX('@', 5, 'alt1.aspmx.l.google.com.'),
 *     MX('@', 5, 'alt2.aspmx.l.google.com.'),
 *     MX('@', 10, 'alt3.aspmx.l.google.com.'),
 *     MX('@', 10, 'alt4.aspmx.l.google.com.'),
 * ]
 * 
 * D("example.com", REGISTRAR, DnsProvider(r53),
 *   A("@","1.2.3.4"),
 *   CNAME("test", "foo.example2.com."),
 *   GOOGLE_APPS_DOMAIN_MX
 * );
 * ```
 * 
 * # Split Horizon DNS
 * 
 * DNSControl supports Split Horizon DNS. Simply
 * define the domain two or more times, each with
 * their own unique parameters.
 * 
 * To differentiate the different domains, specify the domains as
 * `domain.tld!tag`, such as `example.com!inside` and
 * `example.com!outside`.
 * 
 * ```javascript
 * var REG = NewRegistrar("Third-Party");
 * var DNS_INSIDE = NewDnsProvider("Cloudflare");
 * var DNS_OUTSIDE = NewDnsProvider("bind");
 * 
 * D("example.com!inside", REG, DnsProvider(DNS_INSIDE),
 *   A("www", "10.10.10.10")
 * );
 * 
 * D("example.com!outside", REG, DnsProvider(DNS_OUTSIDE),
 *   A("www", "20.20.20.20")
 * );
 * 
 * D_EXTEND("example.com!inside",
 *   A("internal", "10.99.99.99")
 * );
 * ```
 * 
 * A domain name without a `!` is assigned a tag that is the empty
 * string. For example, `example.com` and `example.com!` are equivalent.
 * However, we strongly recommend against using the empty tag, as it
 * risks creating confusion.  In other words, if you have `domain.tld`
 * and `domain.tld!external` you now require humans to remember that
 * `domain.tld` is the external one.  I mean... the internal one.  You
 * may have noticed this mistake, but will your coworkers?  Will you in
 * six months? You get the idea.
 * 
 * DNSControl command line flag `--domains` matches the full name (with the "!").  If you
 * define domains `example.com!george` and `example.com!john` then:
 * 
 * * `--domains=example.com` will not match either domain.
 * * `--domains='example.com!george'` will match only match the first.
 * * `--domains='example.com!george',example.com!john` will match both.
 * 
 * NOTE: The quotes are required if your shell treats `!` as a special
 * character, which is probably does.  If you see an error that mentions
 * `event not found` you probably forgot the quotes.
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/d
 */
declare function D(name: string, registrar: string, ...modifiers: DomainModifier[]): void;

/**
 * `DEFAULTS` allows you to declare a set of default arguments to apply to all subsequent domains. Subsequent calls to [`D`](D.md) will have these
 * arguments passed as if they were the first modifiers in the argument list.
 * 
 * ## Example
 * 
 * We want to create backup zone files for all domains, but not actually register them. Also create a [`DefaultTTL`](../domain/DefaultTTL.md).
 * The domain `example.com` will have the defaults set.
 * 
 * ```javascript
 * var COMMON = NewDnsProvider("foo");
 * DEFAULTS(
 *   DnsProvider(COMMON, 0),
 *   DefaultTTL('1d')
 * );
 * 
 * D("example.com",
 *   REGISTRAR,
 *   DnsProvider("R53"),
 *   A("@","1.2.3.4")
 * );
 * ```
 * 
 * If you want to clear the defaults, you can do the following.
 * The domain `example2.com` will **not** have the defaults set.
 * 
 * ```javascript
 * DEFAULTS();
 * 
 * D("example2.com",
 *   REGISTRAR,
 *   DnsProvider("R53"),
 *   A("@","1.2.3.4")
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/defaults
 */
declare function DEFAULTS(...modifiers: DomainModifier[]): void;

/**
 * `DOMAIN_ELSEWHERE()` is a helper macro that lets you easily indicate that
 * a domain's zones are managed elsewhere. That is, it permits you easily delegate
 * a domain to a hard-coded list of DNS servers.
 * 
 * `DOMAIN_ELSEWHERE` is useful when you control a domain's registrar but not the
 * DNS servers. For example, suppose you own a domain but the DNS servers are run
 * by someone else, perhaps a SaaS product you've subscribed to or a DNS server
 * that is run by your brother-in-law who doesn't trust you with the API keys that
 * would let you maintain the domain using DNSControl. You need an easy way to
 * point (delegate) the domain at a specific list of DNS servers.
 * 
 * For example these two statements are equivalent:
 * 
 * ```javascript
 * DOMAIN_ELSEWHERE("example.com", REG_NAMEDOTCOM, ["ns1.foo.com", "ns2.foo.com"]);
 * ```
 * 
 * ```javascript
 * D("example.com", REG_NAMEDOTCOM,
 *     NO_PURGE,
 *     NAMESERVER("ns1.foo.com"),
 *     NAMESERVER("ns2.foo.com")
 * );
 * ```
 * 
 * NOTE: The [`NO_PURGE`](../domain/NO_PURGE.md) is used out of abundance of caution but since no
 * `DnsProvider()` statements exist, no updates would be performed.
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/domain_elsewhere
 */
declare function DOMAIN_ELSEWHERE(name: string, registrar: string, nameserver_names: string[]): void;

/**
 * `DOMAIN_ELSEWHERE_AUTO()` is similar to `DOMAIN_ELSEWHERE()` but instead of
 * a hardcoded list of nameservers, a DnsProvider() is queried.
 * 
 * `DOMAIN_ELSEWHERE_AUTO` is useful when you control a domain's registrar but the
 * DNS zones are managed by another system. Luckily you have enough access to that
 * other system that you can query it to determine the zone's nameservers.
 * 
 * For example, suppose you own a domain but the DNS servers for it are in Azure.
 * Further suppose that something in Azure maintains the zones (automatic or
 * human). Azure picks the nameservers for the domains automatically, and that
 * list may change occasionally.  `DOMAIN_ELSEWHERE_AUTO` allows you to easily
 * query Azure to determine the domain's delegations so that you do not need to
 * hard-code them in your dnsconfig.js file.
 * 
 * For example these two statements are equivalent:
 * 
 * ```javascript
 * DOMAIN_ELSEWHERE_AUTO("example.com", REG_NAMEDOTCOM, DSP_AZURE);
 * ```
 * 
 * ```javascript
 * D("example.com", REG_NAMEDOTCOM,
 *     NO_PURGE,
 *     DnsProvider(DSP_AZURE)
 * );
 * ```
 * 
 * NOTE: The [`NO_PURGE`](../domain/NO_PURGE.md) is used to prevent DNSControl from changing the records.
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/domain_elsewhere_auto
 */
declare function DOMAIN_ELSEWHERE_AUTO(name: string, domain: string, registrar: string, dnsProvider: string): void;

/**
 * `D_EXTEND` adds records (and metadata) to a domain previously defined
 * by [`D()`](D.md). It can also be used to add subdomain records (and metadata)
 * to a previously defined domain.
 * 
 * The first argument is a domain name. If it exactly matches a
 * previously defined domain, `D_EXTEND()` behaves the same as [`D()`](D.md),
 * simply adding records as if they had been specified in the original
 * [`D()`](D.md).
 * 
 * If the domain name does not match an existing domain, but could be a
 * (non-delegated) subdomain of an existing domain, the new records (and
 * metadata) are added with the subdomain part appended to all record
 * names (labels), and targets (as appropriate). See the examples below.
 * 
 * Matching the domain name to previously-defined domains is done using a
 * `longest match` algorithm.  If `domain.tld` and `sub.domain.tld` are
 * defined as separate domains via separate [`D()`](D.md) statements, then
 * `D_EXTEND('sub.sub.domain.tld', ...)` would match `sub.domain.tld`,
 * not `domain.tld`.
 * 
 * Some operators only act on an apex domain (e.g.
 * [`CF_REDIRECT`](../domain/CF_REDIRECT.md) and [`CF_TEMP_REDIRECT`](../domain/CF_TEMP_REDIRECT.md)). Using them
 * in a `D_EXTEND` subdomain may not be what you expect.
 * 
 * ```javascript
 * D("domain.tld", REG, DnsProvider(DNS),
 *   A("@", "127.0.0.1"), // domain.tld
 *   A("www", "127.0.0.2"), // www.domain.tld
 *   CNAME("a", "b") // a.domain.tld -> b.domain.tld
 * );
 * D_EXTEND("domain.tld",
 *   A("aaa", "127.0.0.3"), // aaa.domain.tld
 *   CNAME("c", "d") // c.domain.tld -> d.domain.tld
 * );
 * D_EXTEND("sub.domain.tld",
 *   A("bbb", "127.0.0.4"), // bbb.sub.domain.tld
 *   A("ccc", "127.0.0.5"), // ccc.sub.domain.tld
 *   CNAME("e", "f") // e.sub.domain.tld -> f.sub.domain.tld
 * );
 * D_EXTEND("sub.sub.domain.tld",
 *   A("ddd", "127.0.0.6"), // ddd.sub.sub.domain.tld
 *   CNAME("g", "h") // g.sub.sub.domain.tld -> h.sub.sub.domain.tld
 * );
 * D_EXTEND("sub.domain.tld",
 *   A("@", "127.0.0.7"), // sub.domain.tld
 *   CNAME("i", "j") // i.sub.domain.tld -> j.sub.domain.tld
 * );
 * ```
 * 
 * This will end up in the following modifications: (This output assumes the `--full` flag)
 * 
 * ```text
 * ******************** Domain: domain.tld
 * ----- Getting nameservers from: cloudflare
 * ----- DNS Provider: cloudflare...7 corrections
 * #1: CREATE A aaa.domain.tld 127.0.0.3
 * #2: CREATE A bbb.sub.domain.tld 127.0.0.4
 * #3: CREATE A ccc.sub.domain.tld 127.0.0.5
 * #4: CREATE A ddd.sub.sub.domain.tld 127.0.0.6
 * #5: CREATE A sub.domain.tld 127.0.0.7
 * #6: CREATE A www.domain.tld 127.0.0.2
 * #7: CREATE A domain.tld 127.0.0.1
 * #8: CREATE CNAME a.domain.tld b.domain.tld.
 * #9: CREATE CNAME c.domain.tld d.domain.tld.
 * #10: CREATE CNAME e.sub.domain.tld f.sub.domain.tld.
 * #11: CREATE CNAME g.sub.sub.domain.tld h.sub.sub.domain.tld.
 * #12: CREATE CNAME i.sub.domain.tld j.sub.domain.tld.
 * ```
 * 
 * ProTips: `D_EXTEND()` permits you to create very complex and
 * sophisticated configurations, but you shouldn't. Be nice to the next
 * person that edits the file, who may not be as expert as yourself.
 * Enhance readability by putting any `D_EXTEND()` statements immediately
 * after the original [`D()`](D.md), like in above example.  Avoid the temptation
 * to obscure the addition of records to existing domains with randomly
 * placed `D_EXTEND()` statements. Don't build up a domain using loops of
 * `D_EXTEND()` statements. You'll be glad you didn't.
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/d_extend
 */
declare function D_EXTEND(name: string, ...modifiers: DomainModifier[]): void;

/**
 * Converts an IPv4 address from string to an integer. This allows performing mathematical operations with the IP address.
 * 
 * ```javascript
 * var addrA = IP('1.2.3.4')
 * var addrB = addrA + 1
 * // addrB = 1.2.3.5
 * ```
 * 
 * NOTE: `IP()` does not accept IPv6 addresses (PRs gladly accepted!). IPv6 addresses are simply strings:
 * 
 * ```javascript
 * // IPv4 Var
 * var addrA1 = IP("1.2.3.4");
 * var addrA2 = "1.2.3.4";
 * 
 * // IPv6 Var
 * var addrAAAA = "0:0:0:0:0:0:0:0";
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/ip
 */
declare function IP(ip: string): number;

/**
 * NewDnsProvider activates a DNS Service Provider (DSP) specified in `creds.json`.
 * A DSP stores a DNS zone's records and provides DNS service for the zone (i.e.
 * answers on port 53 to queries related to the zone).
 * 
 * * `name` must match the name of an entry in `creds.json`.
 * * `type` specifies a valid DNS provider type identifier listed on the [provider page](../../providers.md).
 *   * Starting with [v3.16](../../v316.md), the type is optional. If it is absent, the `TYPE` field in `creds.json` is used instead. You can leave it out. (Thanks to JavaScript magic, you can leave it out even when there are more fields).
 *   * Starting with v4.0, specifying the type may be an error. Please add the `TYPE` field to `creds.json` and remove this parameter from `dnsconfig.js` to prepare.
 * * `meta` is a way to send additional parameters to the provider.  It is optional and only certain providers use it.  See the [individual provider docs](../../providers.md) for details.
 * 
 * This function will return an opaque string that should be assigned to a variable name for use in [D](D.md) directives.
 * 
 * Prior to [v3.16](../../v316.md):
 * 
 * ```javascript
 * var REG_MYNDC = NewRegistrar("mynamedotcom", "NAMEDOTCOM");
 * var DNS_MYAWS = NewDnsProvider("myaws", "ROUTE53");
 * 
 * D("example.com", REG_MYNDC, DnsProvider(DNS_MYAWS),
 *   A("@","1.2.3.4")
 * );
 * ```
 * 
 * In [v3.16](../../v316.md) and later:
 * 
 * ```javascript
 * var REG_MYNDC = NewRegistrar("mynamedotcom");
 * var DNS_MYAWS = NewDnsProvider("myaws");
 * 
 * D("example.com", REG_MYNDC, DnsProvider(DNS_MYAWS),
 *   A("@","1.2.3.4")
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/newdnsprovider
 */
declare function NewDnsProvider(name: string, type?: string, meta?: object): string;

/**
 * NewRegistrar activates a Registrar Provider specified in `creds.json`.
 * A registrar maintains the domain's registration and delegation (i.e. the
 * nameservers for the domain).  DNSControl only manages the delegation.
 * 
 * * `name` must match the name of an entry in `creds.json`.
 * * `type` specifies a valid DNS provider type identifier listed on the [provider page](../../providers.md).
 *   * Starting with [v3.16](../../v316.md), the type is optional. If it is absent, the `TYPE` field in `creds.json` is used instead. You can leave it out. (Thanks to JavaScript magic, you can leave it out even when there are more fields).
 *   * Starting with v4.0, specifying the type may be an error. Please add the `TYPE` field to `creds.json` and remove this parameter from `dnsconfig.js` to prepare.
 * * `meta` is a way to send additional parameters to the provider.  It is optional and only certain providers use it.  See the [individual provider docs](../../providers.md) for details.
 * 
 * This function will return an opaque string that should be assigned to a variable name for use in [D](D.md) directives.
 * 
 * Prior to [v3.16](../../v316.md):
 * 
 * ```javascript
 * var REG_MYNDC = NewRegistrar("mynamedotcom", "NAMEDOTCOM");
 * var DNS_MYAWS = NewDnsProvider("myaws", "ROUTE53");
 * 
 * D("example.com", REG_MYNDC, DnsProvider(DNS_MYAWS),
 *   A("@","1.2.3.4")
 * );
 * ```
 * 
 * In [v3.16](../../v316.md) and later:
 * 
 * ```javascript
 * var REG_MYNDC = NewRegistrar("mynamedotcom");
 * var DNS_MYAWS = NewDnsProvider("myaws");
 * 
 * D("example.com", REG_MYNDC, DnsProvider(DNS_MYAWS),
 *   A("@","1.2.3.4")
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/newregistrar
 */
declare function NewRegistrar(name: string, type?: string, meta?: object): string;

/**
 * `PANIC` terminates the script and therefore DNSControl with an exit code of 1. This should be used if your script cannot gather enough information to generate records, for example when a HTTP request failed.
 * 
 * ```javascript
 * PANIC("Something really bad has happened");
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/panic
 */
declare function PANIC(message: string): never;

/**
 * `REV` returns the reverse lookup domain for an IP network. For
 * example `REV('1.2.3.0/24')` returns `3.2.1.in-addr.arpa.` and
 * `REV('2001:db8:302::/48)` returns `2.0.3.0.8.b.d.0.1.0.0.2.ip6.arpa.`.
 * This is used in [`D()`](D.md) functions to create reverse DNS lookup zones.
 * 
 * This is a convenience function. You could specify `D('3.2.1.in-addr.arpa',
 * ...` if you like to do things manually but why would you risk making
 * typos?
 * 
 * `REV` complies with RFC2317, "Classless in-addr.arpa delegation"
 * for netmasks of size /25 through /31.
 * While the RFC permits any format, we abide by the recommended format:
 * `FIRST/MASK.C.B.A.in-addr.arpa` where `FIRST` is the first IP address
 * of the zone, `MASK` is the netmask of the zone (25-31 inclusive),
 * and A, B, C are the first 3 octets of the IP address. For example
 * `172.20.18.130/27` is located in a zone named
 * `128/27.18.20.172.in-addr.arpa`
 * 
 * If the address does not include a "/" then `REV` assumes /32 for IPv4 addresses
 * and /128 for IPv6 addresses.
 * 
 * Note that the lower bits (the ones outside the netmask) must be zeros. They are not
 * zeroed out automatically. Thus, `REV('1.2.3.4/24')` is an error.  This is done
 * to catch typos.
 * 
 * ```javascript
 * D(REV('1.2.3.0/24'), REGISTRAR, DnsProvider(BIND),
 *   PTR("1", 'foo.example.com.'),
 *   PTR("2", 'bar.example.com.'),
 *   PTR("3", 'baz.example.com.'),
 *   // These take advantage of DNSControl's ability to generate the right name:
 *   PTR("1.2.3.10", 'ten.example.com.'),
 * );
 * 
 * D(REV('2001:db8:302::/48'), REGISTRAR, DnsProvider(BIND),
 *   PTR("1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0", 'foo.example.com.'),  // 2001:db8:302::1
 *   // These take advantage of DNSControl's ability to generate the right name:
 *   PTR("2001:db8:302::2", 'two.example.com.'),                          // 2.0.0...
 *   PTR("2001:db8:302::3", 'three.example.com.'),                        // 3.0.0...
 * );
 * ```
 * 
 * In the future we plan on adding a flag to [`A()`](../domain/A.md)which will insert
 * the correct PTR() record in the appropriate `D(REV())` domain (i.e. `.arpa` domain) has been
 * defined.
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/rev
 */
declare function REV(address: string): string;

/**
 * `getConfiguredDomains` getConfiguredDomains is a helper function that returns the domain names
 * configured at the time the function is called. Calling this function early or later in
 * `dnsconfig.js` may return different results. Typical usage is to iterate over all
 * domains at the end of your configuration file.
 * 
 * Example for adding records to all configured domains:
 * ```javascript
 * var domains = getConfiguredDomains();
 * for(i = 0; i < domains.length; i++) {
 *   D_EXTEND(domains[i],
 *     TXT('_important', 'BLA') // I know, not really creative.
 *   )
 * }
 * ```
 * 
 * This will end up in following modifications: (All output assumes the `--full` flag)
 * 
 * ```text
 * ******************** Domain: domain1.tld
 * ----- Getting nameservers from: registrar
 * ----- DNS Provider: registrar...2 corrections
 * #1: CREATE TXT _important.domain1.tld "BLA" ttl=43200
 * #2: REFRESH zone domain1.tld
 * 
 * ******************** Domain: domain2.tld
 * ----- Getting nameservers from: registrar
 * ----- DNS Provider: registrar...2 corrections
 * #1: CREATE TXT _important.domain2.tld "BLA" ttl=43200
 * #2: REFRESH zone domain2.tld
 * ```
 * 
 * Example for adding DMARC report records:
 * 
 * This example might be more useful, specially for configuring the DMARC report records. According to DMARC RFC you need to specify `domain2.tld._report.dmarc.domain1.tld` to allow `domain2.tld` to send aggregate/forensic email reports to `domain1.tld`. This can be used to do this in an easy way, without using the wildcard from the RFC.
 * 
 * ```javascript
 * var domains = getConfiguredDomains();
 * for(i = 0; i < domains.length; i++) {
 *     D_EXTEND("domain1.tld",
 *         TXT(domains[i] + '._report._dmarc', 'v=DMARC1')
 *     );
 * }
 * ```
 * 
 * This will end up in following modifications:
 * 
 * ```text
 * ******************** Domain: domain2.tld
 * ----- Getting nameservers from: registrar
 * ----- DNS Provider: registrar...4 corrections
 * #1: CREATE TXT domain1.tld._report._dmarc.domain2.tld "v=DMARC1" ttl=43200
 * #2: CREATE TXT domain3.tld._report._dmarc.domain2.tld "v=DMARC1" ttl=43200
 * #3: CREATE TXT domain4.tld._report._dmarc.domain2.tld "v=DMARC1" ttl=43200
 * #4: REFRESH zone domain2.tld
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/getconfigureddomains
 */
declare function getConfiguredDomains(): string[];

/**
 * `require_glob()` can recursively load `.js` files, optionally non-recursive as well.
 * 
 * Possible parameters are:
 * 
 * - Path as string, where you would like to start including files. Mandatory. Pattern matching possible, see [GoLand path/filepath/#Match docs](https://golang.org/pkg/path/filepath/#Match).
 * - If being recursive. This is a boolean if the search should be recursive or not. Define either `true` or `false`. Default is `true`.
 * 
 * Example to load `.js` files recursively:
 * 
 * ```javascript
 * require_glob("./domains/");
 * ```
 * 
 * Example to load `.js` files only in `domains/`:
 * 
 * ```javascript
 * require_glob("./domains/", false);
 * ```
 * 
 * One more important thing to note: `require_glob()` is as smart as `require()` is. It loads files always relative to the JavaScript
 * file where it's being executed in. Let's go with an example, as it describes it better:
 * 
 * ```javascript
 * require("domains/index.js");
 * ```
 * 
 * ```javascript
 * require_glob("./user1/");
 * ```
 * 
 * This will now load files being present underneath `./domains/user1/` and **NOT** at below `./domains/`, as `require_glob()`
 * is called in the subfolder `domains/`.
 * 
 * @see https://docs.dnscontrol.org/language-reference/top-level-functions/require_glob
 */
declare function require_glob(path: string, recursive: boolean): void;

/**
 * DNSControl contains a `CAA_BUILDER` which can be used to simply create
 * [`CAA()`](../domain/CAA.md) records for your domains. Instead of creating each [`CAA()`](../domain/CAA.md) record
 * individually, you can simply configure your report mail address, the
 * authorized certificate authorities and the builder cares about the rest.
 * 
 * ## Example
 * 
 * For example you can use:
 * 
 * ```javascript
 * CAA_BUILDER({
 *   label: "@",
 *   iodef: "mailto:test@domain.tld",
 *   iodef_critical: true,
 *   issue: [
 *     "letsencrypt.org",
 *     "comodoca.com",
 *   ],
 *   issuewild: "none",
 * })
 * ```
 * 
 * The parameters are:
 * 
 * * `label:` The label of the CAA record. (Optional. Default: `"@"`)
 * * `iodef:` Report all violation to configured mail address.
 * * `iodef_critical:` This can be `true` or `false`. If enabled and CA does not support this record, then certificate issue will be refused. (Optional. Default: `false`)
 * * `issue:` An array of CAs which are allowed to issue certificates. (Use `"none"` to refuse all CAs)
 * * `issuewild:` An array of CAs which are allowed to issue wildcard certificates. (Can be simply `"none"` to refuse issuing wildcard certificates for all CAs)
 * 
 * `CAA_BUILDER()` returns multiple records (when configured as example above):
 * 
 * ```javascript
 * CAA("@", "iodef", "mailto:test@domain.tld", CAA_CRITICAL)
 * CAA("@", "issue", "letsencrypt.org")
 * CAA("@", "issue", "comodoca.com")
 * CAA("@", "issuewild", ";")
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/record-modifiers/caa_builder
 */
declare function CAA_BUILDER(opts: { label?: string; iodef: string; iodef_critical?: boolean; issue: string[]; issuewild: string }): RecordModifier;

/**
 * DNSControl contains a `DMARC_BUILDER` which can be used to simply create
 * DMARC policies for your domains.
 * 
 * ## Example
 * 
 * ### Simple example
 * 
 * ```javascript
 * DMARC_BUILDER({
 *   policy: 'reject',
 *   ruf: [
 *     'mailto:mailauth-reports@example.com',
 *   ],
 * })
 * ```
 * 
 * This yield the following record:
 * 
 * ```text
 * @   IN  TXT "v=DMARC1; p=reject; ruf=mailto:mailauth-reports@example.com"
 * ```
 * 
 * ### Advanced example
 * 
 * ```javascript
 * DMARC_BUILDER({
 *   policy: 'reject',
 *   subdomainPolicy: 'quarantine',
 *   percent: 50,
 *   alignmentSPF: 'r',
 *   alignmentDKIM: 'strict',
 *   rua: [
 *     'mailto:mailauth-reports@example.com',
 *     'https://dmarc.example.com/submit',
 *   ],
 *   ruf: [
 *     'mailto:mailauth-reports@example.com',
 *   ],
 *   failureOptions: '1',
 *   reportInterval: '1h',
 * });
 * ```
 * 
 * ```javascript
 * DMARC_BUILDER({
 *   label: 'insecure',
 *   policy: 'none',
 *   ruf: [
 *     'mailto:mailauth-reports@example.com',
 *   ],
 *   failureOptions: {
 *       SPF: false,
 *       DKIM: true,
 *   },
 * });
 * ```
 * 
 * This yields the following records:
 * 
 * ```text
 * @           IN  TXT "v=DMARC1; p=reject; sp=quarantine; adkim=s; aspf=r; pct=50; rua=mailto:mailauth-reports@example.com,https://dmarc.example.com/submit; ruf=mailto:mailauth-reports@example.com; fo=1; ri=3600"
 * insecure    IN  TXT "v=DMARC1; p=none; ruf=mailto:mailauth-reports@example.com; fo=d"
 * ```
 * 
 * ### Parameters
 * 
 * * `label:` The DNS label for the DMARC record (`_dmarc` prefix is added, default: `'@'`)
 * * `version:` The DMARC version to be used (default: `DMARC1`)
 * * `policy:` The DMARC policy (`p=`), must be one of `'none'`, `'quarantine'`, `'reject'`
 * * `subdomainPolicy:` The DMARC policy for subdomains (`sp=`), must be one of `'none'`, `'quarantine'`, `'reject'` (optional)
 * * `alignmentSPF:` `'strict'`/`'s'` or `'relaxed'`/`'r'` alignment for SPF (`aspf=`, default: `'r'`)
 * * `alignmentDKIM:` `'strict'`/`'s'` or `'relaxed'`/`'r'` alignment for DKIM (`adkim=`, default: `'r'`)
 * * `percent:` Number between `0` and `100`, percentage for which policies are applied (`pct=`, default: `100`)
 * * `rua:` Array of aggregate report targets (optional)
 * * `ruf:` Array of failure report targets (optional)
 * * `failureOptions:` Object or string; Object containing booleans `SPF` and `DKIM`, string is passed raw (`fo=`, default: `'0'`)
 * * `failureFormat:` Format in which failure reports are requested (`rf=`, default: `'afrf'`)
 * * `reportInterval:` Interval in which reports are requested (`ri=`)
 * * `ttl:` Input for `TTL` method (optional)
 * 
 * ### Caveats
 * 
 * * TXT records are automatically split using `AUTOSPLIT`.
 * * URIs in the `rua` and `ruf` arrays are passed raw. You must percent-encode all commas and exclamation points in the URI itself.
 * 
 * @see https://docs.dnscontrol.org/language-reference/record-modifiers/dmarc_builder
 */
declare function DMARC_BUILDER(opts: { label?: string; version?: string; policy: 'none' | 'quarantine' | 'reject'; subdomainPolicy?: 'none' | 'quarantine' | 'reject'; alignmentSPF?: 'strict' | 's' | 'relaxed' | 'r'; alignmentDKIM?: 'strict' | 's' | 'relaxed' | 'r'; percent?: number; rua?: string[]; ruf?: string[]; failureOptions?: { SPF: boolean, DKIM: boolean } | string; failureFormat?: string; reportInterval?: Duration; ttl?: Duration }): RecordModifier;

/**
 * `LOC_BUILDER_DD({})` actually takes an object with the following properties:
 * 
 *   - label (optional, defaults to `@`)
 *   - x (float32)
 *   - y (float32)
 *   - alt (float32, optional)
 *   - ttl (optional)
 * 
 * A helper to build [`LOC`](../domain/LOC.md) records. Supply four parameters instead of 12.
 * 
 * Internally assumes some defaults for [`LOC`](../domain/LOC.md) records.
 * 
 * The cartesian coordinates are decimal degrees, like you typically find in e.g. Google Maps.
 * 
 * Examples.
 * 
 * Big Ben:
 * `51.50084265331501, -0.12462541415599787`
 * 
 * The White House:
 * `38.89775977858357, -77.03655125982903`
 * 
 * ```javascript
 * D("example.com","none"
 *   , LOC_BUILDER_DD({
 *     label: "big-ben",
 *     x: 51.50084265331501,
 *     y: -0.12462541415599787,
 *     alt: 6,
 *   })
 *   , LOC_BUILDER_DD({
 *     label: "white-house",
 *     x: 38.89775977858357,
 *     y: -77.03655125982903,
 *     alt: 19,
 *   })
 *   , LOC_BUILDER_DD({
 *     label: "white-house-ttl",
 *     x: 38.89775977858357,
 *     y: -77.03655125982903,
 *     alt: 19,
 *     ttl: "5m",
 *   })
 * );
 * 
 * ```
 * 
 * Part of the series:
 *  * [`LOC()`](../domain/LOC.md) - build a `LOC` by supplying all 12 parameters
 *  * [`LOC_BUILDER_DD({})`](../record/LOC_BUILDER_DD.md) - accepts cartesian x, y
 *  * [`LOC_BUILDER_DMS_STR({})`](../record/LOC_BUILDER_DMS_STR.md) - accepts DMS 33°51′31″S 151°12′51″E
 *  * [`LOC_BUILDER_DMM_STR({})`](../record/LOC_BUILDER_DMM_STR.md) - accepts DMM 25.24°S 153.15°E
 *  * [`LOC_BUILDER_STR({})`](../record/LOC_BUILDER_STR.md) - tries the cooordinate string in all `LOC_BUILDER_DM*_STR()` functions until one works
 * 
 * @see https://docs.dnscontrol.org/language-reference/record-modifiers/loc_builder_dd
 */
declare function LOC_BUILDER_DD(opts: { label?: string; x: number; y: number; alt?: number; ttl?: Duration }): RecordModifier;

/**
 * `LOC_BUILDER_DMM({})` actually takes an object with the following properties:
 * 
 *   - label (string, optional, defaults to `@`)
 *   - str (string)
 *   - alt (float32, optional)
 *   - ttl (optional)
 * 
 * A helper to build [`LOC`](../domain/LOC.md) records. Supply three parameters instead of 12.
 * 
 * Internally assumes some defaults for [`LOC`](../domain/LOC.md) records.
 * 
 * Accepts a string with decimal minutes (DMM) coordinates in the form: 25.24°S 153.15°E
 * 
 * Note that the following are acceptable forms (symbols differ):
 * * `25.24°S 153.15°E`
 * * `25.24 S 153.15 E`
 * * `25.24° S 153.15° E`
 * * `25.24S 153.15E`
 * 
 * ```javascript
 * D("example.com","none"
 *   LOC_BUILDER_STR({
 *     label: "tasmania",
 *     str: '42°S 147°E',
 *     alt: 3,
 *   })
 * );
 * 
 * ```
 * 
 * Part of the series:
 *  * [`LOC()`](../domain/LOC.md) - build a `LOC` by supplying all 12 parameters
 *  * [`LOC_BUILDER_DD({})`](../record/LOC_BUILDER_DD.md) - accepts cartesian x, y
 *  * [`LOC_BUILDER_DMS_STR({})`](../record/LOC_BUILDER_DMS_STR.md) - accepts DMS 33°51′31″S 151°12′51″E
 *  * [`LOC_BUILDER_DMM_STR({})`](../record/LOC_BUILDER_DMM_STR.md) - accepts DMM 25.24°S 153.15°E
 *  * [`LOC_BUILDER_STR({})`](../record/LOC_BUILDER_STR.md) - tries the cooordinate string in all `LOC_BUILDER_DM*_STR()` functions until one works
 * 
 * @see https://docs.dnscontrol.org/language-reference/record-modifiers/loc_builder_dmm_str
 */
declare function LOC_BUILDER_DMM_STR(opts: { label?: string; str: string; alt?: number; ttl?: Duration }): RecordModifier;

/**
 * `LOC_BUILDER_DMS_STR({})` actually takes an object with the following properties:
 * 
 *   - label (string, optional, defaults to `@`)
 *   - str (string)
 *   - alt (float32, optional)
 *   - ttl (optional)
 * 
 * A helper to build [`LOC`](../domain/LOC.md) records. Supply three parameters instead of 12.
 * 
 * Internally assumes some defaults for [`LOC`](../domain/LOC.md) records.
 * 
 * Accepts a string with degrees, minutes, and seconds (DMS) coordinates in the form: 41°24'12.2"N 2°10'26.5"E
 * 
 * Note that the following are acceptable forms (symbols differ):
 * * `33°51′31″S 151°12′51″E`
 * * `33°51'31"S 151°12'51"E`
 * * `33d51m31sS 151d12m51sE`
 * * `33d51m31s S 151d12m51s E`
 * 
 * ```javascript
 * D("example.com","none"
 *   LOC_BUILDER_DMS_STR({
 *     label: "sydney-opera-house",
 *     str: '33°51′31″S 151°12′51″E',
 *     alt: 4,
 *     ttl: "5m",
 *   })
 * );
 * 
 * ```
 * 
 * Part of the series:
 *  * [`LOC()`](../domain/LOC.md) - build a `LOC` by supplying all 12 parameters
 *  * [`LOC_BUILDER_DD({})`](../record/LOC_BUILDER_DD.md) - accepts cartesian x, y
 *  * [`LOC_BUILDER_DMS_STR({})`](../record/LOC_BUILDER_DMS_STR.md) - accepts DMS 33°51′31″S 151°12′51″E
 *  * [`LOC_BUILDER_DMM_STR({})`](../record/LOC_BUILDER_DMM_STR.md) - accepts DMM 25.24°S 153.15°E
 *  * [`LOC_BUILDER_STR({})`](../record/LOC_BUILDER_STR.md) - tries the cooordinate string in all `LOC_BUILDER_DM*_STR()` functions until one works
 * 
 * @see https://docs.dnscontrol.org/language-reference/record-modifiers/loc_builder_dms_str
 */
declare function LOC_BUILDER_DMS_STR(opts: { label?: string; str: string; alt?: number; ttl?: Duration }): RecordModifier;

/**
 * `LOC_BUILDER_STR({})` actually takes an object with the following: properties.
 * 
 *   - label (optional, defaults to `@`)
 *   - str (string)
 *   - alt (float32, optional)
 *   - ttl (optional)
 * 
 * A helper to build [`LOC`](../domain/LOC.md) records. Supply three parameters instead of 12.
 * 
 * Internally assumes some defaults for [`LOC`](../domain/LOC.md) records.
 * 
 * Accepts a string and tries all `LOC_BUILDER_DM*_STR({})` methods:
 *  * [`LOC_BUILDER_DMS_STR({})`](../record/LOC_BUILDER_DMS_STR.md) - accepts DMS 33°51′31″S 151°12′51″E
 *  * [`LOC_BUILDER_DMM_STR({})`](../record/LOC_BUILDER_DMM_STR.md) - accepts DMM 25.24°S 153.15°E
 * 
 * ```javascript
 * D("example.com","none"
 *   , LOC_BUILDER_STR({
 *     label: "old-faithful",
 *     str: '44.46046°N 110.82815°W',
 *     alt: 2240,
 *   })
 *   , LOC_BUILDER_STR({
 *     label: "ribblehead-viaduct",
 *     str: '54.210436°N 2.370231°W',
 *     alt: 300,
 *   })
 *   , LOC_BUILDER_STR({
 *     label: "guinness-brewery",
 *     str: '53°20′40″N 6°17′20″W',
 *     alt: 300,
 *   })
 * );
 * 
 * ```
 * 
 * Part of the series:
 *  * [`LOC()`](../domain/LOC.md) - build a `LOC` by supplying all 12 parameters
 *  * [`LOC_BUILDER_DD({})`](../record/LOC_BUILDER_DD.md) - accepts cartesian x, y
 *  * [`LOC_BUILDER_DMS_STR({})`](../record/LOC_BUILDER_DMS_STR.md) - accepts DMS 33°51′31″S 151°12′51″E
 *  * [`LOC_BUILDER_DMM_STR({})`](../record/LOC_BUILDER_DMM_STR.md) - accepts DMM 25.24°S 153.15°E
 *  * [`LOC_BUILDER_STR({})`](../record/LOC_BUILDER_STR.md) - tries the cooordinate string in all `LOC_BUILDER_DM*_STR()` functions until one works
 * 
 * @see https://docs.dnscontrol.org/language-reference/record-modifiers/loc_builder_str
 */
declare function LOC_BUILDER_STR(opts: { label?: string; str: string; alt?: number; ttl?: Duration }): RecordModifier;

/**
 * DNSControl offers a `M365_BUILDER` which can be used to simply set up Microsoft 365 for a domain in an opinionated way.
 * 
 * It defaults to a setup without support for legacy Skype for Business applications.
 * It doesn't set up SPF or DMARC. See [`SPF_BUILDER`](/language-reference/record-modifiers/dmarc_builder) and [`DMARC_BUILDER`](/language-reference/record-modifiers/spf_builder).
 * 
 * ## Example
 * 
 * ### Simple example
 * 
 * ```javascript
 * M365_BUILDER({
 *     initialDomain: 'example.onmicrosoft.com',
 * });
 * ```
 * 
 * This sets up `MX` records, Autodiscover, and DKIM.
 * 
 * ### Advanced example
 * 
 * ```javascript
 * M365_BUILDER({
 *     label: 'test',
 *     mx: false,
 *     autodiscover: false,
 *     dkim: false,
 *     mdm: true,
 *     domainGUID: 'test-example-com', // Can be automatically derived in this case, if example.com is the context.
 *     initialDomain: 'example.onmicrosoft.com',
 * });
 * ```
 * 
 * This sets up Mobile Device Management only.
 * 
 * ### Parameters
 * 
 * * `label` The label of the Microsoft 365 domain, useful if it is a subdomain (default: `'@'`)
 * * `mx` Set an `MX` record? (default: `true`)
 * * `autodiscover` Set Autodiscover `CNAME` record? (default: `true`)
 * * `dkim` Set DKIM `CNAME` records? (default: `true`)
 * * `skypeForBusiness` Set Skype for Business/Microsoft Teams records? (default: `false`)
 * * `mdm` Set Mobile Device Management records? (default: `false`)
 * * `domainGUID` The GUID of _this_ Microsoft 365 domain (default: `<label>.<context>` with `.` replaced by `-`, no default if domain contains dashes)
 * * `initialDomain` The initial domain of your Microsoft 365 tenant/account, ends in `onmicrosoft.com`
 * 
 * @see https://docs.dnscontrol.org/language-reference/record-modifiers/m365_builder
 */
declare function M365_BUILDER(opts: { label?: string; mx?: boolean; autodiscover?: boolean; dkim?: boolean; skypeForBusiness?: boolean; mdm?: boolean; domainGUID?: string; initialDomain?: string }): RecordModifier;

/**
 * R53_ZONE lets you specify the AWS Zone ID for an entire domain (D()) or a specific R53_ALIAS() record.
 * 
 * When used with D(), it sets the zone id of the domain. This can be used to differentiate between split horizon domains in public and private zones.
 * 
 * When used with R53_ALIAS() it sets the required Route53 hosted zone id in a R53_ALIAS record. See [R53_ALIAS's documentation](../domain/R53_ALIAS.md) for details.
 * 
 * @see https://docs.dnscontrol.org/language-reference/record-modifiers/service-provider-specific/amazon-route-53/r53_zone
 */
declare function R53_ZONE(zone_id: string): DomainModifier & RecordModifier;

/**
 * DNSControl can optimize the SPF settings on a domain by flattening
 * (inlining) includes and removing duplicates. DNSControl also makes
 * it easier to document your SPF configuration.
 * 
 * WARNING: Flattening SPF includes is risky.  Only flatten an SPF
 * setting if it is absolutely needed to bring the number of "lookups"
 * to be less than 10. In fact, it is debatable whether or not ISPs
 * enforce the "10 lookup rule".
 * 
 * ## The old way
 * 
 * Here is an example of how SPF settings are normally done:
 * 
 * ```javascript
 * D("example.tld", REG, DNS, ...
 *   TXT("v=spf1 ip4:198.252.206.0/24 ip4:192.111.0.0/24 include:_spf.google.com include:mailgun.org include:spf-basic.fogcreek.com include:mail.zendesk.com include:servers.mcsv.net include:sendgrid.net include:450622.spf05.hubspotemail.net ~all")
 * )
 * ```
 * 
 * This has a few problems:
 * 
 * * No comments. It is difficult to add a comment. In particular, we want to be able to list which ticket requested each item in the SPF setting so that history is retained.
 * * Ugly diffs.  If you add an element to the SPF setting, the diff will show the entire line changed, which is difficult to read.
 * * Too many lookups. The SPF RFC says that SPF settings should not require more than 10 DNS lookups. If we manually flatten (i.e. "inline") an include, we have to remember to check back to see if the settings have changed. Humans are not good at that kind of thing.
 * 
 * ## The DNSControl way
 * 
 * ```javascript
 * D("example.tld", REG, DSP, ...
 *   A("@", "10.2.2.2"),
 *   MX("@", "example.tld."),
 *   SPF_BUILDER({
 *     label: "@",
 *     overflow: "_spf%d",
 *     raw: "_rawspf",
 *     ttl: "5m",
 *     parts: [
 *       "v=spf1",
 *       "ip4:198.252.206.0/24", // ny-mail*
 *       "ip4:192.111.0.0/24", // co-mail*
 *       "include:_spf.google.com", // GSuite
 *       "include:mailgun.org", // Greenhouse.io
 *       "include:spf-basic.fogcreek.com", // Fogbugz
 *       "include:mail.zendesk.com", // Zenddesk
 *       "include:servers.mcsv.net", // MailChimp
 *       "include:sendgrid.net", // SendGrid
 *       "include:450622.spf05.hubspotemail.net", // Hubspot (Ticket# SREREQ-107)
 *       "~all"
 *     ],
 *     flatten: [
 *       "spf-basic.fogcreek.com", // Rationale: Being deprecated. Low risk if it breaks.
 *       "450622.spf05.hubspotemail.net" // Rationale: Unlikely to change without warning.
 *     ]
 *   }),
 * );
 * ```
 * 
 * By using the `SPF_BUILDER()` we gain many benefits:
 * 
 * * Comments can appear next to the element they refer to.
 * * Diffs will be shorter and more specific; therefore easier to read.
 * * Automatic flattening.  We can specify which includes should be flattened and DNSControl will do the work. It will even warn us if the includes change.
 * 
 * ## Syntax
 * 
 * When you want to specify SPF settings for a domain, use the
 * `SPF_BUILDER()` function.
 * 
 * ```javascript
 * D("example.tld", REG, DSP, ...
 *   ...
 *   ...
 *   ...
 *   SPF_BUILDER({
 *     label: "@",
 *     overflow: "_spf%d",  // Delete this line if you don't want big strings split.
 *     overhead1: "20",  // There are 20 bytes of other TXT records on this domain.  Compensate for this.
 *     raw: "_rawspf",  // Delete this line if the default is sufficient.
 *     parts: [
 *       "v=spf1",
 *       // fill in your SPF items here
 *       "~all"
 *     ],
 *     flatten: [
 *       // fill in any domains to inline.
 *     ]
 *   }),
 *   ...
 *   ...
 * );
 * ```
 * 
 * The parameters are:
 * 
 * * `label:` The label of the first TXT record. (Optional. Default: `"@"`)
 * * `overflow:` If set, SPF strings longer than 255 chars will be split into multiple TXT records. The value of this setting determines the template for what the additional labels will be named. If not set, no splitting will occur and DNSControl may generate TXT strings that are too long.
 * * `overhead1:` "Overhead for the 1st TXT record".  When calculating the max length of each TXT record, reduce the maximum for the first TXT record in the chain by this amount.
 * * `raw:` The label of the unaltered SPF settings. Setting to an empty string `''` will disable this. (Optional. Default: `"_rawspf"`)
 * * `ttl:` This allows setting a specific TTL on this SPF record. (Optional. Default: using default record TTL)
 * * `txtMaxSize` The maximum size for each TXT record. Values over 255 will result in [multiple strings][multi-string]. General recommendation is to [not go higher than 450][record-size] so that DNS responses will still fit in a UDP packet. (Optional. Default: `"255"`)
 * * `parts:` The individual parts of the SPF settings.
 * * `flatten:` Which includes should be inlined. For safety purposes the flattening is done on an opt-in basis. If `"*"` is listed, all includes will be flattened... this might create more problems than is solves due to length limitations.
 * 
 * [multi-string]: https://tools.ietf.org/html/rfc4408#section-3.1.3
 * [record-size]: https://tools.ietf.org/html/rfc4408#section-3.1.4
 * 
 * `SPF_BUILDER()` returns multiple `TXT()` records:
 * 
 *   * `TXT("@", "v=spf1 .... ~all")`
 *     *  This is the optimized configuration.
 *   * `TXT("_spf1", "...")`
 *     * If the optimizer needs to split a long string across multiple TXT records, the additional TXT records will have labels `_spf1`, `_spf2`, `_spf3`, etc.
 *   * `TXT("_rawspf", "v=spf1 .... ~all")`
 *     * This is the unaltered SPF configuration. This is purely for debugging purposes and is not used by any email or anti-spam system.  It is only generated if flattening is requested.
 * 
 * We recommend first using this without any flattening. Make sure
 * `dnscontrol preview` works as expected. Once that is done, add the
 * flattening required to reduce the number of lookups to 10 or less.
 * 
 * To count the number of lookups, you can use our interactive SPF
 * debugger at [https://stackexchange.github.io/dnscontrol/flattener/index.html](https://stackexchange.github.io/dnscontrol/flattener/index.html)
 * 
 * # The first in a chain is special
 * 
 * When generating the chain of SPF
 * records, each one is max length 255.  For the first item in
 * the chain, the max is 255 - "overhead1".  Setting this to 255 or
 * higher has undefined behavior.
 * 
 * Why is this useful?
 * 
 * Some sites desire having all DNS queries fit in a single packet so
 * that UDP, not TCP, can be used to satisfy all requests. That means all
 * responses have to be relatively small.
 * 
 * When an SPF system does a "TXT" lookup, it gets SPF and non-SPF
 * records.  This makes the first link in the chain extra large.
 * 
 * The bottom line is that if you want the TXT records to fit in a UDP
 * packet, keep increasing the value of `overhead1` until the packet
 * is no longer truncated.
 * 
 * Example:
 * 
 * ```shell
 * dig +short whatexit.org txt | wc -c
 *    118
 * ```
 * 
 * Setting `overhead1` to 118 should be sufficient.
 * 
 * ```shell
 * dig +short stackoverflow.com txt | wc -c
 *      582
 * ```
 * 
 * Since 582 is bigger than 255, it might not be possible to achieve the
 * goal.  Any value larger than 255 will disable all flattening.  Try
 * 170, then 180, 190 until you get the desired results.
 * 
 * A validator such as
 * [https://www.kitterman.com/spf/validate.html](https://www.kitterman.com/spf/validate.html)
 * will tell you if the queries are being truncated and TCP was required
 * to get the entire record. (Sadly it caches heavily.)
 * 
 * ## Notes about the `spfcache.json`
 * 
 * DNSControl keeps a cache of the DNS lookups performed during
 * optimization.  The cache is maintained so that the optimizer does
 * not produce different results depending on the ups and downs of
 * other people's DNS servers. This makes it possible to do `dnscontrol
 * push` even if your or third-party DNS servers are down.
 * 
 * The DNS cache is kept in a file called `spfcache.json`. If it needs
 * to be updated, the proper data will be written to a file called
 * `spfcache.updated.json` and instructions such as the ones below
 * will be output telling you exactly what to do:
 * 
 * ```shell
 * dnscontrol preview
 * 1 Validation errors:
 * WARNING: 2 spf record lookups are out of date with cache (_spf.google.com,_netblocks3.google.com).
 * Wrote changes to spfcache.updated.json. Please rename and commit:
 *     $ mv spfcache.updated.json spfcache.json
 *     $ git commit spfcache.json
 * ```
 * 
 * In this case, you are being asked to replace `spfcache.json` with
 * the newly generated data in `spfcache.updated.json`.
 * 
 * Needing to do this kind of update is considered a validation error
 * and will block `dnscontrol push` from running.
 * 
 * Note: The instructions are hardcoded strings. The filenames will
 * not change.
 * 
 * Note: The instructions assume you use git. If you use something
 * else, please do the appropriate equivalent command.
 * 
 * ## Caveats
 * 
 * 1. DNSControl 'gives up' if it sees SPF records it can't understand.
 * This includes: syntax errors, features that our spflib doesn't know
 * about, overly complex SPF settings, and anything else that we we
 * didn't feel like implementing.
 * 
 * 2. The TXT record that is generated may exceed DNS limits.  dnscontrol
 * will not generate a single TXT record that exceeds DNS limits, but
 * it ignores the fact that there may be other TXT records on the same
 * label.  For example, suppose it generates a TXT record on the bare
 * domain (stackoverflow.com) that is 250 bytes long. That's fine and
 * doesn't require a continuation record.  However if there is another
 * TXT record (not an SPF record, perhaps a TXT record used to verify
 * domain ownership), the total packet size of all the TXT records
 * could exceed 512 bytes, and will require EDNS or a TCP request.
 * 
 * 3. DNSControl does not warn if the number of lookups exceeds 10.
 * We hope to implement this some day.
 * 
 * 4. The `redirect=` directive is only partially implemented.  We only
 * handle the case where redirect is the last item in the SPF record.
 * In which case, it is equivalent to `include:`.
 * 
 * ## Advanced Technique: Interactive SPF Debugger
 * 
 * DNSControl includes an experimental system for viewing
 * SPF settings:
 * 
 * [https://stackexchange.github.io/dnscontrol/flattener/index.html](https://stackexchange.github.io/dnscontrol/flattener/index.html)
 * 
 * You can also run this locally (it is self-contained) by opening
 * `dnscontrol/docs/flattener/index.html` in your browser.
 * 
 * You can use this to determine the minimal number of domains you
 * need to flatten to have fewer than 10 lookups.
 * 
 * The output is as follows:
 * 
 * 1. The top part lists the domain as it current is configured, how
 * many lookups it requires, and includes a checkbox for each item
 * that could be flattened.
 * 
 * 2. Fully flattened: This section shows the SPF configuration if you
 * fully flatten it. i.e. This is what it would look like if all the
 * checkboxes were checked. Note that this result is likely to be
 * longer than 255 bytes, the limit for a single TXT string.
 * 
 * 3. Fully flattened split: This takes the "fully flattened" result
 * and splits it into multiple DNS records.  To continue to the next
 * record an include is added.
 * 
 * ## Advanced Technique: Define once, use many
 * 
 * In some situations we define an SPF setting once and want to re-use
 * it on many domains. Here's how to do this:
 * 
 * ```javascript
 * var SPF_MYSETTINGS = SPF_BUILDER({
 *   label: "@",
 *   overflow: "_spf%d",
 *   raw: "_rawspf",
 *   parts: [
 *     "v=spf1",
 *     ...
 *     "~all"
 *   ],
 *   flatten: [
 *     ...
 *   ]
 * });
 * 
 * D("example.tld", REG, DSP, ...
 *     SPF_MYSETTINGS
 * );
 * 
 * D("example2.tld", REG, DSP, ...
 *      SPF_MYSETTINGS
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/record-modifiers/spf_builder
 */
declare function SPF_BUILDER(opts: { label?: string; overflow?: string; overhead1?: string; raw?: string; ttl?: Duration; txtMaxSize: string[]; parts?: number; flatten?: string[] }): RecordModifier;

/**
 * TTL sets the TTL for a single record only. This will take precedence
 * over the domain's [DefaultTTL](../domain/DefaultTTL.md) if supplied.
 * 
 * The value can be:
 * 
 *   * An integer (number of seconds). Example: `600`
 *   * A string: Integer with single-letter unit: Example: `5m`
 *   * The unit denotes:
 *     * s (seconds)
 *     * m (minutes)
 *     * h (hours)
 *     * d (days)
 *     * w (weeks)
 *     * n (nonths) (30 days in a nonth)
 *     * y (years) (If you set a TTL to a year, we assume you also do crossword puzzles in pen. Show off!)
 *     * If no unit is specified, the default is seconds.
 *   * We highly recommend using units instead of the number of seconds. Would your coworkers understand your intention better if you wrote `14400` or `'4h'`?
 * 
 * ```javascript
 * D('example.com', REGISTRAR, DnsProvider('R53'),
 *   DefaultTTL(2000),
 *   A('@','1.2.3.4'), // uses default
 *   A('foo', '2.3.4.5', TTL(500)), // overrides default
 *   A('demo1', '3.4.5.11', TTL('5d')),  // 5 days
 *   A('demo2', '3.4.5.12', TTL('5w')),  // 5 weeks
 * );
 * ```
 * 
 * @see https://docs.dnscontrol.org/language-reference/record-modifiers/ttl
 */
declare function TTL(ttl: Duration): RecordModifier;


# flarectl

A CLI application for interacting with a Cloudflare account. Powered by [cloudflare-go](https://github.com/cloudflare/cloudflare-go).

## Installation 

Install it when you install our command-line library:

```sh
go install github.com/cloudflare/cloudflare-go/cmd/flarectl@latest
```

# Usage

You must authenticate with Cloudflare using either an API Token or API Key.

To use an API Token, set the `CF_API_TOKEN` environment variable:

```
$ export CF_API_TOKEN=Abc123Xyz
```

To use an API Key, set the `CF_API_KEY` and `CF_API_EMAIL` environment variables:

```
$ export CF_API_KEY=abcdef1234567890
$ export CF_API_EMAIL=someone@example.com
```

Once authenticated, you can run flarectl commands:

```
$ flarectl:

   flarectl - Cloudflare CLI

USAGE:
   flarectl [global options] command [command options] [arguments...]
   
VERSION:
   2017.10.0
   
COMMANDS:
   ips, i                     Print Cloudflare IP ranges
   user, u                    User information
   zone, z                    Zone information
   dns, d                     DNS records
   user-agents, ua            User-Agent blocking
   pagerules, p               Page Rules
   railgun, r                 Railgun information
   firewall, f                Firewall
   origin-ca-root-cert, ocrc  Print Origin CA Root Certificate (in PEM format)
   help, h                    Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --help, -h		show help
   --version, -v	print the version
   
```

## Examples

## Block an IP via the IP Firewall

```sh
flarectl firewall rules create --zone="example.com" --value="8.8.8.8" --mode="block" --notes="Block bad actor"

ID                               Value   Scope Mode  Notes
-------------------------------- ------- ----- ----- ----------------
7bc6fa4569f78777039ef5ebd7b4cedd 8.8.8.8 zone  block Block bad actor
```

### List Firewall Rules

```sh
~ flarectl firewall rules list

ID                               Value           Scope Mode      Notes 
-------------------------------- --------------- ----- --------- ----- 
210173b610198c8ce3dfe39987e4df78 8.8.8.8         user  whitelist       
36e86aebff4cb8cb2020e622c2ff2b90 8.8.4.4         user  whitelist       
ba6bea6e646e2d453c394a41c6ab931a 45.55.2.6       user  whitelist       
edff311e3f81b35e9cd64e4fa9d18465 45.55.2.5       user  whitelist       
```

### Challenge All Requests for a specific User-Agent

```
~ flarectl ua create --zone="example.com" --mode="challenge" --description="Challenge Chrome v61" --value="Mozilla/5.0 (Macintosh Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML like Gecko) Chrome/61.0.3163.100 Safari/537.36"

ID                               Description          Mode      Value                                                                                                                 Paused 
-------------------------------- -------------------- --------- --------------------------------------------------------------------------------------------------------------------- ------ 
a23b50de3c064a5a860e8b84cd2b382c Challenge Chrome v61 challenge Mozilla/5.0 (Macintosh Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML like Gecko) Chrome/61.0.3163.100 Safari/537.36 false  
```

### Add a DNS record

```sh
~ flarectl dns create --zone="example.com" --name="app" --type="CNAME" --content="myapp.herokuapp.com" --proxy

ID                               Name                      Type  Content             TTL Proxiable Proxy
-------------------------------- ------------------------- ----- ------------------- --- --------- -----
5c5d051f7944cf4715127270dd4d05f4 app.questionable.services CNAME myapp.herokuapp.com 1   true      true
```

## License

BSD licensed. See the [LICENSE](LICENSE) file for details.

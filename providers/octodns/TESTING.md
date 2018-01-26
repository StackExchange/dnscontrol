# Testing:

## Code tests

Unit tests:

```
cd ~/src/github.com/StackExchange/dnscontrol/providers/octodns/octoyaml
go test -v
```

Integration tests:

```
cd ~/src/github.com/StackExchange/dnscontrol/integrationTest
go test -v -verbose -provider OCTODNS
```

## Test against OctoDNS-Validate

### Download OctoDNS:

```
mkdir dns
cd dns
virtualenv env
source env/bin/activate
pip install octodns
ln -s ~/gitwork/fakeroot/ExternalDNS/config config
```

### Modify dnsconfig.js

Update dnsconfig.js to have some OCTODNS zones. We did it this way:

Add:

```
var OCT = NewDnsProvider("octodns", "OCTODNS");
```

Add:

```
 DEFAULTS(
   DnsProvider(SERVERFAULT, 0),
+  DnsProvider(OCT, 0),
   { ns_ttl: "172800" },
 END);
```

Add:

```
 var NO_BIND = function(d) {
   delete d.dnsProviders[SERVERFAULT];
+  delete d.dnsProviders[OCT];
 };
```

## Run the tests:

### Step 1: Setup

```
export ODIR=~/gitwork/octodns/dns
export DDIR=~/src/github.com/StackExchange/dnscontrol
export RDIR=~/gitwork/fakeroot/ExternalDNS
```

(`$RDIR` is the location of dnsconfig.js)

### Step 1: Generate the files

This builds the software then generates the yaml files in the config directory:

```
(cd $DDIR && go install ) && cd $RDIR && rm -f config/*.yaml && dnscontrol push -providers=octodns
```

### Step 2: Run the validator:

This runs octodns-validate against the YAMl files we generated.  production.yaml should
list each domain.

```
cp $DDIR/providers/octodns/testdata/production.yaml config/. && env/bin/octodns-validate --log-stream-stdout
```

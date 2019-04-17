# Testing:

## Create the environment.

These variables are used in all other sections of this doc.

```
export DNSCONFIGDIR=~/gitwork/fakeroot/ExternalDNS
export OCTCONFIGDIR=~/gitwork/octodns/dns
export SRCDIR=~/src/github.com/StackExchange/dnscontrol
```

## Code tests

Unit tests:

```
cd $SRCDIR/providers/octodns/octoyaml
go test -v
```

Integration tests:

```
cd $SRCDIR/integrationTest
go test -v -verbose -provider OCTODNS
```

## Test against OctoDNS-Validate

### Download OctoDNS:

```
cd $DNSCONFIGDIR
mkdir dns
cd dns
virtualenv env
source env/bin/activate
pip install octodns
ln -s ~/gitwork/fakeroot/ExternalDNS/config config
```

### Modify dnsconfig.js

Make a copy of dnsconfig.js and modify it to use OCTODNS as a provider. We did it this way:

```
cd $DNSCONFIGDIR/dns
cp ../dnsconfig.js .
cp ../creds.json .
```

Add:

```
var OCT = NewDnsProvider("octodns", "OCTODNS");
```

Add:

```
 DEFAULTS(
   DnsProvider(SERVERFAULT, 0),
+  DnsProvider(OCT, 0),
   NAMESERVER_TTL("2d"),
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

### Step 1: Generate the files

This builds the software then generates the yaml files in the config directory:

```
(cd $SRCDIR && go install ) && cd $DNSCONFIGDIR/dns && rm -f config/*.yaml && dnscontrol push -providers=octodns
```

### Step 2: Run the validator:

This runs octodns-validate against the YAMl files we generated.  production.yaml should
list each domain.

We create production.yaml like this:

```
cd $DNSCONFIGDIR/dns && $SRCDIR/providers/octodns/mkprodyaml.sh
```

Now we can run the validation:

```
cd $DNSCONFIGDIR/dns
cp $SRCDIR/providers/octodns/testdata/production.yaml config/. && env/bin/octodns-validate --log-stream-stdout
```

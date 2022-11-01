var PROVIDER_NONE = NewRegistrar('none');
var PROVIDER_TRANSIP = NewDnsProvider('transip');

require_glob(
    './domains/',
    false
);

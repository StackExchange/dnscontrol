require('pkg/js/parse_tests/complexImports/b/d/d.js');

function b() {
    return [
        d(),
        CNAME("B", "foo.com.")
    ];
}

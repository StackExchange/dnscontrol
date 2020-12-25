require('./a/c/c.js');
require('./b/b.js');

D("foo.com","none",
    A("@","1.2.3.4"),
    c(),
    b()
);

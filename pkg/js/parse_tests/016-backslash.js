dmarc = [
    "v=DMARC1\\;",
    'p=reject\\;',
    'sp=reject\\;',
    'pct=100\\;',
    'rua=mailto:xx...@yyyy.com\\;',
    'ruf=mailto:xx...@yyyy.com\\;',
    'fo=1'
  ].join(' ');

D("foo.com","none",
    TXT('_dmarc', dmarc, TTL(300))
);

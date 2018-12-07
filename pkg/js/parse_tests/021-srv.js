D("foo.com","none"
  , SRV('_ntp._udp', 1, 100, 123, 'one.foo.com.')
  , SRV('_ntp._udp', 2, 100, 123, 'two')
  , SRV('_ntp._udp', 3, 100, 123, 'localhost')
  , SRV('_ntp._udp', 4, 100, 123, 'three.example.com.')
  , SRV('_ntp._udp', 0, 0, 1, 'zeros')
);

This is from https://godoc.org/bitbucket.org/zombiezen/cardcpx/natsort

We modified it to be compatible with the Python natsort package.
We'll move this to vendor/ once he accepts our PR.

However i won't trust it until we add unit tests:

These should be true:
  foo3 < foo10
	ny-dc-vpn < ny-dc01.ds
	* < 20161108174726pm._domainkey
	* < 3553988
	co-dc-vpn.stackexchange.com < co-dc01.ds.stackexchange.com

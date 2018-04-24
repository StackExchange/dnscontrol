package main

// this package only exists to force vendor a few packages that we will add in future prs.
// adding them here lets us do 'govendor add +e' to add the packages to vendor and keep them
// in the vendor directory, even if no other code references it.

import (
	_ "github.com/xenolf/lego/acmev2"
)

func main() {}

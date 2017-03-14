BEGIN {
  print ""
  print "D('" domain ", FILL_IN_REGISTRAR, DnsProvider(FILL_IN_PROVIDER),"
}

END {
  print "END)"
}

$3 == "SOA" || $4 == "SOA" {
  next
}

$3 == "A" || $3 == "CNAME" || $3 == "NS" {
  name = $1
  if (name == domain".") { name = "@" }
  print "\t" $3 "('" name "', '" $4 "')," ;
  next
}

$3 == "MX" {
  name = $1
  if (name == domain".") { name = "@" }
  print "\tMX('" name "', " $4 ", '" $5 "')," ;
  next
}

$3 == "TXT" {
  name = $1
  if (name == domain".") { name = "@" }
  $1 = "";
  $2 = "";
  $3 = "";
  print "\tTXT('" name "', " $0 ")," ;
  next
}

{ print "UNKNOWN:"  $0 }

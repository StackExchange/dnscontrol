#!/bin/bash

# Read the YAML files in the config directory and generate a production.yaml file.

cat <<HERE >production.yaml
---
providers:
  config:
    class: octodns.provider.yaml.YamlProvider
    directory: ./config
zones:
HERE

( cd config && ls *.yaml ) |
grep -v '^provider.yaml$' |
sed 's/.\{4\}$//' |
awk '{ print "    " $1 ":" ; print "    sources:" ; print "      - config" }' >>production.yaml

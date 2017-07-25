"use strict";

var conf = {
    registrars: [],
    dns_providers: [],
    domains: []
};

var defaultArgs = [];

function initialize(){
    conf = {
        registrars: [],
        dns_providers: [],
        domains: []
    };
    defaultArgs = [];
}

function NewRegistrar(name,type,meta) {
    if (type) {
      type == "MANUAL";
    }
    var reg = {name: name, type: type, meta: meta};
    conf.registrars.push(reg);
    return name;
}

function NewDnsProvider(name, type, meta) {
    if  ((typeof meta === 'object') && ('ip_conversions' in meta)) {
        meta.ip_conversions = format_tt(meta.ip_conversions)
    }
    var dsp = {name: name, type: type, meta: meta};
    conf.dns_providers.push(dsp);
    return name;
}

function newDomain(name,registrar) {
    return {name: name, registrar: registrar, meta:{}, records:[], dnsProviders: {}, defaultTTL: 0, nameservers:[]};
}

function processDargs(m, domain) {
        // for each modifier, if it is a...
        // function: call it with domain
        // array: process recursively
        // object: merge it into metadata
        if (_.isFunction(m)) {
               m(domain);
        } else if (_.isArray(m)) {
            for (var j in m) {
              processDargs(m[j], domain)
            }
        } else if (_.isObject(m)) {
            _.extend(domain.meta,m);
        } else {
          throw "WARNING: domain modifier type unsupported: "+ typeof m + " Domain: "+ domain.name;
        }
}

// D(name,registrar): Create a DNS Domain. Use the parameters as records and mods.
function D(name,registrar) {
    var domain = newDomain(name,registrar);
    for (var i = 0; i< defaultArgs.length; i++){
       processDargs(defaultArgs[i],domain)
   }
    for (var i = 2; i<arguments.length; i++) {
        var m = arguments[i];
        processDargs(m, domain)
    }
   conf.domains.push(domain)
}

// DEFAULTS provides a set of default arguments to apply to all future domains.
// Each call to DEFAULTS will clear any previous values set.
function DEFAULTS(){
    defaultArgs = [];
    for (var i = 0; i<arguments.length; i++) {
        defaultArgs.push(arguments[i]);
    }
}

// TTL(v): Set the TTL for a DNS record.
function TTL(v) {
    if (_.isString(v)){
        v = stringToDuration(v);
    }
    return function(r) {
        r.ttl = v;
    }
}

function stringToDuration(v){
    var matches = v.match(/^(\d+)([smhdwny]?)$/);
    if (matches == null){
        throw v + " is not a valid duration string"
    }
    unit = "s"
    if (matches[2]){
        unit = matches[2]
    }
    v = parseInt(matches[1])
    var u = {"s":1, "m":60, "h":3600}
    u["d"] = u.h * 24
    u["w"] = u.d * 7
    u["n"] = u.d * 30
    u["y"] = u.d * 365
    v *= u[unit];
    return v
}

// DefaultTTL(v): Set the default TTL for the domain.
function DefaultTTL(v) {
    if (_.isString(v)){
        v = stringToDuration(v);
    }
    return function(d) {
        d.defaultTTL = v;
    }
}

// CAA_CRITICAL: Critical CAA flag
var CAA_CRITICAL = 1<<0;


// DnsProvider("providerName", 0) 
// nsCount of 0 means don't use or register any nameservers.
// nsCount not provider means use all.
function DnsProvider(name, nsCount){
    if(typeof nsCount === 'undefined'){
        nsCount = -1;
    }
    return function(d) {
        d.dnsProviders[name] = nsCount;
    }
}

// A(name,ip, recordModifiers...)
function A(name, ip) {
    var mods = getModifiers(arguments,2)
    return function(d) {
        addRecord(d,"A",name,ip,mods)
    }
}

// AAAA(name,ip, recordModifiers...)
function AAAA(name, ip) {
    var mods = getModifiers(arguments,2)
    return function(d) {
        addRecord(d,"AAAA",name,ip,mods)
    }
}

// ALIAS(name,target, recordModifiers...)
function ALIAS(name, target) {
    var mods = getModifiers(arguments,2)
    return function(d) {
        addRecord(d,"ALIAS",name,target,mods)
    }
}

// CAA(name,tag,value, recordModifiers...)
function CAA(name, tag, value){
    checkArgs([_.isString, _.isString, _.isString], arguments, "CAA expects (name, tag, value) plus optional flag as a meta argument")

    var mods = getModifiers(arguments,3)
    mods.push({caatag: tag});

    return function(d) {
        addRecord(d,"CAA",name,value,mods)
    }
}


// CNAME(name,target, recordModifiers...)
function CNAME(name, target) {
    var mods = getModifiers(arguments,2)
    return function(d) {
        addRecord(d,"CNAME",name,target,mods)
    }
}

// PTR(name,target, recordModifiers...)
function PTR(name, target) {
    var mods = getModifiers(arguments,2)
    return function(d) {
        addRecord(d,"PTR",name,target,mods)
    }
}

// SRV(name,priority,weight,port,target, recordModifiers...)
function SRV(name, priority, weight, port, target) {
    checkArgs([_.isString, _.isNumber, _.isNumber, _.isNumber, _.isString], arguments, "SRV expects (name, priority, weight, port, target)")
    var mods = getModifiers(arguments,5)
    return function(d) {
        addRecordSRV(d, "SRV", name, priority, weight, port, target, mods)
    }
}

// TXT(name,target, recordModifiers...)
function TXT(name, target) {
    var mods = getModifiers(arguments,2)
    return function(d) {
        addRecord(d,"TXT",name,target,mods)
    }
}

// MX(name,priority,target, recordModifiers...)
function MX(name, priority, target) {
    checkArgs([_.isString, _.isNumber, _.isString], arguments, "MX expects (name, priority, target)")
    var mods = getModifiers(arguments,3)
    return function(d) {
        mods.push(priority);
        addRecord(d, "MX", name, target, mods)
    }
}

function checkArgs(checks, args, desc){
    if (args.length < checks.length){
        throw(desc)
    }
    for (var i = 0; i< checks.length; i++){
        if (!checks[i](args[i])){
            throw(desc+" - argument "+i+" is not correct type")
        }
    }
}

// NS(name,target, recordModifiers...)
function NS(name, target) {
    var mods = getModifiers(arguments,2)
    return function(d) {
        addRecord(d,"NS",name,target,mods)
    }
}

// NAMESERVER(name,target)
function NAMESERVER(name, target) {
    return function(d) {
        d.nameservers.push({name: name, target: target})
    }
}

function format_tt(transform_table) {
  // Turn [[low: 1, high: 2, newBase: 3], [low: 4, high: 5, newIP: 6]]
  // into "1 ~ 2 ~ 3 ~; 4 ~ 5 ~  ~ 6"
  var lines = []
  for (var i=0; i < transform_table.length; i++) {
    var ip = transform_table[i];
    var newIP = ip.newIP;
    if (newIP){
        if(_.isArray(newIP)){
            newIP = _.map(newIP,function(i){return num2dot(i)}).join(",")
        }else{
            newIP = num2dot(newIP);
        }
    }
    var newBase = ip.newBase;
    if (newBase){
        if(_.isArray(newBase)){
            newBase = _.map(newBase,function(i){return num2dot(i)}).join(",")
        }else{
            newBase = num2dot(newBase);
        }
    }
    var row = [
      num2dot(ip.low),
      num2dot(ip.high),
      newBase,
      newIP
    ]
    lines.push(row.join(" ~ "))
  }
  return lines.join(" ; ")
}

// IMPORT_TRANSFORM(translation_table, domain)
function IMPORT_TRANSFORM(translation_table, domain,ttl) {
    return function(d) {
        var rec = addRecord(d, "IMPORT_TRANSFORM", "@", domain, [
            {'transform_table': format_tt(translation_table)}])
        if (ttl){
            rec.ttl = ttl;
        }
    }
}

// PURGE()
function PURGE(d) {
  d.KeepUnknown = false
}

// NO_PURGE()
function NO_PURGE(d) {
  d.KeepUnknown = true
}

function getModifiers(args,start) {
    var mods = [];
    for (var i = start;i<args.length; i++) {
        mods.push(args[i])
    }
    return mods;
}

function addRecord(d,type,name,target,mods) {
    // if target is number, assume ip address. convert it.
    if (_.isNumber(target)) {
        target = num2dot(target);
    }
    var rec = {type: type, name: name, target: target, ttl:d.defaultTTL, priority: 0, meta:{}};
    // for each modifier, decide based on type:
    // - Function: call is with the record as the argument
    // - Object: merge it into the metadata
    // - Number: IF MX record assume it is priority
    if (mods) {
        for (var i = 0; i< mods.length; i++) {
            var m = mods[i]
            if (_.isFunction(m)) {
                m(rec);
            } else if (_.isObject(m) && m.caatag) {
                // caatag is a top level object, not in meta
                rec.caatag = m.caatag;
            } else if (_.isObject(m)) {
                 //convert transforms to strings
                 if (m.transform && _.isArray(m.transform)){
                    m.transform = format_tt(m.transform)
                 }
                _.extend(rec.meta,m);
                _.extend(rec.meta,m);
            } else if (_.isNumber(m) && type == "MX") {
               rec.mxpreference = m;
            } else if (_.isNumber(m) && type == "CAA") {
               rec.caaflags |= m;
            } else {
                console.log("WARNING: Modifier type unsupported:", typeof m, "(Skipping!)");
            }
        }
    }
    d.records.push(rec);
    return rec;
}

function addRecordSRV(d,type,name,srvpriority,srvweight,srvport,target,mods) {
    var rec = {type: type, name: name, srvpriority: srvpriority, srvweight: srvweight, srvport: srvport, target: target, ttl:d.defaultTTL, meta:{}};
    // for each modifier, decide based on type:
    // - Function: call is with the record as the argument
    // - Object: merge it into the metadata
    // FIXME(tlim): Factor this code out to its own function.
    if (mods) {
        for (var i = 0; i< mods.length; i++) {
            var m = mods[i]
            if (_.isFunction(m)) {
                m(rec);
            } else if (_.isObject(m)) {
                 //convert transforms to strings
                 if (m.transform && _.isArray(m.transform)){
                    m.transform = format_tt(m.transform)
                 }
                _.extend(rec.meta,m);
                _.extend(rec.meta,m);
            } else {
                console.log("WARNING: Modifier type unsupported:", typeof m, "(Skipping!)");
            }
        }
    }
    d.records.push(rec);
    return rec;
}

//ip conversion functions from http://stackoverflow.com/a/8105740/121660
// via http://javascript.about.com/library/blipconvert.htm
function IP(dot)
{
    var d = dot.split('.');
    return ((((((+d[0])*256)+(+d[1]))*256)+(+d[2]))*256)+(+d[3]);
}

function num2dot(num)
{
    if(num === undefined){
        return "";
    }
    if (_.isString(num)){
        return num
    }
    var d = num%256;
    for (var i = 3; i > 0; i--)
    {
        num = Math.floor(num/256);
        d = num%256 + '.' + d;
    }
    return d;
}


// Cloudflare aliases:

// Meta settings for individual records.
var CF_PROXY_OFF = {'cloudflare_proxy': 'off'};     // Proxy disabled.
var CF_PROXY_ON = {'cloudflare_proxy': 'on'};       // Proxy enabled.
var CF_PROXY_FULL = {'cloudflare_proxy': 'full'};   // Proxy+Railgun enabled.
// Per-domain meta settings:
// Proxy default off for entire domain (the default):
var CF_PROXY_DEFAULT_OFF = {'cloudflare_proxy_default': 'off'};
// Proxy default on for entire domain:
var CF_PROXY_DEFAULT_ON = {'cloudflare_proxy_default': 'on'};

// CUSTOM, PROVIDER SPECIFIC RECORD TYPES
function CF_REDIRECT(src, dst) {
    return function(d) {
        if (src.indexOf(",") !== -1 || dst.indexOf(",") !== -1){
            throw("redirect src and dst must not have commas")
        }
        addRecord(d,"CF_REDIRECT","@",src+","+dst)
    }
}
function CF_TEMP_REDIRECT(src, dst) {
    return function(d) {
        if (src.indexOf(",") !== -1 || dst.indexOf(",") !== -1){
            throw("redirect src and dst must not have commas")
        }
        addRecord(d,"CF_TEMP_REDIRECT","@",src+","+dst)
    }
}

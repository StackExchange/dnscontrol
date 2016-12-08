"use strict";

var conf = {
    registrars: [],
    dns_service_providers: [],
    domains: []
};

var defaultArgs = [];

function initialize(){
    conf = {
        registrars: [],
        dns_service_providers: [],
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

function NewDSP(name, type, meta) {
    if  ((typeof meta === 'object') && ('ip_conversions' in meta)) {
        meta.ip_conversions = format_tt(meta.ip_conversions)
    }
    var dsp = {name: name, type: type, meta: meta};
    conf.dns_service_providers.push(dsp);
    return name;
}

function newDomain(name,registrar) {
    return {name: name, registrar: registrar, meta:{}, records:[], dsps: {}, defaultTTL: 0, nameservers:[]};
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
    for (var i = 2; i<arguments.length; i++) {
        var m = arguments[i];
        processDargs(m, domain)
    }
   for (var i = 0; i< defaultArgs.length; i++){
       processDargs(defaultArgs[i],domain)
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
    return function(r) {
        r.ttl = v;
    }
}

// DefaultTTL(v): Set the default TTL for the domain.
function DefaultTTL(v) {
    return function(d) {
        d.defaultTTL = v;
    }
}



// DSP("providerName", 0) 
// nsCount of 0 means don't use or register any nameservers.
// nsCount not provider means use all.
function DSP(name, nsCount){
    if(typeof nsCount === 'undefined'){
        nsCount = -1;
    }
    return function(d) {
        d.dsps[name] = nsCount;
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

// CNAME(name,target, recordModifiers...)
function CNAME(name, target) {
    var mods = getModifiers(arguments,2)
    return function(d) {
        addRecord(d,"CNAME",name,target,mods)
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
    var mods = getModifiers(arguments,3)
    return function(d) {
        mods.push(priority);
        addRecord(d, "MX", name, target, mods)
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
            } else if (_.isObject(m)) {
                 //convert transforms to strings
                 if (m.transform && _.isArray(m.transform)){
                    m.transform = format_tt(m.transform)
                 }
                _.extend(rec.meta,m);
                _.extend(rec.meta,m);
            } else if (_.isNumber(m) && type == "MX") {
               rec.priority = m;
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

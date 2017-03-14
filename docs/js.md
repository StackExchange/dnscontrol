---
layout: default
---

# Javascript DSL

DNSControl uses javascript as its primary input language to provide power and flexibility to configure your domains. The ultimate purpose of the javascript is to consturct a
[DNSConfig](https://godoc.org/github.com/StackExchange/dnscontrol/models#DNSConfig) object that will be passed to the go backend and operated on. 

{% include funcList.md title="Top Level Functions" dir="global" %}

{% include funcList.md title="Domain Modifiers" dir="domain" %}

{% include funcList.md title="Record Modifiers" dir="record" %}

<script>
    $(function(){
        var f = function(){
            $("div.panel").removeClass("panel-success")
            var jmp = window.location.hash;
            if(jmp){
                $("div"+jmp).addClass("panel-success")
            }
        }
        f();
        $(window).on('hashchange',f); 
    })
</script>
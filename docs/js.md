---
layout: default
title: JavaScript DSL
---

# Javascript DSL

DNSControl uses javascript as its primary input language to provide power and flexibility to configure your domains. The ultimate purpose of the javascript is to construct a
[DNSConfig](https://godoc.org/github.com/StackExchange/dnscontrol/models#DNSConfig) object that will be passed to the go backend and operated on.

<table class="table-of-contents">
  <tr>
    <td>
        {% include table-of-contents.md
            docs-functions-dir="domain"
            html-anchor="domain-modifiers"
            title="Domain Modifiers"
        %}
    </td>
    <td>
        {% include table-of-contents.md
            docs-functions-dir="global"
            html-anchor="top-level-functions"
            title="Top Level Functions"
        %}
        {% include table-of-contents.md
            docs-functions-dir="record"
            html-anchor="record-modifiers"
            title="Record Modifiers"
        %}
    </td>
  </tr>
</table>

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

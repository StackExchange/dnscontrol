---
title: JavaScript DSL
---

# JavaScript DSL

DNSControl uses JavaScript as its primary input language to provide power and flexibility to configure your domains. The ultimate purpose of the JavaScript is to construct a
[DNSConfig](https://pkg.go.dev/github.com/StackExchange/dnscontrol/models#DNSConfig) object that will be passed to the go backend and operated on.

<table class="table-of-contents">
  <tr>
    <td>
        {% include table-of-contents.md
            docs-functions-dir="domain"
            html-anchor="domain-modifiers"
            title="Domain Modifiers"
        %}
        {% assign showProviders = 'AKAMAIEDGEDNS, AZURE_DNS, CLOUDFLAREAPI, ROUTE53' %}
        {% for provider in site.providers %}
            {% if showProviders contains provider.jsId %}
                {% include table-of-contents.md
                    docs-functions-dir="domain"
                    html-anchor="domain-modifiers"
                    title="Domain Modifiers"
                    provider-name=provider.name
                    provider-jsId=provider.jsId
                %}
            {% endif %}
        {% endfor %}
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

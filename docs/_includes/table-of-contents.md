<a href="#{{include.html-anchor}}">
    <strong>{{include.title}}</strong>
</a>

<ul>
{% for function in site.functions %}
    {% if function.path contains include.docs-functions-dir %}
        <li><a href="#{{function.name}}">{{function.name}}</a></li>
    {% endif %}
{% endfor %}
</ul>

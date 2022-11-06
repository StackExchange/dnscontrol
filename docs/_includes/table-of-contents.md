<a href="#{{include.html-anchor}}">
    <strong>{{include.title}}</strong>
</a>
{% if include.provider-jsId != nill %}
    for
    <a href="service-providers/{{include.provider-jsId | downcase}}">
        <strong>
            {{include.provider-name}}
        </strong>
    </a>
{% endif %}

<ul>
{% for function in site.functions %}
    {% if function.path contains include.docs-functions-dir %}
        {% if include.docs-functions-dir == "domain" %}
            {% if include.provider-jsId == nill and function.provider == nill %}
                <li><a href="#{{function.name}}">{{function.name}}</a></li>
            {% endif %}
            {% if include.provider-jsId != nill and function.provider == include.provider-jsId %}
                <li><a href="#{{function.name}}">{{function.name}}</a></li>
            {% endif %}
        {% else %}
            <li><a href="#{{function.name}}">{{function.name}}</a></li>
        {% endif %}
    {% endif %}
{% endfor %}
</ul>

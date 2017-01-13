## {{include.title}}
{% assign fs = site.functions %}
{% if include.dir == "global" %}
  {% assign fs = fs | reverse %}
{% endif %}


{% for f in fs %}
  {% if f.path contains include.dir %}
    {% include func.html f=f %}
  {% endif %}
{% endfor %}
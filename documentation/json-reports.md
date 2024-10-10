# JSON Reports

DNSControl can generate a machine-parseable report of changes.

The report is JSON formated and contains the zonename, the provider or
registrar name, and the number of changes.

To generate the report, add the `--report <filename>` option to a preview or
push command (this includes `preview`, `ppreview`, `push`,
`ppush`).


The report lists the changes that would be (preview) or are (push) attempted,
whether they are successful or not.

If a fatal error happens during the run, no report is generated.

## Sample output

{% code title="report.json" %}
```json
[
  {
    "domain": "private.example.com",
    "corrections": 10,
    "provider": "bind"
  },
  {
    "domain": "private.example.com",
    "corrections": 0,
    "registrar": "none"
  },
  {
    "domain": "admin.example.com",
    "corrections": 5,
    "provider": "bind"
  },
  {
    "domain": "admin.example.com",
    "corrections": 0,
    "registrar": "none"
  }
]
```
{% endcode %}

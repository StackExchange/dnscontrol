# JSON Reports

DNSControl has build in functionality to generate a machine-parseable report after pushing changes. This report is JSON formated and contains the zonename, the provider or registrar name and the amount of performed changes.

## Usage

To enable the report option you must use the `push` operation in combination with the `--report <filename>` option. This generates the json file.

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

# schemas

### Key presence

This controls whether the key has to appear in the body at all. Use the local constants from `schemas.go`:

| `.Required(...)` | key absent from body                | key present (any value) |
| ---              | ---                                 | ---                     |
| `Require`        | rejected: `required key is missing` | inner rules decide      |
| `Optional`       | skipped, no error                   | inner rules decide      |

`Require` and `Optional` are just `bool` aliases (`true` / `false`), they read better at the call site.

### Value guard

This runs ONLY when the key is present. It decides whether `null` / empty / zero are acceptable before the content rules
(length, format, enum, etc.) run. This is the table to memorize:

| first inner rule                      | `null`   | `""` (empty string) | `0` (a JSON number) | a real value |
| ---                                   | ---      | ---                 | ---                 | ---          |
| _none_ (just content rules)           | accepted | accepted            | accepted            | validated    |
| `validation.NotNil`                   | rejected | accepted            | accepted            | validated    |
| `validation.Required`                 | rejected | rejected            | accepted (see note) | validated    |
| `validation.OneOf(validation.Nil, R)` | accepted | per `R`             | per `R`             | validated    |

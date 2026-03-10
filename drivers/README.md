# drivers/

Vendor driver implementations used by Elsecall compilation/rendering.

## Layout
- `drivers/vendor/*`: canonical vendor drivers (junos, eos, iosxe, iosxr)
- `drivers/vendor/driver.go`: shared renderer contract vocabulary

## Conventions
- Keep output deterministic for fixture-based tests.
- Prefer explicit `TODO(truthwatcher)` markers for unsupported intent sections.
- Do not claim full vendor parity without fixture and integration coverage.

## TODO
- TODO(truthwatcher): register additional vendor drivers in default compiler wiring once baseline fixtures are stable.

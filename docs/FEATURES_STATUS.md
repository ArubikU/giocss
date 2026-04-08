# Features Status

This matrix tracks current giocss browser-like coverage.

## Implemented

- Layout core: flex flow, gaps, wrapping behavior, absolute/fixed exclusions in flow calculations.
- Text rendering stability: descender clipping fixes, placeholder alignment, inherited text CSS merge.
- Interaction states: `:hover`, `:active`, `:focus`, `:disabled`, `:checked`, `:invalid`.
- Form primitives: input text/date/time/number normalization, checkbox/radio state models, native select cycle model, and HTML-like `select` with `option` children.
- Snapshot runtime navigation pattern (state-driven views) demonstrated by [samples/sample-18-docs-viewer/main.go](../samples/sample-18-docs-viewer/main.go).

## Partial

- Native select parity: option cycling exists, but no full expanded dropdown list UX.
- CSS transitions: supported for core paths, but not full browser timing/function parity.
- Validation parity: required and core type checks exist; broader HTML constraint behavior is still incomplete.

## Missing

- Advanced selector grammar (`:nth-*`, combinators beyond current scope, attribute selector parity).
- Full form control set parity (e.g., complete textarea/select keyboard semantics, rich option groups).
- Browser-like layout edge cases (full min/max-content behavior and advanced fragmentation rules).

## Notes

- `disabled` now blocks pointer press/drag interaction for controls and prevents focus transitions.
- State-sensitive CSS cache keys include focus/disabled/checked/invalid dimensions to avoid stale styles.
- Native `select` now skips disabled options when cycling and becomes no-op when all options are disabled.
- Form submit collection now skips disabled controls, omits unchecked checkbox/radio controls, and aggregates repeated field names as arrays.
- Disabled `fieldset` submit behavior now preserves controls under the first `legend`, while excluding other descendants.
- Submit dispatch now includes the triggering submitter (`name`/`value`) and ignores non-submit buttons (`type=button`).

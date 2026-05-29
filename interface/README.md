# monetr's UI

This directory contains monetr's frontend code. This code is bundled using rsbuild and is output into
`../server/ui/static` which is then embedded into the go binary at build time such that it can serve the UI assets as a
standalone application.

This readme is a work in progress in order to improve code going forward since monetr now uses scss modules for all
styling.

## Typescript

- Don't use `any` and don't use `as unknown as Foo` thats just any but masked as some bullshit. The latter might be
  necessary but rarely.

## Styling

- Sizes of things should use `variables.$size` and should be calculated at build time. This way all sizes are relative
  to each other and scaling works nicely on all screens.
- Colors should never use scss variables like size does, colors are not calculated at build time and are instead handled
  at runtime. They should use regular css variables such as `var(--background)`.
- `mergeClasses` should not be used to do unconditional style merges. Merges should be done by writing entirely
  standalone classes in scss or using @extend such that they are combined at build time. `mergeClasses` should only be
  used for conditional styling or when a component accepts a `className` property. Even then `cva` is preferred for
  conditional styling when a component does not require a `className` property.
- Don't write `33.333333333333333333333333%` instead use `math.percentage(math.div(1, 3))`

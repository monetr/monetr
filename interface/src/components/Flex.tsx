import { cva } from 'class-variance-authority';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Flex.module.scss';

export const flexVariants = cva([styles.root], {
  variants: {
    gap: {
      none: undefined,
      sm: styles.gapSmall,
      md: styles.gapMedium,
      lg: styles.gapLarge,
      xl: styles.gapExtraLarge,
    },
    justify: {
      default: undefined,
      start: styles.justifyStart,
      center: styles.justifyCenter,
      between: styles.justifyBetween,
      end: styles.justifyEnd,
    },
    align: {
      default: undefined,
      center: styles.alignCenter,
    },
    orientation: {
      row: styles.flexRow,
      column: styles.flexColumn,
      stackSmall: styles.flexStackSmall,
      stackMedium: styles.flexStackMedium,
    },
    flex: {
      default: undefined,
      grow: styles.flexGrow,
      shrink: styles.flexShrink,
    },
    shrink: {
      default: undefined,
      none: styles.shrinkNone,
    },
    width: {
      default: styles.widthDefault,
      fit: styles.widthFit,
    },
  },
  defaultVariants: {
    justify: 'default',
    align: 'default',
    gap: 'md',
    orientation: 'row',
    shrink: 'default',
    width: 'default',
  },
});

export type VariantProps = Omit<Parameters<typeof flexVariants>[0], 'className' | 'class'>;

export type FlexProps = VariantProps & React.HTMLAttributes<HTMLDivElement>;

export default function Flex({
  gap,
  justify,
  align,
  orientation,
  flex,
  shrink,
  width,
  className,
  ...props
}: FlexProps): React.JSX.Element {
  return (
    <div
      className={mergeTailwind(flexVariants({ gap, justify, align, orientation, shrink, flex, width }), className)}
      {...props}
    />
  );
}

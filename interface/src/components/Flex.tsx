import { cva } from 'class-variance-authority';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Flex.module.scss';

export const flexVariants = cva([styles.root], {
  variants: {
    gap: {
      sm: styles.gapSmall,
      md: styles.gapMedium,
      lg: styles.gapLarge,
      xl: styles.gapExtraLarge,
    },
    justify: {
      default: undefined,
      center: styles.justifyCenter,
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
  },
  defaultVariants: {
    justify: 'default',
    align: 'default',
    gap: 'md',
    orientation: 'row',
  },
});

type VariantProps = Omit<Parameters<typeof flexVariants>[0], 'className' | 'class'>;

type FlexProps = VariantProps & React.HTMLAttributes<HTMLDivElement>;

export default function Flex({ gap, justify, align, orientation, className, ...props }: FlexProps): React.JSX.Element {
  return <div className={mergeTailwind(flexVariants({ gap, justify, align, orientation }), className)} {...props} />;
}

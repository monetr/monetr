import { cva } from 'class-variance-authority';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Typography.module.scss';

export const textVariants = cva([styles.root], {
  variants: {
    component: {
      span: undefined,
      p: undefined,
      code: styles.code,
    },
    color: {
      default: styles.colorDefault,
      muted: styles.colorMuted,
      subtle: styles.colorSubtle,
      emphasis: styles.colorEmphasis,
      inherit: styles.colorInherit,
    },
    ellipsis: {
      true: styles.ellipsis,
      false: undefined,
    },
    size: {
      xs: styles.textExtraSmall,
      sm: styles.textSmall,
      md: styles.textBase,
      lg: styles.textLarge,
      xl: styles.textExtraLarge,
      '5xl': styles.text5ExtraLarge,
    },
    weight: {
      normal: styles.weightNormal,
      medium: styles.weightMedium,
      semibold: styles.weightSemibold,
      bold: styles.weightBold,
    },
  },
  defaultVariants: {
    component: 'span',
    color: 'default',
    ellipsis: false,
    size: 'md',
    weight: 'normal',
  },
});

type VariantProps = Omit<Parameters<typeof textVariants>[0], 'className' | 'class'>;

export type TypographyProps = React.PropsWithChildren<
  VariantProps &
    (
      | ({ component?: 'span' } & React.HTMLAttributes<HTMLSpanElement>)
      | ({ component?: 'p' } & React.HTMLAttributes<HTMLParagraphElement>)
      | ({ component?: 'code' } & React.HTMLAttributes<HTMLElement>)
    )
>;

export default function Typography({
  component = 'span',
  color,
  ellipsis,
  size,
  weight,
  className,
  ...props
}: TypographyProps): React.JSX.Element {
  const TextElement: React.ElementType = component ?? 'span';
  return (
    <TextElement
      className={mergeTailwind(
        textVariants({
          component,
          color,
          ellipsis,
          size,
          weight,
        }),
        className,
      )}
      {...props}
    />
  );
}

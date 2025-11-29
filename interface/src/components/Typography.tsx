import { cva } from 'class-variance-authority';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Typography.module.scss';

export const textSizes = {
  xs: styles.textExtraSmall,
  sm: styles.textSmall,
  md: styles.textBase,
  lg: styles.textLarge,
  xl: styles.textExtraLarge,
  '2xl': styles.text2ExtraLarge,
  '5xl': styles.text5ExtraLarge,
};

export const textWeights = {
  normal: styles.weightNormal,
  medium: styles.weightMedium,
  semibold: styles.weightSemibold,
  bold: styles.weightBold,
};

export const textVariants = cva([styles.root], {
  variants: {
    component: {
      span: undefined,
      p: undefined,
      h3: undefined,
      code: styles.code,
    },
    color: {
      default: styles.colorDefault,
      muted: styles.colorMuted,
      subtle: styles.colorSubtle,
      emphasis: styles.colorEmphasis,
      inherit: styles.colorInherit,
      negative: styles.colorNegative,
      positive: styles.colorPositive,
    },
    ellipsis: {
      true: styles.ellipsis,
      false: undefined,
    },
    size: textSizes,
    weight: textWeights,
    align: {
      left: styles.alignLeft,
      center: styles.alignCenter,
      right: styles.alignRight,
    },
    wrapping: {
      default: undefined,
      nowrap: styles.noWrap,
    },
  },
  defaultVariants: {
    component: 'span',
    color: 'default',
    ellipsis: false,
    size: 'md',
    weight: 'normal',
    wrapping: 'default',
  },
});

type VariantProps = Omit<Parameters<typeof textVariants>[0], 'className' | 'class'>;

export type TypographyProps = React.PropsWithChildren<
  VariantProps &
    (
      | ({ component?: 'span' } & React.HTMLAttributes<HTMLSpanElement>)
      | ({ component?: 'h3' } & React.HTMLAttributes<HTMLHeadingElement>)
      | ({ component?: 'p' } & React.HTMLAttributes<HTMLParagraphElement>)
      | ({ component?: 'code' } & React.HTMLAttributes<HTMLElement>)
    )
>;

export default function Typography({
  component = 'span',
  align,
  color,
  ellipsis,
  size,
  weight,
  wrapping,
  className,
  ...props
}: TypographyProps): React.JSX.Element {
  const TextElement: React.ElementType = component ?? 'span';
  return (
    <TextElement
      className={mergeTailwind(
        textVariants({
          align,
          component,
          color,
          ellipsis,
          size,
          weight,
          wrapping,
        }),
        className,
      )}
      {...props}
    />
  );
}

import type { ReactNode } from 'react';

interface GradientHeadingProps {
  wrapperClassName: string;
  blurClassName: string;
  foregroundClassName: string;
  children: ReactNode;
  as?: 'h1' | 'h2' | 'span';
}

export default function GradientHeading(props: GradientHeadingProps): JSX.Element {
  const Heading = props.as ?? 'h1';
  return (
    <div className={props.wrapperClassName}>
      <span aria-hidden='true' className={props.blurClassName}>
        {props.children}
      </span>
      <Heading className={props.foregroundClassName}>{props.children}</Heading>
    </div>
  );
}

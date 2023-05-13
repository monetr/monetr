import clsx from "clsx";
import React from "react";

export interface MSpanProps {
  children: string | React.ReactNode | JSX.Element;
}

export default function MSpan(props: MSpanProps): JSX.Element {

  const classNames = clsx(
    'text-gray-900',
  );

  return (
    <span className={ classNames }>
      { props.children }
    </span>
  )

}

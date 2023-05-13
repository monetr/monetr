import clsx from "clsx";
import React from "react";
import { Link, LinkProps } from "react-router-dom";

type BaseLinkProps = LinkProps & React.RefAttributes<HTMLAnchorElement>
export interface MLinkProps extends BaseLinkProps {
  children: string | React.ReactNode | JSX.Element;
}

export default function MLink(props: MLinkProps): JSX.Element {
  const classNames = clsx(
    'font-semibold',
    'text-purple-500',
    'hover:text-purple-600',
  );

  return (
    <Link
      { ...props }
      className={ classNames }
    >
      { props.children }
    </Link>
  )
}

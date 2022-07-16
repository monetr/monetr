import React, { Fragment } from 'react';

import 'components/Code/CodeBasic/styles/CodeBasic.scss';

interface CodeBasicProps {
  className?: string;
  id?: string;
  children: string | JSX.Element | JSX.Element[];
}

export default function CodeBasic(props: CodeBasicProps): JSX.Element {
  let className = 'code-basic';
  if (props.className) {
    className += ` ${props.className}`;
  }

  return (
    <Fragment>
      <code className={ className } id={ props.id }>
        { props.children }
      </code>
    </Fragment>
  );
}

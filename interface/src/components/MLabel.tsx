import React from 'react';

import { ReactElement } from './types';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MLabelDecoratorProps {
  name?: string;
  disabled?: boolean;
}

export type MLabelDecorator = React.FC<MLabelDecoratorProps>;

export interface MLabelProps {
  htmlFor?: string;
  label?: string;
  required?: boolean;
  disabled?: boolean;
  children?: ReactElement;
}

export default function MLabel(props: MLabelProps): JSX.Element {


  function MaybeLabel(): JSX.Element {
    if (!props.label) return null;

    const labelClassNames = mergeTailwind(
      'mb-1',
      'block',
      'text-sm',
      'font-medium',
      'leading-6',
      {
        'text-gray-900': !props.disabled,
        'text-gray-500': props.disabled,
        'dark:text-dark-monetr-content-emphasis': !props.disabled,
      },
    );

    return (
      <label
        htmlFor={ props.htmlFor }
        className={ labelClassNames }
      >
        {props.label}
      </label>
    );
  }

  function MaybeRequired(): JSX.Element {
    if (!props.required) return null;
    return (
      <span className='text-red-500'>
        *
      </span>
    );
  }

  return (
    <div className="flex items-center justify-between">
      <div className='flex items-center gap-0.5'>
        <MaybeLabel />
        <MaybeRequired />
      </div>
      { props.children }
    </div>
  );
}

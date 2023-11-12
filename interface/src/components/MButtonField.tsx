import React from 'react';
import { ButtonBase, ButtonBaseProps } from '@mui/material';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';


export interface MButtonProps extends ButtonBaseProps {
  label?: string;
  error?: string;
  labelDecorator?: () => JSX.Element;
}


/**
 *  MButtonField is meant to be used in place of a select in the mobile UI.
 *  Since select's can sometimes cause a bad user experience depending on
 *  the device being used. This is meant to be a button that summons a proper
 *  menu to perform the selection.
 */
export default function MButtonField(props: MButtonProps): JSX.Element {
  const classNames = mergeTailwind(
    'dark:hover:ring-zinc-400',
    'dark:focus:ring-dark-monetr-brand',
    'dark:ring-dark-monetr-border-string',
    'dark:text-gray-400',
    'block',
    'border-0',
    'focus:ring-2',
    'focus:ring-inset',
    'text-gray-400',
    'px-3',
    'py-1.5',
    'ring-1',
    'ring-inset',
    'rounded-lg',
    'shadow-sm',
    'sm:leading-6',
    'sm:text-sm',
    'w-full',
    'dark:caret-zinc-50',
    'min-h-[38px]',
    'text-start',
    props.className,
  );

  const { labelDecorator, label, error, ...otherProps } = props;
  const noLabel = () => null;
  const LabelDecorator = labelDecorator || noLabel;

  function Error() {
    if (!error) return null;

    return (
      <p className='text-xs font-medium text-red-500 mt-0.5'>
        { error }
      </p>
    );
  }

  function LabelText(): JSX.Element {
    if (!label) return null;

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
        htmlFor={ props.id }
        className={ labelClassNames }
      >
        {label}
      </label>
    );
  }

  return (
    <div className='pb-[18px] w-full'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-0.5'>
          <LabelText />
        </div>
        <LabelDecorator />
      </div>
      <ButtonBase
        { ...otherProps }
        className={ classNames }
      />
      <Error />
    </div>
  );
}




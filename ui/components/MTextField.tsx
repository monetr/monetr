import clsx from "clsx";
import React from "react";

type InputProps = React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement>;
export interface MTextFieldProps extends InputProps {
  label?: string;
}

const MTextFieldPropsDefaults: MTextFieldProps = {
  label: null,
  disabled: false,
};

export default function MTextField(props: MTextFieldProps = MTextFieldPropsDefaults): JSX.Element {
  const labelClassNames = clsx(
    'mb-1',
    'block',
    'text-sm',
    'font-medium',
    'leading-6',
    {
      'text-gray-900': !props.disabled,
      'text-gray-400': props.disabled,
    },
  )

  function Label() {
    if (!props.label ) return null;

    return (
      <label
        htmlFor={ props.id }
        className={ labelClassNames }
      >
        { props.label }
      </label>
    );
  }

  const classNames = clsx(
    {
      'ring-gray-300': !props.disabled,
      'ring-gray-200': props.disabled,
    },
    'block',
    'border-0',
    'focus:ring-2',
    'focus:ring-inset',
    'focus:ring-purple-400',
    'placeholder:text-gray-400',
    'px-3',
    'py-1.5',
    'ring-1',
    'ring-inset',
    'rounded-lg',
    'shadow-sm',
    'sm:leading-6',
    'sm:text-sm',
    'text-gray-900',
    'w-full',
    props.className,
  );

  return (
    <div className={ props.className }>
      <Label />
      <div>
        <input
          { ...props }
          className={ classNames }
        />
      </div>
    </div>
  )
}

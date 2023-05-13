import styled from "@emotion/styled";
import React from "react";

export interface MCheckboxProps {
  id?: string;
  label: string;
}

export default function MCheckbox(props: MCheckboxProps): JSX.Element {
  const Checkbox = styled('input')(() => ({
    MozAppearance: 'none',
    WebkitAppearance: 'none',
    appearance: 'none',
    padding: '0',
    WebkitPrintColorAdjust: 'exact',
    colorAdjust: 'exact',
    display: 'inline-block',
    verticalAlign: 'middle',
    backgroundOrigin: 'border-box',
    WebkitUserSelect: 'none',
    MozUserSelect: 'none',
    userSelect: 'none',
    flexShrink: '0',
    height: '1rem',
    width: '1rem',
    color: 'red',
    backgroundColor: 'white',
    borderColor: 'gray',
    borderWidth: '1px',
    backgroundSize: '100% 100%',
    cursor: 'pointer',
    borderRadius: '0.25rem',
    '&:checked': {
      backgroundImage:
        "url(\"data:image/svg+xml;charset=utf-8,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3E%3Cpath" +
        " fill-rule='evenodd' clip-rule='evenodd' d='M12 5c-.28 0-.53.11-.71.29L7 9.59l-2.29-2.3a1.003 " +
        "1.003 0 00-1.42 1.42l3 3c.18.18.43.29.71.29s.53-.11.71-.29l5-5A1.003 1.003 0 0012 5z' fill='%23fff'/%3E%3C/svg%3E\")",
      backgroundColor: '#4E1AA0',
    },
  }));

  return (
    <div className="relative flex gap-x-3">
      <div className="flex h-6 items-center">
        <Checkbox
          id={ props.id }
          name="remember"
          type="checkbox"
          className="h-4 w-4 rounded border-gray-300 accent-purple-500 focus:ring-purple-500"
        />
      </div>
      <div className="text-sm leading-6">
        <label htmlFor={ props.id } className="font-medium text-gray-900">
          Remember me for 30 days
        </label>
        <p className="text-gray-500">Keep me logged in for a while.</p>
      </div>
    </div>
  )
}

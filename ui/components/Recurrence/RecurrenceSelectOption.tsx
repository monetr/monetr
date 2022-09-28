import React from 'react';
import { components, OptionProps } from 'react-select';

export interface SelectOption {
  readonly label: string;
  readonly value: number;
}

export function RecurrenceSelectionOption({ children, ...props }: OptionProps<SelectOption>): JSX.Element {
  // If the current amount is specified then format the amount, if it is not then use N/A.
  return (
    <components.Option { ...props }>
      <div className="w-full flex items-center">
        <span className="font-semibold">{ props.label }</span>
      </div>
    </components.Option>
  );
}

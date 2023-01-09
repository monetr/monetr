import React, { FocusEventHandler, useEffect, useState } from 'react';
import { FormControl, InputLabel, MenuItem, OutlinedInput } from '@mui/material';
import Select, { SelectChangeEvent } from '@mui/material/Select';

import getRecurrencesForDate from 'components/Recurrence/getRecurrencesForDate';
import Recurrence from 'components/Recurrence/Recurrence';
import clsx from 'clsx';

interface Props<T extends HTMLElement>{
  // TODO Add a way to pass a current value to the RecurrenceSelect component.
  className?: string;
  menuRef?: T;
  date: moment.Moment;
  onChange: { (value: Recurrence | null): void };
  disabled?: boolean;
  onBlur?: FocusEventHandler<HTMLInputElement>;
  label?: string;
}

export default function RecurrenceSelect<T extends HTMLElement>(props: Props<T>): JSX.Element {
  const [selectedIndex, setSelectedIndex] = useState<number>(-1);
  const rules = getRecurrencesForDate(props.date);

  useEffect(() => {
    if (selectedIndex === rules.length) {
      setSelectedIndex(null);
      props.onChange(null);
      return
    }

    props.onChange(rules[selectedIndex]);
  }, [props.date])

  function handleRecurrenceChange(event: SelectChangeEvent<number>, _: React.ReactNode) {
    const index = +event.target.value;

    setSelectedIndex(index);
    props.onChange(rules[index]);
  }

  const options = rules.map((item, index) => ({
    label: item.name,
    value: index,
  }));

  const value = selectedIndex !== null && selectedIndex >= 0 && selectedIndex < options.length ?
    options[selectedIndex] :
    { label: 'Select a frequency...', value: -1 };

  return (
    <FormControl
      className='w-full'
    >
      <InputLabel>{ props.label }</InputLabel>
      <Select
        displayEmpty
        renderValue={ (selected: number | null) => {
          if (selected !== null && selected >= 0) {
            return <span>{ options[selected].label }</span>
          }
          return <span className='text-gray-600'>Select a frequency...</span>
        }}
        value={ selectedIndex }
        onChange={ handleRecurrenceChange }
        onBlur={ props.onBlur }
        input={ <OutlinedInput label={ props.label } /> }
        disabled={ props.disabled }
      >
        <MenuItem disabled value={ -1 }>
          <em>Select a frequency...</em>
        </MenuItem>
        { options.map((option, idx) => (
          <MenuItem
            className={ clsx({
              'bg-purple-200': idx === value.value,
            })}
            key={ idx }
            value={ option.value }
          >
            <span
              className={ clsx({
                'font-medium': idx === value.value,
              })}
            >
              { option.label }
            </span>
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
}

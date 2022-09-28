import React, { useState } from 'react';
import Select, { ActionMeta, FormatOptionLabelMeta, OnChangeValue, Theme } from 'react-select';
import { lighten } from '@mui/material';

import getRecurrencesForDate from 'components/Recurrence/getRecurrencesForDate';
import Recurrence from 'components/Recurrence/Recurrence';
import { RecurrenceSelectionOption, SelectOption } from 'components/Recurrence/RecurrenceSelectOption';
import appTheme from 'theme';

interface Props<T extends HTMLElement>{
  // TODO Add a way to pass a current value to the RecurrenceSelect component.
  className?: string;
  menuRef?: T;
  date: moment.Moment;
  onChange: { (value: Recurrence): void };
  disabled?: boolean;
}

export default function RecurrenceSelect<T extends HTMLElement>(props: Props<T>): JSX.Element {
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);
  const rules = getRecurrencesForDate(props.date);

  function handleRecurrenceChange(newValue: OnChangeValue<SelectOption, false>, _: ActionMeta<SelectOption>) {
    const { onChange } = props;
    setSelectedIndex(newValue.value);
    onChange(rules[newValue.value]);
  }

  function formatOptionsLabel(option: SelectOption, meta: FormatOptionLabelMeta<SelectOption>): React.ReactNode {
    if (meta.context === 'value') {
      return option.label;
    }
    return option.label;
  }

  const options = rules.map((item, index) => ({
    label: item.name,
    value: index,
  }));

  const ref = props?.menuRef || document.body;
  const value = selectedIndex !== null && selectedIndex >= 0 && selectedIndex < options.length ?
    options[selectedIndex] :
    { label: 'Select a frequency...', value: -1 };

  const customStyles = {
    control: (base: object) => ({
      ...base,
      height: appTheme.components.MuiInputBase.styleOverrides.root['height'],
    }),
    menuPortal: (base: object) => ({
      ...base,
      zIndex: 9999,
    }),
  };

  return (
    <Select
      theme={ (theme: Theme): Theme => ({
        ...theme,
        borderRadius: +appTheme.shape.borderRadius,
        colors: {
          ...theme.colors,
          primary: appTheme.palette.primary.main,
          primary25: lighten(appTheme.palette.primary.main, 0.75),
          primary50: lighten(appTheme.palette.primary.main, 0.50),
          primary75: lighten(appTheme.palette.primary.main, 0.25),
          neutral0: appTheme.palette.background.default,
          // neutral30: appTheme.palette.primary.main,
          // neutral80: appTheme.palette.primary.contrastText,
          // neutral90: appTheme.palette.primary.contrastText,
        },
      }) }
      components={ {
        Option: RecurrenceSelectionOption,
      } }
      classNamePrefix="recurrence-select"
      className={ props.className }
      isClearable={ false }
      isDisabled={ props.disabled }
      isLoading={ false }
      onChange={ handleRecurrenceChange }
      options={ options }
      value={ value }
      formatOptionLabel={ formatOptionsLabel }
      styles={ customStyles }
      menuPlacement="auto"
      menuPortalTarget={ ref }
    />
  );
}

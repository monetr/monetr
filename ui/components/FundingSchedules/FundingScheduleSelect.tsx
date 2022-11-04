import React, { FocusEventHandler } from 'react';
import Select, { ActionMeta, components, FormatOptionLabelMeta, OnChangeValue, OptionProps, Theme } from 'react-select';
import { Button, lighten } from '@mui/material';
import moment from 'moment';

import { showCreateFundingScheduleDialog } from 'components/FundingSchedules/CreateFundingScheduleDialog';
import { useFundingSchedules } from 'hooks/fundingSchedules';
import appTheme from 'theme';

interface SelectOption {
  readonly label: string;
  readonly value: number;
  readonly nextOccurrence: moment.Moment;
}

interface Props<T extends HTMLElement>{
  className?: string;
  menuRef?: T;
  onChange: { (value: number): void };
  disabled?: boolean;
  value?: number;
  onBlur?: FocusEventHandler<HTMLInputElement>;
}

function FundingSelectionOption({ children, ...props }: OptionProps<SelectOption>): JSX.Element {
  // If the current amount is specified then format the amount, if it is not then use N/A.
  return (
    <components.Option { ...props }>
      <div className="w-full flex items-center">
        <span className="font-semibold">{ props.label }</span>
      </div>
    </components.Option>
  );
}

export default function FundingScheduleSelect<T extends HTMLElement>(props: Props<T>): JSX.Element {
  const fundingSchedules = useFundingSchedules();

  function handleChange(newValue: OnChangeValue<SelectOption, false>, _: ActionMeta<SelectOption>) {
    const { onChange } = props;
    onChange(newValue.value);
  }

  function formatOptionsLabel(option: SelectOption, meta: FormatOptionLabelMeta<SelectOption>): React.ReactNode {
    if (meta.context === 'value') {
      return option.label;
    }
    return option.label;
  }

  const options = Array.from(fundingSchedules.values()).map(item => ({
    label: item.name,
    value: item.fundingScheduleId,
    nextOcurrence: item.nextOccurrence,
  }));

  const ref = props?.menuRef || document.body;
  const value = options.find(item => item.value === props.value) || {
    label: 'Select a funding schedule...',
    value: -1,
  };

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

  if (fundingSchedules.size === 0) {
    return (
      <Button
        size='large'
        onClick={ showCreateFundingScheduleDialog }
        className={ [props.className, 'mt-3'].join(' ').trim() }
        variant='outlined'
        color='secondary'
      >
        Create your first funding schedule.
      </Button>
    );
  }

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
        Option: FundingSelectionOption,
      } }
      classNamePrefix="recurrence-select"
      className={ props.className }
      isClearable={ false }
      isDisabled={ props.disabled }
      isLoading={ false }
      onChange={ handleChange }
      options={ options }
      value={ value }
      formatOptionLabel={ formatOptionsLabel }
      styles={ customStyles }
      menuPlacement="auto"
      menuPortalTarget={ ref }
      onBlur={ props.onBlur }
    />
  );
}


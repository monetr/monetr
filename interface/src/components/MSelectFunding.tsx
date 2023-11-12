import React from 'react';
import { TodayOutlined } from '@mui/icons-material';
import { useFormikContext } from 'formik';

import { MBaseButton } from './MButton';
import MLabel from './MLabel';
import MSelect from './MSelect';
import { useFundingSchedulesSink } from '@monetr/interface/hooks/fundingSchedules';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';

export interface MSelectFundingProps {
  label?: string;
  name: string;
  required?: boolean;
  className?: string;
  menuPortalTarget?: HTMLElement;
}

export default function MSelectFunding(props: MSelectFundingProps): JSX.Element {
  const formikContext = useFormikContext();
  const {
    data: funding,
    isLoading: fundingIsLoading,
    isError: fundingIsError,
  } = useFundingSchedulesSink();
  const label = props.label ?? 'Select a funding schedule';

  if (fundingIsLoading) {
    return <MSelect
      className={ props?.className }
      disabled
      isLoading
      label={ label }
      menuPortalTarget={ props.menuPortalTarget }
      placeholder='Select a funding schedule...'
      required={ props?.required }
    />;
  }

  if (fundingIsError) {
    return <MSelect
      className={ props?.className }
      disabled
      isLoading
      label={ label }
      menuPortalTarget={ props.menuPortalTarget }
      placeholder='Failed to load funding schedules...'
      required={ props?.required }
    />;
  }

  function createAndSetFunding() {
    showNewFundingModal()
      .then(result => formikContext.setFieldValue(props.name, result.fundingScheduleId));
  }

  if (funding.length === 0) {
    return (
      <div className='h-[84px] w-full'>
        <MLabel
          label={ props.label }
          required={ props.required }
        />
        <MBaseButton
          color='primary'
          variant='solid'
          className='w-full h-[38px] font-normal text-start justify-start gap-2'
          onClick={ createAndSetFunding }
        >
          <TodayOutlined />
          Create a new funding schedule...
        </MBaseButton>
      </div>
    );
  }

  const options = Array.from(funding.values()).map(item => ({
    label: item.name,
    value: item.fundingScheduleId,
  }));

  const value = options.find(option => option.value === formikContext.values[props.name]);

  function onSelect(newValue: { label: string, value: number }) {
    formikContext.setFieldValue(props.name, newValue.value);
  }

  return (
    <MSelect
      label='When do you want to fund the expense?'
      menuPlacement='auto'
      menuPortalTarget={ props.menuPortalTarget }
      menuPosition='fixed'
      name='fundingScheduleId'
      onChange={ onSelect }
      options={ options }
      placeholder='Select a funding schedule...'
      required={ props.required }
      value={ value }
      className={ props.className }
    />
  );
}

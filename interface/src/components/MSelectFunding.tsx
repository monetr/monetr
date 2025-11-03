import { useFormikContext } from 'formik';
import { Calendar } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Label from '@monetr/interface/components/Label';
import Select, { type SelectOption } from '@monetr/interface/components/Select';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
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
  const { data: funding, isLoading: fundingIsLoading, isError: fundingIsError } = useFundingSchedules();
  const label = props.label ?? 'Select a funding schedule';

  if (fundingIsLoading) {
    return (
      <Select
        className={props?.className}
        disabled
        isLoading
        label={label}
        placeholder='Select a funding schedule...'
        required={props?.required}
        options={[]}
        onChange={() => {}}
      />
    );
  }

  if (fundingIsError) {
    return (
      <Select
        className={props?.className}
        disabled
        label={label}
        placeholder='Failed to loading funding schedules...'
        required={props?.required}
        options={[]}
        onChange={() => {}}
      />
    );
  }

  function createAndSetFunding() {
    showNewFundingModal().then(result => formikContext.setFieldValue(props.name, result.fundingScheduleId));
  }

  if (funding.length === 0) {
    return (
      <div className='h-[84px] w-full'>
        <Label label={props.label} required={props.required} />
        <Button
          variant='primary'
          size='select'
          className='w-full font-normal text-start justify-start gap-2'
          onClick={createAndSetFunding}
        >
          <Calendar />
          Create a new funding schedule...
        </Button>
      </div>
    );
  }

  const options = Array.from(funding.values()).map(item => ({
    label: item.name,
    value: item.fundingScheduleId,
  }));

  const value = options.find(option => option.value === formikContext.values[props.name]);

  function onSelect(newValue: SelectOption<string>) {
    formikContext.setFieldValue(props.name, newValue.value);
  }

  return (
    <Select
      label={props.label ?? 'Funding'}
      name='fundingScheduleId'
      onChange={onSelect}
      options={options}
      placeholder='Select a funding schedule...'
      required={props.required}
      value={value}
      className={props.className}
    />
  );
}

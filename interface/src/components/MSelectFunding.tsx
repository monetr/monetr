import { useFormikContext } from 'formik';
import { Calendar } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Label from '@monetr/interface/components/Label';
import Select, { type SelectOption } from '@monetr/interface/components/Select';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';

import styles from './MSelectFunding.module.scss';

export interface MSelectFundingProps {
  label?: string;
  name: string;
  required?: boolean;
  className?: string;
  menuPortalTarget?: HTMLElement;
}

export default function MSelectFunding(props: MSelectFundingProps): React.JSX.Element {
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
        onChange={() => {}}
        options={[]}
        placeholder='Select a funding schedule...'
        required={props?.required}
      />
    );
  }

  if (fundingIsError) {
    return (
      <Select
        className={props?.className}
        disabled
        label={label}
        onChange={() => {}}
        options={[]}
        placeholder='Failed to loading funding schedules...'
        required={props?.required}
      />
    );
  }

  function createAndSetFunding() {
    showNewFundingModal().then(result => formikContext.setFieldValue(props.name, result.fundingScheduleId));
  }

  if (funding.length === 0) {
    return (
      <div className={styles.emptyState}>
        <Label label={props.label} required={props.required} />
        <Button className={styles.createButton} onClick={createAndSetFunding} size='select' variant='primary'>
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
      className={props.className}
      label={props.label ?? 'Funding'}
      name='fundingScheduleId'
      onChange={onSelect}
      options={options}
      placeholder='Select a funding schedule...'
      required={props.required}
      value={value}
    />
  );
}

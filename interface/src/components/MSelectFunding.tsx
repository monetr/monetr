import { useEffect } from 'react';
import { useFormikContext } from 'formik';
import { Calendar } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Label from '@monetr/interface/components/Label';
import Select, { type SelectOption } from '@monetr/interface/components/Select';
import { useFundingSchedules } from '@monetr/interface/hooks/useFundingSchedules';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';
import { ID } from '@monetr/interface/models/ID';

import styles from './MSelectFunding.module.scss';

export interface MSelectFundingProps {
  label?: string;
  name: string;
  required?: boolean;
  className?: string;
  menuPortalTarget?: HTMLElement;
}

export default function MSelectFunding(props: MSelectFundingProps): React.JSX.Element {
  const formikContext = useFormikContext<Record<string, any>>();
  const { data: funding, isLoading: fundingIsLoading, isError: fundingIsError } = useFundingSchedules();
  const label = props.label ?? 'Select a funding schedule';

  // if theres only one funding schedule then theres not really a choice to make, so just pick it for them as long as
  // nothing is picked yet
  const onlyFundingScheduleId = funding?.length === 1 ? funding[0]?.fundingScheduleId : undefined;
  const currentValue = formikContext.values[props.name];
  // biome-ignore lint/correctness/useExhaustiveDependencies: only want this to run when the funding schedules or the value change
  useEffect(() => {
    if (onlyFundingScheduleId && ID.isZero(currentValue)) {
      formikContext.setFieldValue(props.name, onlyFundingScheduleId);
    }
  }, [onlyFundingScheduleId, currentValue]);

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
    showNewFundingModal().then(result => result && formikContext.setFieldValue(props.name, result.fundingScheduleId));
  }

  if (!funding || funding.length === 0) {
    return (
      <div className={styles.emptyState}>
        <Label label={props.label} required={props.required} />
        <Button
          className={styles.createButton}
          disabled={formikContext.isSubmitting}
          onClick={createAndSetFunding}
          size='select'
          variant='primary'
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
      className={props.className}
      disabled={formikContext.isSubmitting}
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

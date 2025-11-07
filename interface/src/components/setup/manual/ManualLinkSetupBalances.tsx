import { Fragment } from 'react';
import type { FormikHelpers } from 'formik';

import FormAmountField from '@monetr/interface/components/FormAmountField';
import MForm from '@monetr/interface/components/MForm';
import MSpan from '@monetr/interface/components/MSpan';
import SelectCurrency from '@monetr/interface/components/SelectCurrency';
import type { ManualLinkSetupForm } from '@monetr/interface/components/setup/manual/ManualLinkSetup';
import ManualLinkSetupButtons from '@monetr/interface/components/setup/manual/ManualLinkSetupButtons';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import { useViewContext } from '@monetr/interface/components/ViewManager';

export type ManualLinkSetupBalancesValues = {
  startingBalance: number;
  currency: string;
};

export default function ManualLinkSetupBalances(): JSX.Element {
  const viewContext = useViewContext<ManualLinkSetupSteps, unknown, ManualLinkSetupForm>();
  const initialValues: ManualLinkSetupBalancesValues = {
    startingBalance: 0.0,
    currency: 'USD',
    ...viewContext.formData,
  };

  function submit(values: ManualLinkSetupBalancesValues, helpers: FormikHelpers<ManualLinkSetupBalancesValues>) {
    helpers.setSubmitting(true);
    viewContext.updateFormData(values);
    viewContext.goToView(ManualLinkSetupSteps.Income);
  }

  return (
    <MForm
      initialValues={initialValues}
      onSubmit={submit}
      className='w-full flex flex-col justify-center items-center gap-2'
    >
      {({ values: { currency } }) => (
        <Fragment>
          <MSpan size='lg' color='subtle' className='text-center'>
            What is your current available balance? monetr will use this as a starting point, you can modify this at any
            time later on.
          </MSpan>
          <SelectCurrency name='currency' className='w-full' />
          <FormAmountField
            name='startingBalance'
            label='Starting Balance'
            className='w-full'
            currency={currency}
            required
            allowNegative
          />
          <ManualLinkSetupButtons />
        </Fragment>
      )}
    </MForm>
  );
}

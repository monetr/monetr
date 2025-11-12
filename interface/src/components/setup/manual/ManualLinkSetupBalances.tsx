import { Fragment } from 'react';
import type { FormikHelpers } from 'formik';

import { flexVariants } from '@monetr/interface/components/Flex';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import MForm from '@monetr/interface/components/MForm';
import SelectCurrency from '@monetr/interface/components/SelectCurrency';
import type { ManualLinkSetupForm } from '@monetr/interface/components/setup/manual/ManualLinkSetup';
import ManualLinkSetupButtons from '@monetr/interface/components/setup/manual/ManualLinkSetupButtons';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useViewContext } from '@monetr/interface/components/ViewManager';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';

export type ManualLinkSetupBalancesValues = {
  startingBalance: number;
  currency: string;
};

export default function ManualLinkSetupBalances(): JSX.Element {
  const viewContext = useViewContext<ManualLinkSetupSteps, unknown, ManualLinkSetupForm>();
  const { data: currency } = useLocaleCurrency();

  function submit(values: ManualLinkSetupBalancesValues, helpers: FormikHelpers<ManualLinkSetupBalancesValues>) {
    helpers.setSubmitting(true);
    viewContext.updateFormData(values);
    viewContext.goToView(ManualLinkSetupSteps.Income);
  }

  const initialValues: ManualLinkSetupBalancesValues = {
    startingBalance: 0.0,
    currency: currency.currency,
    ...viewContext.formData,
  };

  return (
    <MForm
      className={flexVariants({
        orientation: 'column',
        justify: 'center',
        align: 'center',
      })}
      initialValues={initialValues}
      onSubmit={submit}
    >
      {({ values: { currency } }) => (
        <Fragment>
          <Typography align='center' color='subtle' size='lg'>
            What is your current available balance? monetr will use this as a starting point, you can modify this at any
            time later on.
          </Typography>
          <SelectCurrency className='w-full' name='currency' />
          <FormAmountField
            allowNegative
            className='w-full'
            currency={currency}
            label='Starting Balance'
            name='startingBalance'
            required
          />
          <ManualLinkSetupButtons />
        </Fragment>
      )}
    </MForm>
  );
}

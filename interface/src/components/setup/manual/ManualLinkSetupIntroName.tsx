import type { FormikHelpers } from 'formik';

import { flexVariants } from '@monetr/interface/components/Flex';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import type { ManualLinkSetupForm } from '@monetr/interface/components/setup/manual/ManualLinkSetup';
import ManualLinkSetupButtons from '@monetr/interface/components/setup/manual/ManualLinkSetupButtons';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useViewContext } from '@monetr/interface/components/ViewManager';

export type ManualLinkSetupIntroNameValues = {
  budgetName: string;
};

export default function ManualLinkSetupIntroName(): JSX.Element {
  const viewContext = useViewContext<ManualLinkSetupSteps, unknown, ManualLinkSetupForm>();
  const initialValues: ManualLinkSetupIntroNameValues = {
    budgetName: '',
    ...viewContext.formData,
  };

  function submit(values: ManualLinkSetupIntroNameValues, helpers: FormikHelpers<ManualLinkSetupIntroNameValues>) {
    helpers.setSubmitting(true);
    viewContext.updateFormData(values);
    viewContext.goToView(ManualLinkSetupSteps.AccountName);
  }

  return (
    <MForm
      initialValues={initialValues}
      onSubmit={submit}
      className={flexVariants({
        orientation: 'column',
        justify: 'center',
        align: 'center',
      })}
    >
      <Typography size='2xl' weight='medium'>
        Welcome to monetr!
      </Typography>
      <Typography size='lg' color='subtle' align='center'>
        Let's create a new budget to get started. What do you want to call this budget?
      </Typography>
      <FormTextField
        name='budgetName'
        label='Bank or Budget Name'
        className='w-full'
        placeholder='My Primary Bank'
        autoFocus
        required
      />
      <ManualLinkSetupButtons />
    </MForm>
  );
}

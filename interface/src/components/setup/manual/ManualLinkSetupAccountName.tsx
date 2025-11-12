import type { FormikHelpers } from 'formik';

import { flexVariants } from '@monetr/interface/components/Flex';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import type { ManualLinkSetupForm } from '@monetr/interface/components/setup/manual/ManualLinkSetup';
import ManualLinkSetupButtons from '@monetr/interface/components/setup/manual/ManualLinkSetupButtons';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useViewContext } from '@monetr/interface/components/ViewManager';

export interface ManualLinkSetupAccountNameValues {
  accountName: string;
}

export default function ManualLinkSetupAccountName(): JSX.Element {
  const viewContext = useViewContext<ManualLinkSetupSteps, unknown, ManualLinkSetupForm>();
  const initialValues: ManualLinkSetupAccountNameValues = {
    accountName: '',
    ...viewContext.formData,
  };

  function submit(values: ManualLinkSetupAccountNameValues, helpers: FormikHelpers<ManualLinkSetupAccountNameValues>) {
    helpers.setSubmitting(true);
    viewContext.updateFormData(values);
    viewContext.goToView(ManualLinkSetupSteps.Balances);
  }

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
      <Typography align='center' color='subtle' size='lg'>
        What do you want to call the primary account you want to use for budgeting? For example; your checking account?
      </Typography>
      <FormTextField
        autoFocus
        className='w-full'
        label='Account Name'
        name='accountName'
        placeholder='My Checking Account'
        required
      />
      <ManualLinkSetupButtons />
    </MForm>
  );
}

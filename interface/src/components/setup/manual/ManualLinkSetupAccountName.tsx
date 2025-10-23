import type { FormikHelpers } from 'formik';

import MForm from '@monetr/interface/components/MForm';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import ManualLinkSetupButtons from '@monetr/interface/components/setup/manual/ManualLinkSetupButtons';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import { useViewContext } from '@monetr/interface/components/ViewManager';

interface Values {
  accountName: string;
}

export default function ManualLinkSetupAccountName(): JSX.Element {
  const viewContext = useViewContext<ManualLinkSetupSteps, unknown>();
  const initialValues: Values = {
    accountName: '',
    ...viewContext.formData,
  };

  function submit(values: Values, helpers: FormikHelpers<Values>) {
    helpers.setSubmitting(true);
    viewContext.updateFormData(values);
    viewContext.goToView(ManualLinkSetupSteps.Balances);
  }

  return (
    <MForm
      initialValues={initialValues}
      onSubmit={submit}
      className='w-full flex flex-col justify-center items-center gap-2'
    >
      <MSpan size='lg' color='subtle' className='text-center'>
        What do you want to call the primary account you want to use for budgeting? For example; your checking account?
      </MSpan>
      <MTextField
        name='accountName'
        label='Account Name'
        className='w-full'
        placeholder='My Checking Account'
        autoFocus
        required
      />
      <ManualLinkSetupButtons />
    </MForm>
  );
}

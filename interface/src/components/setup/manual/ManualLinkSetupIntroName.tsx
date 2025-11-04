import type { FormikHelpers } from 'formik';

import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MSpan from '@monetr/interface/components/MSpan';
import ManualLinkSetupButtons from '@monetr/interface/components/setup/manual/ManualLinkSetupButtons';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import { useViewContext } from '@monetr/interface/components/ViewManager';

interface Values {
  budgetName: string;
}

export default function ManualLinkSetupIntroName(): JSX.Element {
  const viewContext = useViewContext<ManualLinkSetupSteps, unknown>();
  const initialValues: Values = {
    budgetName: '',
    ...viewContext.formData,
  };

  function submit(values: Values, helpers: FormikHelpers<Values>) {
    helpers.setSubmitting(true);
    viewContext.updateFormData(values);
    viewContext.goToView(ManualLinkSetupSteps.AccountName);
  }

  return (
    <MForm
      initialValues={initialValues}
      onSubmit={submit}
      className='w-full flex flex-col justify-center items-center gap-2'
    >
      <MSpan size='2xl' weight='medium'>
        Welcome to monetr!
      </MSpan>
      <MSpan size='lg' color='subtle' className='text-center'>
        Let's create a new budget to get started. What do you want to call this budget?
      </MSpan>
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

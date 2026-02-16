import { useCallback, useMemo } from 'react';
import type { FormikHelpers } from 'formik';
import { useNavigate } from 'react-router-dom';

import { flexVariants } from '@monetr/interface/components/Flex';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import { layoutVariants } from '@monetr/interface/components/Layout';
import MForm from '@monetr/interface/components/MForm';
import LunchFlowSetupLayout from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupLayout';
import { LunchFlowSetupSteps } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import LunchFlowLink from '@monetr/interface/models/LunchFlowLink';
import request from '@monetr/interface/util/request';

export type LunchFlowSetupIntroValues = {
  name: string;
  apiURL: string;
  apiKey: string;
};

export default function LunchFlowSetupIntro(): React.JSX.Element {
  const { data: config } = useAppConfiguration();
  const navigate = useNavigate();

  const initialValues: LunchFlowSetupIntroValues = useMemo(
    () => ({
      name: '',
      apiKey: '',
      apiURL: config.lunchFlowDefaultAPIURL,
    }),
    [config],
  );

  const submit = useCallback(
    (values: LunchFlowSetupIntroValues, helpers: FormikHelpers<LunchFlowSetupIntroValues>) => {
      helpers.setSubmitting(true);
      request()
        .post<Partial<LunchFlowLink>>(`/lunch_flow/link`, {
          name: values.name,
          lunchFlowURL: values.apiURL,
          apiKey: values.apiKey,
        })
        .then(result => new LunchFlowLink(result?.data))
        .then(lunchFlowLink =>
          navigate(lunchFlowLink.lunchFlowLinkId, {
            relative: 'path',
          }),
        )
        .catch(error => {
          helpers.setSubmitting(false);
          console.error(error);
        });
    },
    [navigate],
  );

  return (
    <LunchFlowSetupLayout step={LunchFlowSetupSteps.Intro}>
      <MForm
        className={flexVariants({
          orientation: 'column',
          justify: 'center',
          align: 'center',
        })}
        initialValues={initialValues}
        onSubmit={submit}
      >
        <Typography size='2xl' weight='medium'>
          Lunch Flow Setup
        </Typography>
        <Typography align='center' color='subtle' size='lg'>
          Let us know what you want to call this budget, and what your Lunch Flow API key is.
        </Typography>
        <FormTextField
          autoFocus
          className={layoutVariants({ width: 'full' })}
          data-1p-ignore
          label='Budget Name'
          name='name'
          placeholder='My Primary Bank'
          required
        />
        <FormTextField
          className={layoutVariants({ width: 'full' })}
          data-1p-ignore
          label='API URL'
          name='apiURL'
          placeholder='https://.../api/v1'
          required
          type='url'
        />
        <FormTextField
          className={layoutVariants({ width: 'full' })}
          data-1p-ignore
          label='API Secret'
          name='apiKey'
          required
          type='password'
        />
        <FormButton type='submit' variant='primary'>
          Next
        </FormButton>
      </MForm>
    </LunchFlowSetupLayout>
  );
}

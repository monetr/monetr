import { useCallback, useMemo } from 'react';
import type { FormikHelpers } from 'formik';
import { useFormikContext } from 'formik';
import { useSnackbar } from 'notistack';
import { useNavigate } from 'react-router-dom';

import type { ApiError } from '@monetr/interface/api/client';
import { flexVariants } from '@monetr/interface/components/Flex';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import { layoutVariants } from '@monetr/interface/components/Layout';
import MForm from '@monetr/interface/components/MForm';
import Select, { type SelectOption } from '@monetr/interface/components/Select';
import LunchFlowSetupLayout from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupLayout';
import { LunchFlowSetupSteps } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import LunchFlowLink from '@monetr/interface/models/LunchFlowLink';
import request, { type APIError } from '@monetr/interface/util/request';

export type LunchFlowSetupIntroValues = {
  name: string;
  apiURL: string;
  apiKey: string;
};

export default function LunchFlowSetupIntro(): React.JSX.Element {
  const { data: config } = useAppConfiguration();
  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();

  const allowedAPIURLs = config.lunchFlowAllowedAPIURLs ?? [];
  const initialApiURL = allowedAPIURLs.length > 0 ? allowedAPIURLs[0] : '';

  const initialValues: LunchFlowSetupIntroValues = useMemo(
    () => ({
      name: '',
      apiKey: '',
      apiURL: initialApiURL,
    }),
    [initialApiURL],
  );

  const submit = useCallback(
    (values: LunchFlowSetupIntroValues, helpers: FormikHelpers<LunchFlowSetupIntroValues>) => {
      helpers.setSubmitting(true);
      request<Partial<LunchFlowLink>>({
        method: 'POST',
        url: '/api/lunch_flow/link',
        data: {
          name: values.name,
          lunchFlowURL: values.apiURL,
          apiKey: values.apiKey,
        },
      })
        .then(result => new LunchFlowLink(result?.data))
        .then(lunchFlowLink =>
          navigate(lunchFlowLink.lunchFlowLinkId, {
            relative: 'path',
          }),
        )
        .catch((error: ApiError<APIError>) =>
          enqueueSnackbar(
            <div>
              <Typography size='sm'>{error?.response?.data?.error || 'Failed to create Lunch Flow link!'}</Typography>
              {Object.entries(error?.response?.data?.problems).map(([key, problem]) => (
                <Typography component='code' key={key} size='xs'>
                  {`${key}: ${problem}`}
                </Typography>
              ))}
            </div>,
            {
              variant: 'error',
              disableWindowBlurListener: true,
            },
          ),
        )
        .finally(() => helpers.setSubmitting(false));
    },
    [navigate, enqueueSnackbar],
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
          autoComplete='off'
          autoFocus
          className={layoutVariants({ width: 'full' })}
          data-1p-ignore
          label='Budget Name'
          name='name'
          placeholder='Lunch Flow'
          required
        />
        <FormTextField
          autoComplete='off'
          className={layoutVariants({ width: 'full' })}
          data-1p-ignore
          label='API Key'
          name='apiKey'
          placeholder='●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●'
          required
          type='password'
        />
        <LunchFlowURLField allowedAPIURLs={allowedAPIURLs} />
        <FormButton type='submit' variant='primary'>
          Next
        </FormButton>
      </MForm>
    </LunchFlowSetupLayout>
  );
}

interface LunchFlowURLFieldProps {
  allowedAPIURLs: string[];
}

function LunchFlowURLField({ allowedAPIURLs }: LunchFlowURLFieldProps): React.JSX.Element {
  const formikContext = useFormikContext<LunchFlowSetupIntroValues>();

  if (allowedAPIURLs.length > 1) {
    const options: SelectOption<string>[] = allowedAPIURLs.map(url => ({ label: url, value: url }));
    const value = options.find(option => option.value === formikContext.values.apiURL);
    return (
      <Select
        className={layoutVariants({ width: 'full' })}
        label='API URL'
        name='apiURL'
        onChange={(newValue: SelectOption<string>) => formikContext.setFieldValue('apiURL', newValue.value)}
        options={options}
        required
        value={value}
      />
    );
  }

  return (
    <FormTextField
      className={layoutVariants({ width: 'full' })}
      data-1p-ignore
      disabled
      label='API URL'
      name='apiURL'
      placeholder='No API URLs configured!'
      required
      spellCheck='false'
      type='url'
    />
  );
}

import { useCallback } from 'react';
import type { FormikHelpers } from 'formik';
import { useLocation } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { flexVariants } from '@monetr/interface/components/Flex';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import { layoutVariants } from '@monetr/interface/components/Layout';
import MForm from '@monetr/interface/components/MForm';
import SimpleFinSetupLayout from '@monetr/interface/components/setup/simplefin/SimpleFinSetupLayout';
import { SimpleFinSetupSteps } from '@monetr/interface/components/setup/simplefin/SimpleFinSetupSteps';
import SimpleFinSetupTokenLink from '@monetr/interface/components/setup/simplefin/SimpleFinSetupTokenLink';
import Typography from '@monetr/interface/components/Typography';
import SimpleFinLink from '@monetr/interface/models/SimpleFinLink';
import type { WithJsonValues } from '@monetr/interface/util/json';
import request, { type APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

export type SimpleFinSetupIntroValues = {
  name: string;
  setupToken: string;
};

export default function SimpleFinSetupIntro(): React.JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const [pathname, navigate] = useLocation();

  const initialValues: SimpleFinSetupIntroValues = {
    name: '',
    setupToken: '',
  };

  const submit = useCallback(
    (values: SimpleFinSetupIntroValues, helpers: FormikHelpers<SimpleFinSetupIntroValues>) => {
      helpers.setSubmitting(true);
      request<WithJsonValues<SimpleFinLink>>({
        method: 'POST',
        url: '/api/simplefin/link',
        data: {
          name: values.name,
          setupToken: values.setupToken,
        },
      })
        .then(result => new SimpleFinLink(result.data))
        .then(simpleFinLink => navigate(`${pathname}/${simpleFinLink.simpleFinLinkId}`))
        .catch((error: ApiError<APIError>) =>
          enqueueSnackbar(
            <div>
              <Typography size='sm'>{error?.response?.data?.error || 'Failed to create SimpleFIN link!'}</Typography>
              {Object.entries(error?.response?.data?.problems ?? {}).map(([key, problem]) => (
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
    [navigate, pathname, enqueueSnackbar],
  );

  return (
    <SimpleFinSetupLayout step={SimpleFinSetupSteps.Intro}>
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
          SimpleFIN Setup
        </Typography>
        <Typography align='center' color='subtle' size='lg'>
          Let us know what you want to call this budget, and paste in your SimpleFIN setup token.
        </Typography>
        <FormTextField<SimpleFinSetupIntroValues>
          autoComplete='off'
          autoFocus
          className={layoutVariants({ width: 'full' })}
          data-1p-ignore
          label='Budget Name'
          name='name'
          placeholder='SimpleFIN'
          required
        />
        <FormTextField<SimpleFinSetupIntroValues>
          autoComplete='off'
          className={layoutVariants({ width: 'full' })}
          data-1p-ignore
          label='Setup Token'
          labelDecorator={SimpleFinSetupTokenLink}
          name='setupToken'
          placeholder='Paste your setup token...'
          required
          type='password'
        />
        <FormButton type='submit' variant='primary'>
          Next
        </FormButton>
      </MForm>
    </SimpleFinSetupLayout>
  );
}

import React, { ForwardedRef } from 'react';
import { Formik, FormikConfig, FormikProps, FormikValues } from 'formik';

import { ReactElement } from './types';

interface MFormProps<Values extends FormikValues = FormikValues> extends FormikConfig<Values> {
  className?: string;
  children: ReactElement;
}

export type MFormRef = HTMLFormElement;

export default React.forwardRef<MFormRef, MFormProps<FormikValues>>(
  function MForm<Values extends FormikValues = FormikValues>(
    props: MFormProps<Values>,
    ref: ForwardedRef<MFormRef>,
  ): JSX.Element {
    const { className, children, ...formikConfig } = props;

    return (
      <Formik { ...formikConfig }>
        {(formik: FormikProps<Values>) => (
          <form
            onSubmit={ formik.handleSubmit }
            className={ className }
            data-testid={ props['data-testid'] }
            ref={ ref }
          >
            {children}
          </form>
        )}
      </Formik>
    );
  }
);


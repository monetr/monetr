import React from 'react';
import { Formik, FormikConfig, FormikProps, FormikValues } from 'formik';

import { ReactElement } from './types';

interface MFormProps<Values extends FormikValues = FormikValues> extends FormikConfig<Values> {
  className?: string;
  children: ReactElement;
  'data-testid'?: string;
}

export default function MForm<Values extends FormikValues = FormikValues>(props: MFormProps<Values>): JSX.Element {
  const { className, children, ...formikConfig } = props;

  return (
    <Formik { ...formikConfig }>
      {(formik: FormikProps<Values>) => (
        <form onSubmit={ formik.handleSubmit } className={ className } data-testid={ props['data-testid'] }>
          {children}
        </form>
      )}
    </Formik>
  );
}

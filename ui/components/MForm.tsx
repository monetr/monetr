import React from 'react';
import { Formik, FormikConfig, FormikProps, FormikValues } from 'formik';

import { ReactElement } from './types';

interface MFormProps<Values extends FormikValues = FormikValues> extends FormikConfig<Values> {
  className?: string;
  children: ReactElement;
}

export default function MForm<Values extends FormikValues = FormikValues>(props: MFormProps<Values>): JSX.Element {
  const { className, children, ...formikConfig } = props;

  return (
    <Formik { ...formikConfig }>
      { (formik: FormikProps<Values>) => (
        <form onSubmit={ formik.handleSubmit } className={ className }>
          { children }
        </form>
      )}
    </Formik>
  );
}

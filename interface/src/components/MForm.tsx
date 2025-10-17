import React, { type ForwardedRef } from 'react';
import { Formik, type FormikConfig, type FormikProps, type FormikValues } from 'formik';

interface MFormProps<Values extends FormikValues = FormikValues> extends FormikConfig<Values> {
  className?: string;
  children?: ((props: FormikProps<Values>) => React.ReactNode) | React.ReactNode;
}

export type MFormRef = HTMLFormElement;

export default React.forwardRef<MFormRef, MFormProps<FormikValues>>(function MForm<
  Values extends FormikValues = FormikValues,
>(props: MFormProps<Values>, ref: ForwardedRef<MFormRef>): JSX.Element {
  const { className, children, ...formikConfig } = props;

  return (
    <Formik {...formikConfig}>
      {(formik: FormikProps<Values>) => (
        <form onSubmit={formik.handleSubmit} className={className} data-testid={props['data-testid']} ref={ref}>
          {typeof children === 'function' ? children(formik) : children}
        </form>
      )}
    </Formik>
  );
});

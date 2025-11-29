import type React from 'react';
import { Formik, type FormikConfig, type FormikProps, type FormikValues } from 'formik';

interface MFormProps<Values extends FormikValues = FormikValues> extends FormikConfig<Values> {
  className?: string;
  children?: ((props: FormikProps<Values>) => React.ReactNode) | React.ReactNode;
}

export default function MForm<Values extends FormikValues = FormikValues>(
  props: MFormProps<Values>,
): React.JSX.Element {
  const { className, children, ...formikConfig } = props;
  return (
    <Formik<Values> {...formikConfig}>
      {(formik: FormikProps<Values>) => (
        <form className={className} data-testid={props['data-testid']} onSubmit={formik.handleSubmit}>
          {typeof children === 'function' ? children(formik) : children}
        </form>
      )}
    </Formik>
  );
}

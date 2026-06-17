import type React from 'react';
import { Formik, type FormikConfig, type FormikProps, type FormikValues } from 'formik';

interface MFormProps<Values extends FormikValues = FormikValues> extends FormikConfig<Values> {
  className?: string;
  children?: ((props: FormikProps<Values>) => React.ReactNode) | React.ReactNode;
  'data-testid'?: string;
  // Forwarded onto the underlying form element. The auth pages use this to warm up the proof of work challenge on the
  // first keystroke without every form having to wire up its own input handler.
  onInput?: React.ComponentProps<'form'>['onInput'];
}

export default function MForm<Values extends FormikValues = FormikValues>(
  props: MFormProps<Values>,
): React.JSX.Element {
  const { className, children, onInput, ...formikConfig } = props;
  return (
    <Formik<Values> {...formikConfig}>
      {(formik: FormikProps<Values>) => (
        <form className={className} data-testid={props['data-testid']} onInput={onInput} onSubmit={formik.handleSubmit}>
          {typeof children === 'function' ? children(formik) : children}
        </form>
      )}
    </Formik>
  );
}

import type { ComponentProps } from 'react';
import { useFormikContext } from 'formik';

import { Button } from '@monetr/interface/components/Button';

export type FormButtonProps = ComponentProps<typeof Button>;

export default function FormButton(props: FormButtonProps): JSX.Element {
  const formikContext = useFormikContext();
  props = {
    ...props,
    // disabled is true when we are submitting, when the form is not valid, or when the form has not even been touched
    // (but only when the initial values of the form are not valid). OR when the prop to disable the button is hard
    // coded to true.
    disabled:
      formikContext?.isSubmitting ||
      props?.disabled ||
      !formikContext?.isValid ||
      (!formikContext?.dirty && !formikContext?.isInitialValid),
    onSubmit: props?.onSubmit || (props.type === 'submit' ? formikContext?.submitForm : undefined),
  };
  return <Button {...props} />;
}

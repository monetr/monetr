import React, { type ComponentProps } from 'react';
import { useFormikContext } from 'formik';

import { Button } from '@monetr/interface/components/Button';

export type FormButtonProps = ComponentProps<typeof Button>;

export default function FormButton(props: FormButtonProps): JSX.Element {
  const formikContext = useFormikContext();
  props = {
    ...props,
    disabled: formikContext?.isSubmitting || props?.disabled,
    onSubmit: props?.onSubmit || (props.type === 'submit' ? formikContext?.submitForm : undefined),
  };
  return <Button {...props} />;
}

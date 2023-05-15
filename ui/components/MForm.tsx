import { useFormikContext } from "formik";
import React from "react";

type FormProps = React.DetailedHTMLProps<React.FormHTMLAttributes<HTMLFormElement>, HTMLFormElement>;

interface MFormProps extends FormProps {

}

export default function MForm(props: MFormProps): JSX.Element {
  const formikContext = useFormikContext();

  return (
    <form
      onSubmit={ formikContext?.handleSubmit }
      { ...props }
    />
  );
}

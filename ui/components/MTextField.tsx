import { TextField, TextFieldProps } from "@mui/material";
import { useFormikContext } from "formik";
import React from "react";

interface MTextFieldBaseProps {
  name: string;
}

export type MTextFieldProps = MTextFieldBaseProps & TextFieldProps;

export default function MTextField<T>(props: MTextFieldProps): JSX.Element {
  const formik = useFormikContext<T>();
  const touched = formik.touched[props.name];
  const value = formik.values[props.name];
  const error = formik.errors[props.name] as string;
  const hasError = Boolean(touched && value && !!error);
  const helperText = hasError ? error : '';

  return (
    <TextField
      { ...props }
      error={ hasError }
      helperText={ helperText }
      className="w-full"
      onChange={ formik.handleChange }
      onBlur={ formik.handleBlur }
      value={ value }
      disabled={ formik.isSubmitting || props.disabled }
    />
  )
}

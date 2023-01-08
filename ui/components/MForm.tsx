import { useFormikContext } from "formik";
import React from "react";

export type MFormProps = React.DetailedHTMLProps<React.FormHTMLAttributes<HTMLFormElement>, HTMLFormElement>

export default function MForm<T>(props: MFormProps): JSX.Element {
  const formik = useFormikContext<T>();
  return (
    <form { ...props } onSubmit={ formik.handleSubmit }/>
  )
}

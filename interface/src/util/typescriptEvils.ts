/**
 * Given a functional component `T` with props `P` this will return a type representing `P`.
 */
export type ExtractProps<T> = T extends React.FC<infer P> ? P : never;

export type FormFieldName<T, V> = Extract<
  {
    [K in keyof T]-?: T[K] extends V ? K : never;
  }[keyof T],
  string
>;

export type FormValues<T> = {
  [P in keyof T]: T[P];
};

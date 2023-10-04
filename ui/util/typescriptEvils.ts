/**
 * Given a functional component `T` with props `P` this will return a type representing `P`.
 */
export type ExtractProps<T> = T extends React.FC<infer P> ? P : never;

import type React from 'react';

export type HrProps = React.ComponentPropsWithoutRef<'hr'>;

export function Hr({ style, ...props }: HrProps) {
  return <hr style={style} {...props} />;
}

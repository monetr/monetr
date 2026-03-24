import type React from 'react';

export type HrProps = React.ComponentPropsWithoutRef<'hr'>;

export function Hr({ style, ...props }: HrProps) {
  return (
    <hr
      style={{
        width: '100%',
        border: 'none',
        borderTop: '1px solid #eaeaea',
        ...style,
      }}
      {...props}
    />
  );
}

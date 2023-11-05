import React from 'react';
import { Badge, styled } from '@mui/material';

import MerchantIcon, { MerchantIconProps } from './MerchantIcon';
import useTheme from '@monetr/interface/hooks/useTheme';

export interface TransactionMerchantIconProps extends MerchantIconProps {
  pending?: boolean;
}

export default function TransactionMerchantIcon(props: TransactionMerchantIconProps): JSX.Element {
  const theme = useTheme();
  const { pending, ...merchantIconProps } = props;
  const StyledBadge = styled(Badge)(() => ({
    '& .MuiBadge-badge': {
      backgroundColor: theme.tailwind.colors['blue']['500'],
      color:  theme.tailwind.colors['blue']['500'],
      boxShadow: `0 0 0 2px ${theme.tailwind.colors['dark-monetr']['background']['DEFAULT']}`,
      '&::after': {
        position: 'absolute',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        borderRadius: '50%',
        animation: 'ripple-pending-txn 1.2s infinite ease-in-out',
        border: '1px solid currentColor',
        content: '""',
      },
    },
    '@keyframes ripple-pending-txn': {
      '0%': {
        transform: 'scale(.8)',
        opacity: 1,
      },
      '100%': {
        transform: 'scale(2.4)',
        opacity: 0,
      },
    },
  }));

  if (pending) {
    return (
      <StyledBadge
        overlap='circular'
        anchorOrigin={ { vertical: 'bottom', horizontal: 'right' } }
        variant='dot'
      >
        <MerchantIcon { ...merchantIconProps } />
      </StyledBadge>
    );
  }

  return (
    <MerchantIcon { ...merchantIconProps } />
  );

}

import { Avatar } from '@mui/material';
import Transaction from 'models/Transaction';
import React from 'react';

interface Props {
  transaction: Transaction;
}

export default function TransactionIcon(props: Props): JSX.Element {
  const letter = props.transaction.name.toUpperCase().charAt(0);

  return (
    <Avatar>
      { letter }
    </Avatar>
  )
}

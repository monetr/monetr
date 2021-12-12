import BalanceNavDisplay from 'components/Balance/BalanceNavDisplay';
import React from 'react';
import testRenderer from 'testutils/renderer';
import { screen } from '@testing-library/react';

describe('Balance Nav Display', () => {

  it('will not render without balance', () => {
    testRenderer(<BalanceNavDisplay/>);
    expect(screen.queryByText('Safe-To-Spend')).not.toBeInTheDocument();
  });

});

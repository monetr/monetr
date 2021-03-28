import { render } from '@testing-library/react'
import { TransactionItem } from "components/Transactions/TransactionItem";
import Spending from "data/Spending";
import Transaction from "data/Transaction";
import moment from "moment";
import { queryText } from "testutils/queryText";

TransactionItem.defaultProps = {
  isSelected: false,
  selectTransaction: jest.fn(),
};

describe('transaction item', () => {
  it('will render', () => {
    render(<TransactionItem
      transactionId={ 123 }
      transaction={ new Transaction({
        name: 'Dumb Stuff',
        date: moment(),
      }) }
    />);

    // Make sure it's actually there.
    expect(document.querySelector('.transactions-item')).not.toBeEmptyDOMElement();
    expect(queryText('.transaction-item-name')).toBe('Dumb Stuff');
  });

  it('will render with expense', () => {
    render(<TransactionItem
      transactionId={ 123 }
      transaction={ new Transaction({
        name: 'Dumb Stuff',
        date: moment(),
      }) }
      spending={ new Spending({
        name: 'Dumb Stuff Budget'
      }) }
    />);

    expect(queryText('.transaction-expense-name')).toBe('Spent From Dumb Stuff Budget');
  });

  it('will render a deposit', () => {
    render(<TransactionItem
      transactionId={ 123 }
      transaction={ new Transaction({
        name: 'Dumb Stuff',
        date: moment(),
        amount: -100, // $1.00
      }) }
    />);

    expect(queryText('.transaction-expense-name')).toBe('Deposited Into Safe-To-Spend');
    expect(queryText('.amount')).toBe('+ $1.00');
  });
});

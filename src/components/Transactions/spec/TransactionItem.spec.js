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

  it('will render a deposit', () => {
    render(<TransactionItem
      transactionId={ 123 }
      transaction={ new Transaction({
        name: 'Dumb Stuff',
        date: moment(),
        amount: -100, // $1.00
      }) }
    />);

    expect(queryText('.transaction-expense-name')).toBe("");
    expect(queryText('.amount')).toBe('+ $1.00');
  });
});

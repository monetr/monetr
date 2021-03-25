import { TransactionItem } from "components/Transactions/TransactionItem";
import Transaction from "data/Transaction";
import moment from "moment";
import { render, screen } from '@testing-library/react'


describe('transaction item', () => {
  it('will render', () => {
    const item = render(<TransactionItem
      transactionId={ 123 }
      transaction={ new Transaction({
        name: 'Dumb Stuff',
        date: moment(),
      }) }
    />);

    // Make sure it's actually there.
    expect(document.querySelector('.transactions-item')).not.toBeEmptyDOMElement();
    expect(document.querySelector('.transaction-item-name').textContent).toBe('Dumb Stuff');
  });
});

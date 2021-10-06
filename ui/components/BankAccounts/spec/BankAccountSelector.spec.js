import { render } from '@testing-library/react'
import { BankAccountSelector } from "components/BankAccounts/BankAccountSelector";
import { Map } from 'immutable';

describe('bank account selector', () => {
  it('will not render when bank accounts are loading', () => {
    const bankSelector = render(<BankAccountSelector
      selectedBankAccountId={ null }
      setSelectedBankAccountId={ jest.fn() }
      bankAccounts={ new Map() }
      bankAccountsLoading={ true }
      links={ new Map() }
      linksLoading={ false }
    />);

    expect(bankSelector.container).toBeEmptyDOMElement();
  });

  it('will not render when links are loading', () => {
    const bankSelector = render(<BankAccountSelector
      selectedBankAccountId={ null }
      setSelectedBankAccountId={ jest.fn() }
      bankAccounts={ new Map() }
      bankAccountsLoading={ false }
      links={ new Map() }
      linksLoading={ true }
    />);

    expect(bankSelector.container).toBeEmptyDOMElement();
  });

  it('will not render when both links and bank accounts are loading', () => {
    const bankSelector = render(<BankAccountSelector
      selectedBankAccountId={ null }
      setSelectedBankAccountId={ jest.fn() }
      bankAccounts={ new Map() }
      bankAccountsLoading={ true }
      links={ new Map() }
      linksLoading={ true }
    />);

    expect(bankSelector.container).toBeEmptyDOMElement();
  });
});

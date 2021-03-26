import { combineReducers } from "redux";
import authentication from 'shared/authentication/reducer';
import bankAccounts from 'shared/bankAccounts/reducer';
import bootstrap from 'shared/bootstrap/reducer';
import spending from 'shared/spending/reducer';
import links from 'shared/links/reducer';
import transactions from 'shared/transactions/reducer';

export default combineReducers({
  authentication,
  bankAccounts,
  bootstrap,
  spending,
  links,
  transactions,
});



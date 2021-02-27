import { combineReducers } from "redux";
import authentication from 'shared/authentication/reducer';
import bootstrap from 'shared/bootstrap/reducer';
import links from 'shared/links/reducer';
import bankAccounts from 'shared/bankAccounts/reducer';
import transactions from 'shared/transactions/reducer';

export default combineReducers({
  authentication,
  bankAccounts,
  bootstrap,
  links,
  transactions,
});



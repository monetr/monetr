import Expense from "data/Expense";
import { Map } from 'immutable';

export default class ExpenseState {
  items: Map<number, Map<number, Expense>>;
  loaded: boolean;
  loading: boolean;

  constructor() {
    this.items = Map<number, Map<number, Expense>>();
  }
}

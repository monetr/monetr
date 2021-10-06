import Balance from 'data/Balance';
import { Map } from 'immutable';

export default class BalancesState {
  items: Map<number, Balance>;
  loaded: boolean;
  loading: boolean;

  constructor() {
    this.items = Map<number, Balance>();
  }
}

import Balance from 'models/Balance';
import { Map } from 'immutable';

export default class BalancesState {
  items: Map<number, Balance>;
  loaded: boolean;
  loading: boolean;

  constructor() {
    this.items = Map<number, Balance>();
  }
}

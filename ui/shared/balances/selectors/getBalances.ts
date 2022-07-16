import { Map } from 'immutable';
import Balance from 'models/Balance';
import { AppState } from 'store';

export const getBalances = (state: AppState): Map<number, Balance> => state.balances.items;

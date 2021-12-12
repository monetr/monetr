import Balance from 'models/Balance';
import { AppState } from 'store';
import { Map } from 'immutable';

export const getBalances = (state: AppState): Map<number, Balance> => state.balances.items;

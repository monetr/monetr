import { Moment } from 'moment';
import { Plan } from 'shared/bootstrap/state';
import { AppState } from 'store';

export const getIsBootstrapped = (state: AppState): boolean => state.bootstrap.isReady;

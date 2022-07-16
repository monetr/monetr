import { Map } from 'immutable';
import BankAccount from 'models/BankAccount';
import FundingSchedule from 'models/FundingSchedule';
import { Logout } from 'shared/authentication/actions';
import { CreateFundingSchedule, FetchFundingSchedules, FundingScheduleActions } from 'shared/fundingSchedules/actions';
import FundingScheduleState from 'shared/fundingSchedules/state';
import { RemoveLink } from 'shared/links/actions';

export default function reducer(state: FundingScheduleState = new FundingScheduleState(), action: FundingScheduleActions): FundingScheduleState {
  switch (action.type) {
    case CreateFundingSchedule.Request:
    case FetchFundingSchedules.Request:
      return {
        ...state,
        loading: true,
      };
    case FetchFundingSchedules.Failure:
    case CreateFundingSchedule.Failure:
      return {
        ...state,
        loading: false,
      };
    case FetchFundingSchedules.Success:
      return {
        ...state,
        loading: false,
        items: state.items.mergeDeep(action.payload),
      };
    case CreateFundingSchedule.Success:
      // With create funding schedule the payload will be a single funding schedule that was just created. So we can
      // just do a set in for the current items to add the new schedule.
      return {
        ...state,
        loading: false,
        items: state.items.setIn([
          action.payload.bankAccountId,
          action.payload.fundingScheduleId,
        ], action.payload),
      };
    case RemoveLink.Success:
      return {
        ...state,
        loading: false,
        // This is a bit goofy. Basically when we remove a link we are returned the link itself, and all of the bank
        // accounts associated with that link. We then look at all of our funding schedules, but only return those that
        // do not belong to bank accounts that are in that payload. The ones in the payload are being removed.
        items: state.items.filter((fundingSchedule: Map<number, FundingSchedule>, bankAccountId: number): boolean => {
          return !action.payload.bankAccounts.find((bankAccount: BankAccount) => bankAccount.linkId === bankAccountId);
        }),
      };
    case Logout.Success:
      return new FundingScheduleState();
    default:
      return state;
  }
}

import FundingSchedule from "data/FundingSchedule";
import { Map } from 'immutable';


export default class FundingScheduleState {
  items: Map<number, Map<number, FundingSchedule>>;
  loaded: boolean;
  loading: boolean;

  constructor() {
    this.items = Map<number, Map<number, FundingSchedule>>();
  }
}

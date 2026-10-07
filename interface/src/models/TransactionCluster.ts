import { idPrefix } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class TransactionCluster {
  readonly [idPrefix] = 'tcl';

  readonly transactionClusterId: string;
  readonly bankAccountId: string;
  name: string;
  // originalMemo is the memo the transactions in this group came in with from the bank, can be blank
  originalMemo: string;
  members: Array<string>;
  createdAt: Date;

  constructor(data: WithJsonValues<TransactionCluster>) {
    this.transactionClusterId = data.transactionClusterId;
    this.bankAccountId = data.bankAccountId;
    this.name = data.name;
    this.originalMemo = data.originalMemo;
    this.members = data.members;
    this.createdAt = parseDate(data.createdAt);
  }
}

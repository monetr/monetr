import parseDate from '@monetr/interface/util/parseDate';

export default class TransactionCluster {
  transactionClusterId: string;
  bankAccountId: string;
  name: string;
  members: Array<string>;
  createdAt: Date;

  constructor(data?: Partial<TransactionCluster>) {
    if (data) {
      Object.assign(this, {
        ...data,
        createdAt: parseDate(data?.createdAt),
      });
    } else {
      Object.assign(this, {
        members: [],
      });
    }
  }
}

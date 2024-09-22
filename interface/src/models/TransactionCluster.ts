import { parseJSON } from 'date-fns';

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
        createdAt: data.createdAt ?? parseJSON(data.createdAt),
      });
    } else {
      Object.assign(this, {
        members: [],
      });
    }
  }
}

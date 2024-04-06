import { parseJSON } from 'date-fns';

export default class TransactionCluster {
  transactionClusterId: string;
  bankAccountId: number;
  name: string;
  members: Array<number>;
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

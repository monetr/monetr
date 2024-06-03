import { parseJSON } from 'date-fns';

export enum TransactionUploadStatus {
  Pending = 'pending',
  Processing = 'processing',
  Failed = 'failed',
  Complete = 'complete'
}

export default class TransactionUpload {
  transactionUploadId: string;
  bankAccountId: string;
  fileId: string;
  status: TransactionUploadStatus;
  error: string | null;
  createdAt: Date;
  createdBy: string;
  processedAt: Date | null;
  completedAt: Date | null;

  constructor(data?: Partial<TransactionUpload>) {
    if (data) Object.assign(this, {
      ...data,
      createdAt: data?.createdAt && parseJSON(data?.createdAt),
      processedAt: data?.processedAt && parseJSON(data?.processedAt),
      completedAt: data?.completedAt && parseJSON(data?.completedAt),
    });
  }
}

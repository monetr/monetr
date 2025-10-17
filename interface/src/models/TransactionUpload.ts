import type MonetrFile from './File';
import parseDate from '@monetr/interface/util/parseDate';

export enum TransactionUploadStatus {
  Pending = 'pending',
  Processing = 'processing',
  Failed = 'failed',
  Complete = 'complete',
}

export default class TransactionUpload {
  transactionUploadId: string;
  bankAccountId: string;
  fileId: string;
  file?: MonetrFile;
  status: TransactionUploadStatus;
  error: string | null;
  createdAt: Date;
  createdBy: string;
  processedAt: Date | null;
  completedAt: Date | null;

  constructor(data?: Partial<TransactionUpload>) {
    if (data) {
      Object.assign(this, {
        ...data,
        createdAt: parseDate(data?.createdAt),
        processedAt: parseDate(data?.processedAt),
        completedAt: parseDate(data?.completedAt),
      });
    }
  }
}

import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

import MonetrFile from './File';

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

  constructor(data: WithJsonValues<TransactionUpload>) {
    this.transactionUploadId = data.transactionUploadId;
    this.bankAccountId = data.bankAccountId;
    this.fileId = data.fileId;
    this.file = data.file ? new MonetrFile(data.file) : undefined;
    this.status = data.status;
    this.error = data.error ?? null;
    this.createdAt = parseDate(data.createdAt);
    this.createdBy = data.createdBy;
    this.processedAt = parseDate(data.processedAt);
    this.completedAt = parseDate(data.completedAt);
  }
}

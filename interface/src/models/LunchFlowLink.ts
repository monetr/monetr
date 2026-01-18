import parseDate from '@monetr/interface/util/parseDate';

export enum LunchFlowLinkStatus {
  Active = 'active',
  Deactivated = 'deactivated',
  Error = 'error',
}

export default class LunchFlowLink {
  lunchFlowLinkId: string;
  apiUrl: string;
  status: LunchFlowLinkStatus;
  lastManualSync?: Date;
  lastSuccessfulUpdate?: Date;
  lastAttemptedUpdate?: Date;
  updatedAt: Date;
  createdAt: Date;
  deletedAt?: Date;
  createdBy: string;

  constructor(data?: Partial<LunchFlowLink>) {
    if (data) {
      Object.assign(this, {
        ...data,
        lastManualSync: parseDate(data?.lastManualSync),
        lastSuccessfulUpdate: parseDate(data?.lastSuccessfulUpdate),
        lastAttemptedUpdate: parseDate(data?.lastAttemptedUpdate),
        updatedAt: parseDate(data?.updatedAt),
        createdAt: parseDate(data?.createdAt),
        deletedAt: parseDate(data?.deletedAt),
      });
    }
  }
}

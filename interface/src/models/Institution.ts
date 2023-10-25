import { parseJSON } from 'date-fns';

export default class Institution {
  name: string;
  url: string | null;
  primaryColor: string | null;
  logo: string | null;
  status: InstitutionStatus;
  readonly timestamp: Date;

  constructor(data?: Partial<Institution>) {
    if (data) {
      Object.assign(this, {
        ...data,
        status: new InstitutionStatus(data.status),
        timestamp: new Date(),
      });
    }
  }
}

export class InstitutionStatus {
  transactions_updates: PlaidProductStatus;
  plaidIncidents: InstitutionPlaidIncident[];

  constructor(data?: Partial<InstitutionStatus>) {
    if (data) {
      Object.assign(this, {
        ...data,
        plaidIncidents: (data?.plaidIncidents || []).map(item => new InstitutionPlaidIncident(item)),
      });
    }
  }
}

export type PlaidStatus = 'HEALTHY' | 'DEGRADED' | 'DOWN';
export type RefreshInterval = 'DELAYED' | 'STOPPED';

export class PlaidProductStatus {
  status: PlaidStatus;
  last_status_change: Date;
  breakdown: {
    success: number;
    error_plaid: number;
    error_institution: number;
    refresh_interval: RefreshInterval | null;
  };
}

export class InstitutionPlaidIncident {
  start: Date;
  end: Date | null;
  title: string;

  constructor(data?: Partial<InstitutionPlaidIncident>) {
    if (data) {
      Object.assign(this, {
        ...data,
        start: parseJSON(data.start),
        end: data.end && parseJSON(data.end),
      });
    }
  }
}

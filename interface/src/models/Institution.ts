import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class Institution {
  name: string;
  url: string | null;
  primaryColor: string | null;
  logo: string | null;
  status: InstitutionStatus | null;
  // timestamp is not returned by the API, we stamp it client side when we hydrate the model so we know roughly how
  // stale the data we are looking at is.
  readonly timestamp: Date;

  constructor(data: Omit<WithJsonValues<Institution>, 'timestamp'>) {
    this.name = data.name;
    this.url = data.url ?? null;
    this.primaryColor = data.primaryColor ?? null;
    this.logo = data.logo ?? null;
    this.status = data.status ? new InstitutionStatus(data.status) : null;
    this.timestamp = new Date();
  }
}

export class InstitutionStatus {
  transactions_updates: PlaidProductStatus | null;
  plaidIncidents: InstitutionPlaidIncident[];

  constructor(data: WithJsonValues<InstitutionStatus>) {
    // transactions_updates is a plain object rather than its own model, but it carries a date so we cannot just copy it
    // across. Rebuild it so last_status_change is an actual Date.
    this.transactions_updates = data.transactions_updates
      ? {
          status: data.transactions_updates.status,
          last_status_change: parseDate(data.transactions_updates.last_status_change),
          breakdown: data.transactions_updates.breakdown,
        }
      : null;
    this.plaidIncidents = (data.plaidIncidents ?? []).map(item => new InstitutionPlaidIncident(item));
  }
}

export type PlaidStatus = 'HEALTHY' | 'DEGRADED' | 'DOWN';
export type RefreshInterval = 'NORMAL' | 'DELAYED' | 'STOPPED';

export interface PlaidProductStatus {
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

  constructor(data: WithJsonValues<InstitutionPlaidIncident>) {
    this.start = parseDate(data.start);
    this.end = parseDate(data.end);
    this.title = data.title;
  }
}

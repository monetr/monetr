import { ID, idPrefix } from '@monetr/interface/models/ID';
import type User from '@monetr/interface/models/User';
import type { WithJsonValues } from '@monetr/interface/util/json';
import parseDate from '@monetr/interface/util/parseDate';

export default class ApiKey {
  readonly [idPrefix] = 'key';

  readonly apiKeyId: ID<ApiKey>;
  readonly name: string;
  readonly createdAt: Date;
  readonly createdBy: ID<User>;
  readonly updatedAt: Date;
  readonly deletedAt: Date | null;

  constructor(data: WithJsonValues<ApiKey>) {
    this.apiKeyId = ID.from(data.apiKeyId);
    this.name = data.name;
    this.createdAt = parseDate(data.createdAt);
    this.createdBy = ID.from(data.createdBy);
    this.updatedAt = parseDate(data.updatedAt);
    this.deletedAt = data.deletedAt ? parseDate(data.deletedAt) : null;
  }
}

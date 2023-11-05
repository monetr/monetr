
import { addDays, startOfDay } from 'date-fns';

import Institution, { InstitutionPlaidIncident, InstitutionStatus } from '@monetr/interface/models/Institution';

describe('InstitutionPlaidIncident', () => {
  it('will construct with no end', () => {
    const start = startOfDay(new Date());
    const incident = new InstitutionPlaidIncident({
      start: start,
      end: null,
      title: 'This institution is currently experiencing problems.',
    });

    expect(incident.start.toISOString()).toBe(start.toISOString());
    expect(incident.end).toBeNull();
    expect(incident.title).toBe('This institution is currently experiencing problems.');
  });

  it('will construct with an end', () => {
    const start = startOfDay(new Date());
    const end = addDays(start, 1);
    const incident = new InstitutionPlaidIncident({
      start: start,
      end: end,
      title: 'This institution was experiencing problems.',
    });

    expect(incident.start.toISOString()).toBe(start.toISOString());
    expect(incident.end.toISOString()).toBe(end.toISOString());
    expect(incident.title).toBe('This institution was experiencing problems.');
  });

  it('will properly parse a generic object', () => {
    const start = startOfDay(new Date());
    const end = addDays(start, 1);
    const input = {
      start,
      end,
      title: 'This institution was experiencing problems.',
    };

    // @ts-ignore We want to be able to pass a slightly invalid object date wise and make sure it gets parsed.
    const result = new InstitutionPlaidIncident(input);
    expect(result.start.toISOString()).toBe(start.toISOString());
    expect(result.end.toISOString()).toBe(end.toISOString());
    expect(result.title).toBe('This institution was experiencing problems.');
  });
});

describe('InstitutionStatus', () => {
  it('will create without any incidents', () => {
    const result = new InstitutionStatus({
      transactions_updates: {
        status: 'HEALTHY',
        last_status_change: new Date(),
        breakdown: {
          success: 0.9,
          error_institution: 0.0,
          error_plaid: 0.0,
          refresh_interval: null,
        },
      },
      plaidIncidents: [],
    });

    expect(result.transactions_updates.status).toEqual('HEALTHY');
    expect(result.plaidIncidents).toHaveLength(0);
  });

  it('will create with some incidents', () => {
    const result = new InstitutionStatus({
      plaidIncidents: [
        {
          start: addDays(new Date(), -2),
          end: startOfDay(new Date()),
          title: 'There was a problem.',
        },
        {
          start: startOfDay(new Date()),
          end: null,
          title: 'There is an ongoing problem.',
        },
      ],
    });

    expect(result.plaidIncidents).toHaveLength(2);
    expect(result.plaidIncidents[0] instanceof InstitutionPlaidIncident).toBeTruthy();
    expect(result.plaidIncidents[1] instanceof InstitutionPlaidIncident).toBeTruthy();
  });
});

describe('Institutions', () => {
  it('create', () => {
    const item = new Institution({
      name: 'U.S. Bank',
      url: 'https://www.usbank.com/',
      primaryColor: 'blue',
      logo: 'I am a picture',
    });

    expect(item.name).toBe('U.S. Bank');
    expect(item.url).toBe('https://www.usbank.com/');
    expect(item.primaryColor).toBe('blue');
    expect(item.logo).toBe('I am a picture');
  });
});

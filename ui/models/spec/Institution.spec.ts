import Institution, { InstitutionPlaidIncident, InstitutionStatus } from 'models/Institution';
import moment from 'moment';
import { instanceOf } from 'prop-types';

describe('InstitutionPlaidIncident', () => {
  it('will construct with no end', () => {
    const start = moment().startOf('day');
    const incident = new InstitutionPlaidIncident({
      start: start,
      end: null,
      title: 'This institution is currently experiencing problems.'
    });

    expect(incident.start.toISOString()).toBe(start.toISOString());
    expect(incident.end).toBeNull();
    expect(incident.title).toBe('This institution is currently experiencing problems.');
  });

  it('will construct with an end', () => {
    const start = moment().startOf('day');
    const end = start.add(1, 'day');
    const incident = new InstitutionPlaidIncident({
      start: start,
      end: end,
      title: 'This institution was experiencing problems.'
    });

    expect(incident.start.toISOString()).toBe(start.toISOString());
    expect(incident.end.toISOString()).toBe(end.toISOString());
    expect(incident.title).toBe('This institution was experiencing problems.');
  });

  it('will properly parse a generic object', () => {
    const start = moment().startOf('day');
    const end = start.add(1, 'day');
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

  it('what happens with no start', () => {
    expect(() => new InstitutionPlaidIncident({
      title: 'This institution was experiencing problems.'
    })).toThrow('input to mustParseToMoment was not a valid date time');
  });
});

describe('InstitutionStatus', () => {
  it('will create without any incidents', () => {
    const result = new InstitutionStatus({
      login: true,
      transactions: false,
      balance: true,
      plaidIncidents: [],
    });

    expect(result.login).toBeTruthy();
    expect(result.transactions).toBeFalsy();
    expect(result.balance).toBeTruthy();
    expect(result.plaidIncidents).toHaveLength(0);
  });

  it('will create with some incidents', () => {
    const result = new InstitutionStatus({
      login: true,
      transactions: false,
      balance: true,
      plaidIncidents: [
        {
          start: moment().add(-2, 'day'),
          end: moment().startOf('day'),
          title: 'There was a problem.',
        },
        {
          start: moment().startOf('day'),
          end: null,
          title: 'There is an ongoing problem.',
        }
      ],
    });

    expect(result.login).toBeTruthy();
    expect(result.transactions).toBeFalsy();
    expect(result.balance).toBeTruthy();
    expect(result.plaidIncidents).toHaveLength(2);
    expect(result.plaidIncidents[0] instanceof InstitutionPlaidIncident).toBeTruthy();
    expect(result.plaidIncidents[1] instanceof InstitutionPlaidIncident).toBeTruthy();
  });
});

describe('Institutions', function () {
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
import moment from 'moment';

export const APIDateFormat = 'YYYY-MM-DDTHH:mm:ssZ';

export function mustParseToMoment(input: string | moment.Moment): moment.Moment {
  const result = moment(input, APIDateFormat);
  if (result.isValid()) {
    return result;
  }

  throw new Error('input to mustParseToMoment was not a valid date time');
}

export function parseToMomentMaybe(input: string | moment.Moment | null): moment.Moment | null {
  if (input) {
    return mustParseToMoment(input);
  }

  return null;
}

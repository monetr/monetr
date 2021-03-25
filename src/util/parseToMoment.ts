import moment from "moment";

export const APIDateFormat = "YYYY-MM-DDTHH:mm:ss.SSSSSSZ";

export function parseToMoment(input: string|moment.Moment): moment.Moment {
  return moment(input, APIDateFormat)
}

export function parseToMomentMaybe(input?: string|moment.Moment): moment.Moment|null {
  if (input) {
    return parseToMoment(input);
  }

  return null;
}

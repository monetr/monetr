import moment from "moment";

export const APIDateFormat = "YYYY-MM-DDTHH:mm:ss.SSSSSSZ";

export default function parseToMoment(input: string): moment.Moment {
  return moment(input, APIDateFormat)
}

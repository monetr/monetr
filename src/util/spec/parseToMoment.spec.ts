import parseToMoment from "util/parseToMoment";
import Moment from "moment";

describe('parse to moment', () => {
  it('will parse api date to moment', () => {
    const input = '2021-02-22T00:00:00Z';
    const result = parseToMoment(input);
    expect(result instanceof Moment).toBeTruthy();
    expect(result.utc().format('YYYY-MM-DD')).toBe('2021-02-22');
  });
});

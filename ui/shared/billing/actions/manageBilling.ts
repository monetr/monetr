import request from 'shared/util/request';

export default function manageBilling(): Promise<void> {
  return request().get('/billing/portal')
    .then(result => window.location.assign(result.data.url));
}

export const getSentryUser = (state: object) => {
  return {
    // @ts-ignore
    id: state.authentication.user.userId,
  };
}

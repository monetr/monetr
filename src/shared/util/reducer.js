
export default function reducer(state = {}, action) {
  switch (action.type) {
    case 'API_CLIENT_SETUP':
      return {
        ...state,
        setup: true,
      };


  }
  return state;
}

import { useEffect } from 'react';

// On mobile the window itself is the scroll container, and wouter does not reset the scroll position when the route
// changes. So navigating forward from a scrolled page would carry that scroll position over to the next page, clamped
// to however tall the next page happens to be. wouter monkey patches `history.pushState` to dispatch a `pushState`
// event on the window, so we listen for that and reset the scroll position for forward navigation only. Back and
// forward buttons go through `popstate` instead, which leaves the browser's native scroll restoration alone.
export default function ScrollToTopOnNavigate(): null {
  useEffect(() => {
    const scrollToTop = () => window.scrollTo(0, 0);
    window.addEventListener('pushState', scrollToTop);
    return () => window.removeEventListener('pushState', scrollToTop);
  }, []);

  return null;
}

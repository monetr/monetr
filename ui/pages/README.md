# Routes

This folder contains the route contents for the UI. It is structured such that the path to the component is
representative of the URL path that is used in the application. This is not done automatically like it is in next.js,
but this pattern is copied from next.js because of its intuitiveness.

---

Paths which do not have any subpaths should not use an `index.tsx` file. For example; if you have `/password/reset` and
`/password/forgot`, these should be represented as `/password/reset.tsx` and `/password/forgot.tsx`. The only time an
`index.tsx` should be used is if there is another immediate subpath. For example; `/verify/email` and
`/verify/email/resend` should be represented as `/verify/email/index.tsx` and `/verify/email/resend.tsx`.

Unlike all the other component files, these files should be named as they are represented in their URL path. If a path
requires any component or pattern that is being re-used in other areas. That path should have a "Page" component here,
but have all of its contents represented sin the `components` directory as "Views".

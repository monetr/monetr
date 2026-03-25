// SSR entry: re-exports the template registry for the node build environment.
// The rsbuild plugin imports this bundle after compilation to render each template.
export { templates } from '../templates';

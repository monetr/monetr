// This file is no longer used. The build pipeline is now handled entirely by rsbuild:
//
//   rsbuild build              — builds email templates to dist/emails/
//   EMAIL_OUT_DIR=/path rsbuild build  — builds to a custom output directory
//   rsbuild dev                — starts the dev preview server
//
// See rsbuild.config.ts and src/build/rsbuildPluginEmail.ts for the implementation.

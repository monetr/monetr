import mockAxios from "jest-mock-axios";

module.export = global.CONFIG = {
    BOOTSTRAP_CONFIG_JSON: false,
    USE_LOCAL_STORAGE: false,
    COOKIE_DOMAIN: 'app.monetr.mini',
    ENVIRONMENT: "Testing",
    API_URL: "https://app.monetr.mini",
    API_DOMAIN: "app.monetr.mini",
    SENTRY_DSN: null,
};

// When we are testing, make sure that our API calls are routed through our mock interface.
window.API = mockAxios;

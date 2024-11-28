import meta from "../../../src/pages/_meta.ts";
import documentation_meta from "../../../src/pages/documentation/_meta.ts";
import documentation_development_meta from "../../../src/pages/documentation/development/_meta.ts";
import documentation_use_meta from "../../../src/pages/documentation/use/_meta.ts";
import policy_meta from "../../../src/pages/policy/_meta.ts";
export const pageMap = [{
  data: meta
}, {
  name: "about",
  route: "/about",
  frontMatter: {
    "title": "About"
  }
}, {
  name: "blog",
  route: "/blog",
  frontMatter: {
    "title": "Blog"
  }
}, {
  name: "contact",
  route: "/contact",
  frontMatter: {
    "sidebarTitle": "Contact"
  }
}, {
  name: "documentation",
  route: "/documentation",
  children: [{
    data: documentation_meta
  }, {
    name: "configure",
    route: "/documentation/configure",
    children: [{
      name: "captcha",
      route: "/documentation/configure/captcha",
      frontMatter: {
        "title": "ReCAPTCHA"
      }
    }, {
      name: "cors",
      route: "/documentation/configure/cors",
      frontMatter: {
        "title": "CORS"
      }
    }, {
      name: "email",
      route: "/documentation/configure/email",
      frontMatter: {
        "sidebarTitle": "Email"
      }
    }, {
      name: "kms",
      route: "/documentation/configure/kms",
      frontMatter: {
        "title": "Key Management"
      }
    }, {
      name: "links",
      route: "/documentation/configure/links",
      frontMatter: {
        "sidebarTitle": "Links"
      }
    }, {
      name: "logging",
      route: "/documentation/configure/logging",
      frontMatter: {
        "sidebarTitle": "Logging"
      }
    }, {
      name: "plaid",
      route: "/documentation/configure/plaid",
      frontMatter: {
        "sidebarTitle": "Plaid"
      }
    }, {
      name: "postgres",
      route: "/documentation/configure/postgres",
      frontMatter: {
        "sidebarTitle": "Postgres"
      }
    }, {
      name: "redis",
      route: "/documentation/configure/redis",
      frontMatter: {
        "sidebarTitle": "Redis"
      }
    }, {
      name: "security",
      route: "/documentation/configure/security",
      frontMatter: {
        "sidebarTitle": "Security"
      }
    }, {
      name: "sentry",
      route: "/documentation/configure/sentry",
      frontMatter: {
        "sidebarTitle": "Sentry"
      }
    }, {
      name: "server",
      route: "/documentation/configure/server",
      frontMatter: {
        "sidebarTitle": "Server"
      }
    }, {
      name: "storage",
      route: "/documentation/configure/storage",
      frontMatter: {
        "sidebarTitle": "Storage"
      }
    }]
  }, {
    name: "configure",
    route: "/documentation/configure",
    frontMatter: {
      "title": "Configuration",
      "description": "Configure self-hosted monetr servers"
    }
  }, {
    name: "development",
    route: "/documentation/development",
    children: [{
      data: documentation_development_meta
    }, {
      name: "build",
      route: "/documentation/development/build",
      frontMatter: {
        "sidebarTitle": "Build"
      }
    }, {
      name: "code_of_conduct",
      route: "/documentation/development/code_of_conduct",
      frontMatter: {
        "sidebarTitle": "Code of Conduct"
      }
    }, {
      name: "credentials",
      route: "/documentation/development/credentials",
      frontMatter: {
        "sidebarTitle": "Credentials"
      }
    }, {
      name: "documentation",
      route: "/documentation/development/documentation",
      frontMatter: {
        "sidebarTitle": "Documentation"
      }
    }, {
      name: "local_development",
      route: "/documentation/development/local_development",
      frontMatter: {
        "sidebarTitle": "Local Development"
      }
    }]
  }, {
    name: "development",
    route: "/documentation/development",
    frontMatter: {
      "title": "Contributing",
      "description": "Guides on how to contribute to monetr, make changes to the application's code."
    }
  }, {
    name: "index",
    route: "/documentation",
    frontMatter: {
      "title": "Documentation",
      "description": "Guides on how to use, self-host, or develop against monetr."
    }
  }, {
    name: "install",
    route: "/documentation/install",
    children: [{
      name: "docker",
      route: "/documentation/install/docker",
      frontMatter: {
        "title": "Self-Host via Docker",
        "description": "Self-host monetr via Docker containers"
      }
    }]
  }, {
    name: "install",
    route: "/documentation/install",
    frontMatter: {
      "title": "Self-Host Installation",
      "description": "Options on how to run monetr yourself for free."
    }
  }, {
    name: "use",
    route: "/documentation/use",
    children: [{
      data: documentation_use_meta
    }, {
      name: "expense",
      route: "/documentation/use/expense",
      frontMatter: {
        "title": "Expenses",
        "description": "Keep track of your regular or planned spending easily using expenses."
      }
    }, {
      name: "free_to_use",
      route: "/documentation/use/free_to_use",
      frontMatter: {
        "sidebarTitle": "Free to Use"
      }
    }, {
      name: "funding_schedule",
      route: "/documentation/use/funding_schedule",
      frontMatter: {
        "title": "Funding Schedules",
        "description": "Contribute to your budgets on a regular basis, like every time you get paid."
      }
    }, {
      name: "goal",
      route: "/documentation/use/goal",
      frontMatter: {
        "sidebarTitle": "Goal"
      }
    }, {
      name: "security",
      route: "/documentation/use/security",
      children: [{
        name: "user_password",
        route: "/documentation/use/security/user_password",
        frontMatter: {
          "sidebarTitle": "User Password"
        }
      }]
    }, {
      name: "starting_fresh",
      route: "/documentation/use/starting_fresh",
      frontMatter: {
        "sidebarTitle": "Starting Fresh"
      }
    }]
  }, {
    name: "use",
    route: "/documentation/use",
    frontMatter: {
      "title": "Using monetr",
      "description": "How to use and get the most out of monetr"
    }
  }]
}, {
  name: "index",
  route: "/",
  frontMatter: {
    "title": "monetr",
    "description": "Always know what you can spend. Put a bit of money aside every time you get paid. Always be sure you'll have enough to cover your bills, and know what you have left-over to save or spend on whatever you'd like."
  }
}, {
  name: "policy",
  route: "/policy",
  children: [{
    data: policy_meta
  }, {
    name: "privacy",
    route: "/policy/privacy",
    frontMatter: {
      "sidebarTitle": "Privacy"
    }
  }, {
    name: "terms",
    route: "/policy/terms",
    frontMatter: {
      "sidebarTitle": "Terms"
    }
  }]
}, {
  name: "pricing",
  route: "/pricing",
  frontMatter: {
    "title": "Pricing"
  }
}];
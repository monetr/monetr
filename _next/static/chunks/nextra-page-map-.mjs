import meta from "../../../src/pages/_meta.ts";
import documentation_meta from "../../../src/pages/documentation/_meta.ts";
import documentation_development_meta from "../../../src/pages/documentation/development/_meta.ts";
import documentation_install_meta from "../../../src/pages/documentation/install/_meta.ts";
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
      name: "recaptcha",
      route: "/documentation/configure/recaptcha",
      frontMatter: {
        "title": "ReCAPTCHA"
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
      "description": "Learn how to configure your self-hosted monetr installation using the comprehensive YAML configuration file. Explore detailed guides for customizing server, database, email, security, and more."
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
      "description": "Explore the monetr documentation to learn how to get started, host the application, and contribute to development. Find all the resources you need to effectively manage your finances with monetr."
    }
  }, {
    name: "install",
    route: "/documentation/install",
    children: [{
      data: documentation_install_meta
    }, {
      name: "docker",
      route: "/documentation/install/docker",
      frontMatter: {
        "title": "Self-Host with Docker Compose",
        "description": "Learn how to self-host monetr using Docker Compose. Follow step-by-step instructions to set up monetr, manage updates, and troubleshoot common issues for a seamless self-hosting experience."
      }
    }]
  }, {
    name: "install",
    route: "/documentation/install",
    frontMatter: {
      "title": "Self-Hosted Installation",
      "description": "Learn how to self-host monetr for free using Docker or Podman. Explore the benefits of self-hosting and get an overview of installation requirements and options."
    }
  }, {
    name: "use",
    route: "/documentation/use",
    children: [{
      data: documentation_use_meta
    }, {
      name: "billing",
      route: "/documentation/use/billing",
      frontMatter: {
        "title": "Billing",
        "description": "Learn about monetr's billing process, including the 30-day free trial, subscription details, and how to manage or cancel your subscription. Stay informed about payments, access, and managing your account."
      }
    }, {
      name: "expense",
      route: "/documentation/use/expense",
      frontMatter: {
        "title": "Expenses",
        "description": "Learn how to manage recurring expenses like rent, subscriptions, and credit card payments with monetr. This guide covers creating, tracking, and optimizing expenses to ensure consistent budgeting and predictable Free-To-Use funds."
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
        "description": "Discover how to set up and optimize funding schedules in monetr to manage your budgets effectively. Learn how funding schedules allocate funds for recurring expenses, ensure consistent budgeting, and maintain predictable Free-To-Use funds with every paycheck."
      }
    }, {
      name: "getting_started",
      route: "/documentation/use/getting_started",
      frontMatter: {
        "title": "Getting Started",
        "description": "Learn how to set up monetr for effective financial management. This guide walks you through connecting your bank account via Plaid or setting up a manual budget, configuring budgets, and creating a funding schedule to take control of your finances."
      }
    }, {
      name: "goal",
      route: "/documentation/use/goal",
      frontMatter: {
        "title": "Goals",
        "description": "Learn how to use monetr's Goals feature to save for one-time financial targets like vacations, loans, or down payments. Understand how Goals track contributions and spending, helping you plan effectively and meet your financial objectives without overfunding."
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
      name: "transactions",
      route: "/documentation/use/transactions",
      frontMatter: {
        "sidebarTitle": "Transactions"
      }
    }]
  }, {
    name: "use",
    route: "/documentation/use",
    frontMatter: {
      "title": "Using monetr",
      "description": "Discover how to use monetr to effectively manage your finances. Explore guides on setting up your account, managing recurring expenses, creating funding schedules, planning savings goals, and customizing your budget."
    }
  }]
}, {
  name: "index",
  route: "/",
  frontMatter: {
    "title": "monetr: Take Control of Your Finances",
    "description": "Take control of your finances, paycheck by paycheck, with monetr. Put aside what you need, spend what you want, and confidently manage your money with ease. Always know you’ll have enough for your bills and what’s left to save or spend."
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
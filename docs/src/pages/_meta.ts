export default {
  'index': {
    'type': 'page',
    'title': 'monetr',
    'display': 'hidden',
    'theme': {
      'layout': 'raw',
    },
  },
  'pricing': {
    'type': 'page',
    'title': 'Pricing',
    'theme': {
      'layout': 'raw',
    },
  },
  'blog': {
    'type': 'page',
    'title': 'Blog',
    'theme': {
      'layout': 'raw',
    },
  },
  'documentation': {
    'type': 'page',
    'title': 'Documentation',
  },
  '---': {
    type: 'separator',
  },
  'sign-up': {
    // This will show up in the dropdown menu when the screen is very small and the actual sign up button is hidden.
    title: 'Sign Up',
    href: 'https://my.monetr.app/register',
    newWindow: true,
  },
  // These are linked elsewhere and should not show up in the dropdown menu or top nav.
  'contact': {
    'type': 'page',
    'title': 'Contact',
    'display': 'hidden',
  },
  'policy': {
    'type': 'page',
    'title': 'Policies',
    'display': 'hidden',
  },
};

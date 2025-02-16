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
  'roadmap': {
    'type': 'page',
    'title': 'Roadmap',
    theme: {
      toc: false,
      sidebar: false,
      pagination: true,
      typesetting: 'article',
      layout: 'default',
    },
  },
  'documentation': {
    'type': 'page',
    'title': 'Documentation',
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

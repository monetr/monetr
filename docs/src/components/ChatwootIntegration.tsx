import { Fragment } from 'react';

export default function ChatwootIntegration(): JSX.Element {
  if (process.env.CHATWOOT_URL && process.env.CHATWOOT_TOKEN) {
    return (
      <script
        dangerouslySetInnerHTML={{
          __html: `
        (function(d,t) {
          var BASE_URL="${process.env.CHATWOOT_URL}";
          var g=d.createElement(t),s=d.getElementsByTagName(t)[0];
          g.src=BASE_URL+"/packs/js/sdk.js";
          g.defer = true;
          g.async = true;
          s.parentNode.insertBefore(g,s);
          g.onload=function(){
            window.chatwootSDK.run({
              websiteToken: '${process.env.CHATWOOT_TOKEN}',
              baseUrl: BASE_URL
            })
          }
        })(document,"script");
        `,
        }}
      />
    );
  }

  return <Fragment />;
}

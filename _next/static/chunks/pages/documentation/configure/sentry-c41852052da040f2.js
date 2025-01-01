(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[1847],{8754:(e,s,i)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/sentry",function(){return i(468)}])},468:(e,s,i)=>{"use strict";i.r(s),i.d(s,{default:()=>l,useTOC:()=>h});var n=i(2540),r=i(7933),t=i(7170),d=i(8795);function h(e){return[]}let l=(0,r.e)(function(e){let s={a:"a",code:"code",h1:"h1",p:"p",pre:"pre",span:"span",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,d.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(s.h1,{children:"Sentry Configuration"}),"\n",(0,n.jsxs)(s.p,{children:["monetr uses ",(0,n.jsx)(s.a,{href:"https://github.com/getsentry/sentry",children:"Sentry"})," for error reporting and performance monitoring. But it\nsupports using a separate DSN for the frontend versus the backend. Or you can use the same DSN for both."]}),"\n",(0,n.jsx)(s.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,n.jsxs)(s.code,{children:[(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"sentry"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  enabled"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:"<true|false>"})]}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  dsn"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"           # DSN for the backend, API and job runner."})]}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  externalDsn"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"   # DSN that is used by the frontend at runtime."})]}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  sampleRate"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"1.0"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:"      # Sample rate for errors"})]}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  traceSampleRate"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"1.0"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#6A737D","--shiki-dark":"#6A737D"},children:" # Sample rate for performance traces"})]}),"\n",(0,n.jsxs)(s.span,{children:[(0,n.jsx)(s.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  securityHeaderEndpoint"}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,n.jsx)(s.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'})]})]})}),"\n",(0,n.jsxs)(s.table,{children:[(0,n.jsx)(s.thead,{children:(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Name"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Type"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Default"})}),(0,n.jsx)(s.th,{children:(0,n.jsx)(s.strong,{children:"Description"})})]})}),(0,n.jsxs)(s.tbody,{children:[(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"enabled"})}),(0,n.jsx)(s.td,{children:"Boolean"}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"false"})}),(0,n.jsx)(s.td,{children:"Enable the Sentry integration with monetr, allowing you to gather debug information about your instance if you run into any issue."})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"dsn"})}),(0,n.jsx)(s.td,{children:"String"}),(0,n.jsx)(s.td,{}),(0,n.jsx)(s.td,{children:"Specify the DSN that the backend will use for its errors and performance traces. This DSN is never exposed publicly."})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"externalDsn"})}),(0,n.jsx)(s.td,{children:"String"}),(0,n.jsx)(s.td,{}),(0,n.jsxs)(s.td,{children:["Specify the DSN that the frontend portion of monetr will use. ",(0,n.jsx)(s.strong,{children:"Note"}),": This DSN is publicly visible even without authentication as it is loaded into the ",(0,n.jsx)(s.code,{children:"index.html"})," content served for the frontend with each request."]})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"sampleRate"})}),(0,n.jsx)(s.td,{children:"Float"}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"1.0"})}),(0,n.jsxs)(s.td,{children:["Specify a sample rate for errors, ",(0,n.jsx)(s.code,{children:"1.0"})," would be sampling every error where ",(0,n.jsx)(s.code,{children:"0.0"})," would be sampling none of them."]})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"traceSampleRate"})}),(0,n.jsx)(s.td,{children:"Float"}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"1.0"})}),(0,n.jsxs)(s.td,{children:["Specify a sample rate for performance traces, ",(0,n.jsx)(s.code,{children:"1.0"})," would be sampling every transaction span where ",(0,n.jsx)(s.code,{children:"0.0"})," would be sampling none of them."]})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"securityHeaderEndpoint"})}),(0,n.jsx)(s.td,{children:"String"}),(0,n.jsx)(s.td,{}),(0,n.jsx)(s.td,{children:"Specify a sentry URL to report CSP violations to. monetr enforces a strict CSP policy by default."})]})]})]}),"\n",(0,n.jsx)(s.p,{children:"The following environment variables map to the following configuration file fields. Each field is documented below."}),"\n",(0,n.jsxs)(s.table,{children:[(0,n.jsx)(s.thead,{children:(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.th,{children:"Variable"}),(0,n.jsx)(s.th,{children:"Config File Field"})]})}),(0,n.jsxs)(s.tbody,{children:[(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"MONETR_SENTRY_ENABLED"})}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"sentry.enabled"})})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"MONETR_SENTRY_DSN"})}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"sentry.dsn"})})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"MONETR_SENTRY_EXTERNAL_DSN"})}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"sentry.externalDsn"})})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"MONETR_SENTRY_SAMPLE_RATE"})}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"sentry.sampleRate"})})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"MONETR_SENTRY_TRACE_SAMPLE_RATE"})}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"sentry.traceSampleRate"})})]}),(0,n.jsxs)(s.tr,{children:[(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"MONETR_SENTRY_CSP_ENDPOINT"})}),(0,n.jsx)(s.td,{children:(0,n.jsx)(s.code,{children:"sentry.securityHeaderEndpoint"})})]})]})]})]})},"/documentation/configure/sentry",{filePath:"src/pages/documentation/configure/sentry.mdx",timestamp:173561954e4,pageMap:t.O,frontMatter:{},title:"Sentry Configuration"},"undefined"==typeof RemoteContent?h:RemoteContent.useTOC)}},e=>{var s=s=>e(e.s=s);e.O(0,[7933,7170,636,6593,8792],()=>s(8754)),_N_E=e.O()}]);
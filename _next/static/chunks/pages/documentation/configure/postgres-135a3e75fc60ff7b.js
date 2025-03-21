(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[4583],{3608:(e,s,i)=>{"use strict";i.r(s),i.d(s,{default:()=>a,useTOC:()=>h});var r=i(6514),t=i(7017),d=i(235),n=i(9493),l=i(4299);function h(e){return[{value:"Database Migrations",id:"database-migrations",depth:2}]}let a=(0,t.e)(function(e){let{toc:s=h(e)}=e,i={code:"code",h1:"h1",h2:"h2",p:"p",pre:"pre",span:"span",strong:"strong",table:"table",tbody:"tbody",td:"td",th:"th",thead:"thead",tr:"tr",...(0,n.R)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(i.h1,{children:"PostgreSQL Configuration"}),"\n",(0,r.jsxs)(i.p,{children:["monetr’s primary database is PostgreSQL and is required in order for monetr to run. monetr also uses PostgreSQL as a\nbasic pub-sub system via ",(0,r.jsx)(i.code,{children:"LISTEN"})," and ",(0,r.jsx)(i.code,{children:"NOTIFY"})," commands."]}),"\n",(0,r.jsx)(i.p,{children:"Officially monetr supports PostgreSQL version 16 and higher."}),"\n",(0,r.jsx)(i.pre,{tabIndex:"0","data-language":"yaml","data-word-wrap":"","data-filename":"config.yaml",children:(0,r.jsxs)(i.code,{children:[(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"postgresql"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:":"})]}),"\n",(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  address"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"localhost"'})]}),"\n",(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  port"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"5432"})]}),"\n",(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  username"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"monetr"'})]}),"\n",(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  password"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"..."'})]}),"\n",(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  database"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"monetr"'})]}),"\n",(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  insecureSkipVerify"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#005CC5","--shiki-dark":"#79B8FF"},children:"false"})]}),"\n",(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  caCertificatePath"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"</tls/ca.cert>"'})]}),"\n",(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  keyPath"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"</tls/tls.key>"'})]}),"\n",(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#22863A","--shiki-dark":"#85E89D"},children:"  certificatePath"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#24292E","--shiki-dark":"#E1E4E8"},children:": "}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:'"</tls/tls.cert>"'})]})]})}),"\n",(0,r.jsxs)(i.table,{children:[(0,r.jsx)(i.thead,{children:(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.th,{children:(0,r.jsx)(i.strong,{children:"Name"})}),(0,r.jsx)(i.th,{children:(0,r.jsx)(i.strong,{children:"Type"})}),(0,r.jsx)(i.th,{children:(0,r.jsx)(i.strong,{children:"Default"})}),(0,r.jsx)(i.th,{children:(0,r.jsx)(i.strong,{children:"Description"})})]})}),(0,r.jsxs)(i.tbody,{children:[(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"address"})}),(0,r.jsx)(i.td,{children:"String"}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"localhost"})}),(0,r.jsx)(i.td,{children:"The IP, or DNS resolvable address of your PostgreSQL database server."})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"port"})}),(0,r.jsx)(i.td,{children:"Number"}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"5432"})}),(0,r.jsx)(i.td,{children:"Port that the PostgreSQL server can be reached at."})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"username"})}),(0,r.jsx)(i.td,{children:"String"}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgres"})}),(0,r.jsx)(i.td,{children:"Username that monetr should use to authenticate the PostgreSQL server."})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"password"})}),(0,r.jsx)(i.td,{children:"String"}),(0,r.jsx)(i.td,{}),(0,r.jsx)(i.td,{children:"Password that monetr should use to authenticate the PostgreSQL server."})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"database"})}),(0,r.jsx)(i.td,{children:"String"}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgres"})}),(0,r.jsx)(i.td,{children:"Database that monetr should use, monetr may attempt to run migrations on startup. The user monetr is using should have permissions to create tables and extensions."})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"insecureSkipVerify"})}),(0,r.jsx)(i.td,{children:"Boolean"}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"false"})}),(0,r.jsx)(i.td,{children:"If you are using TLS with PostgreSQL but are not distributing a certificate authority file, then you may need to skip TLS verification."})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"caCertificatePath"})}),(0,r.jsx)(i.td,{children:"String"}),(0,r.jsx)(i.td,{}),(0,r.jsx)(i.td,{children:"Path to the certificate authority certificate file. If you are verifying your TLS connection then this is required or the server certificate must be among the hosts certificate authorities already."})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"keyPath"})}),(0,r.jsx)(i.td,{children:"String"}),(0,r.jsx)(i.td,{}),(0,r.jsx)(i.td,{children:"Path to the client TLS key that monetr should use to connect to the PostgreSQL server."})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"certificatePath"})}),(0,r.jsx)(i.td,{children:"String"}),(0,r.jsx)(i.td,{}),(0,r.jsx)(i.td,{children:"Path to the client TLS certificate that monetr should use to connect to the PostgreSQL server."})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"migrate"})}),(0,r.jsx)(i.td,{children:"Boolean"}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"true"})}),(0,r.jsx)(i.td,{children:"Automatically apply database migrations on startup."})]})]})]}),"\n",(0,r.jsx)(l.P,{type:"info",children:(0,r.jsx)(i.p,{children:"monetr does watch for certificate changes on the filesystem to facilitate certificate rotation without needing to\nrestart the server. However this functionality does not always work and should not be relied on at this time."})}),"\n",(0,r.jsx)(i.p,{children:"The following environment variables map to the following configuration file fields. Each field is documented below."}),"\n",(0,r.jsxs)(i.table,{children:[(0,r.jsx)(i.thead,{children:(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.th,{children:"Variable"}),(0,r.jsx)(i.th,{children:"Config File Field"})]})}),(0,r.jsxs)(i.tbody,{children:[(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"MONETR_PG_ADDRESS"})}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgresql.address"})})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"MONETR_PG_PORT"})}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgresql.port"})})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"MONETR_PG_USERNAME"})}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgresql.username"})})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"MONETR_PG_PASSWORD"})}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgresql.password"})})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"MONETR_PG_DATABASE"})}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgresql.database"})})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"MONETR_PG_INSECURE_SKIP_VERIFY"})}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgresql.insecureSkipVerify"})})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"MONETR_PG_CA_PATH"})}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgresql.caCertificatePath"})})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"MONETR_PG_KEY_PATH"})}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgresql.keyPath"})})]}),(0,r.jsxs)(i.tr,{children:[(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"MONETR_PG_CERT_PATH"})}),(0,r.jsx)(i.td,{children:(0,r.jsx)(i.code,{children:"postgresql.certificatePath"})})]})]})]}),"\n",(0,r.jsx)(i.h2,{id:s[0].id,children:s[0].value}),"\n",(0,r.jsxs)(i.p,{children:["The provided Docker Compose file will automatically run database migrations on startup as needed. However if you want to\nrun the migrations manually you can remove the ",(0,r.jsx)(i.code,{children:"--migrate"})," flag from the serve command in the compose file."]}),"\n",(0,r.jsx)(i.p,{children:"To run database migrations manually run the following command:"}),"\n",(0,r.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Run Database Migrations",children:(0,r.jsx)(i.code,{children:(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"monetr"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" database"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" migrate"})]})})}),"\n",(0,r.jsx)(i.p,{children:"To see the current database schema version run the following command:"}),"\n",(0,r.jsx)(i.pre,{tabIndex:"0","data-language":"shell","data-word-wrap":"","data-filename":"Database Schema Version",children:(0,r.jsx)(i.code,{children:(0,r.jsxs)(i.span,{children:[(0,r.jsx)(i.span,{style:{"--shiki-light":"#6F42C1","--shiki-dark":"#B392F0"},children:"monetr"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" database"}),(0,r.jsx)(i.span,{style:{"--shiki-light":"#032F62","--shiki-dark":"#9ECBFF"},children:" version"})]})})})]})},"/documentation/configure/postgres",{filePath:"src/pages/documentation/configure/postgres.mdx",timestamp:173980644e4,pageMap:d.O,frontMatter:{title:"PostgreSQL"},title:"PostgreSQL"},"undefined"==typeof RemoteContent?h:RemoteContent.useTOC)},4299:(e,s,i)=>{"use strict";i.d(s,{P:()=>h});var r=i(6514),t=i(3367),d=i(3413);let n={default:"\uD83D\uDCA1",error:"\uD83D\uDEAB",info:(0,r.jsx)(d.KS,{className:"_mt-1"}),warning:"⚠️"},l={default:(0,t.A)("_border-orange-100 _bg-orange-50 _text-orange-800 dark:_border-orange-400/30 dark:_bg-orange-400/20 dark:_text-orange-300"),error:(0,t.A)("_border-red-200 _bg-red-100 _text-red-900 dark:_border-red-200/30 dark:_bg-red-900/30 dark:_text-red-200"),info:(0,t.A)("_border-blue-200 _bg-blue-100 _text-blue-900 dark:_border-blue-200/30 dark:_bg-blue-900/30 dark:_text-blue-200"),warning:(0,t.A)("_border-yellow-100 _bg-yellow-50 _text-yellow-900 dark:_border-yellow-200/30 dark:_bg-yellow-700/30 dark:_text-yellow-200")};function h({children:e,type:s="default",emoji:i=n[s]}){return(0,r.jsxs)("div",{className:(0,t.A)("nextra-callout _overflow-x-auto _mt-6 _flex _rounded-lg _border _py-2 ltr:_pr-4 rtl:_pl-4","contrast-more:_border-current contrast-more:dark:_border-current",l[s]),children:[(0,r.jsx)("div",{className:"_select-none _text-xl ltr:_pl-3 ltr:_pr-2 rtl:_pr-3 rtl:_pl-2",style:{fontFamily:'"Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"'},children:i}),(0,r.jsx)("div",{className:"_w-full _min-w-0 _leading-7",children:e})]})}},9920:(e,s,i)=>{(window.__NEXT_P=window.__NEXT_P||[]).push(["/documentation/configure/postgres",function(){return i(3608)}])}},e=>{var s=s=>e(e.s=s);e.O(0,[7017,235,636,6593,8792],()=>s(9920)),_N_E=e.O()}]);
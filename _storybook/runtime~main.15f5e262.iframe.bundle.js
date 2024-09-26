!function(){var e,r,n,o,t,s,d,i,c,a,u,l={},f={};function p(e){var r=f[e];if(void 0!==r){if(void 0!==r.error)throw r.error;return r.exports}var n=f[e]={id:e,exports:{}};try{var o={id:e,module:n,factory:l[e],require:p};p.i.forEach(function(e){e(o)}),n=o.module,!o.factory&&console.error("undefined factory",e),o.factory.call(n.exports,n,n.exports,o.require)}catch(e){throw n.error=e,e}return n.exports}p.m=l,p.c=f,p.i=[],p.f={},p.e=function(e){return Promise.all(Object.keys(p.f).reduce(function(r,n){return p.f[n](e,r),r},[]))},!function(){var e,r,n,o={},t=p.c,s=[],d=[],i="idle",c=0,a=[];p.hmrD=o,p.i.push(function(a){var h=a.module,m=function(r,n){var o=t[n];if(!o)return r;var d=function(d){if(o.hot.active){if(t[d]){var i=t[d].parents;-1===i.indexOf(n)&&i.push(n)}else s=[n],e=d;-1===o.children.indexOf(d)&&o.children.push(d)}else console.warn("[HMR] unexpected require("+d+") from disposed module "+n),s=[];return r(d)},a=function(e){return{configurable:!0,enumerable:!0,get:function(){return r[e]},set:function(n){r[e]=n}}};for(var f in r)Object.prototype.hasOwnProperty.call(r,f)&&"e"!==f&&Object.defineProperty(d,f,a(f));return d.e=function(e){return function(e){switch(i){case"ready":u("prepare");case"prepare":return c++,e.then(l,l),e;default:return e}}(r.e(e))},d}(a.require,a.id);h.hot=function(t,c){var a=e!==t,l={_acceptedDependencies:{},_acceptedErrorHandlers:{},_declinedDependencies:{},_selfAccepted:!1,_selfDeclined:!1,_selfInvalidated:!1,_disposeHandlers:[],_main:a,_requireSelf:function(){s=c.parents.slice(),e=a?void 0:t,p(t)},active:!0,accept:function(e,r,n){if(void 0===e)l._selfAccepted=!0;else if("function"==typeof e)l._selfAccepted=e;else if("object"==typeof e&&null!==e)for(var o=0;o<e.length;o++)l._acceptedDependencies[e[o]]=r||function(){},l._acceptedErrorHandlers[e[o]]=n;else l._acceptedDependencies[e]=r||function(){},l._acceptedErrorHandlers[e]=n},decline:function(e){if(void 0===e)l._selfDeclined=!0;else if("object"==typeof e&&null!==e)for(var r=0;r<e.length;r++)l._declinedDependencies[e[r]]=!0;else l._declinedDependencies[e]=!0},dispose:function(e){l._disposeHandlers.push(e)},addDisposeHandler:function(e){l._disposeHandlers.push(e)},removeDisposeHandler:function(e){var r=l._disposeHandlers.indexOf(e);r>0&&l._disposeHandlers.splice(r,1)},invalidate:function(){switch(this._selfInvalidated=!0,i){case"idle":r=[],Object.keys(p.hmrI).forEach(function(e){p.hmrI[e](t,r)}),u("ready");break;case"ready":Object.keys(p.hmrI).forEach(function(e){p.hmrI[e](t,r)});break;case"prepare":case"check":case"dispose":case"apply":(n=n||[]).push(t)}},check:f,apply:_,status:function(e){if(!e)return i;d.push(e)},addStatusHandler:function(e){d.push(e)},removeStatusHandler:function(e){var r=d.indexOf(e);r>=0&&d.splice(r,1)},data:o[t]};return e=void 0,l}(a.id,h),h.parents=s,h.children=[],s=[],a.require=m}),p.hmrC={},p.hmrI={};function u(e){i=e;for(var r=[],n=0;n<d.length;n++)r[n]=d[n].call(null,e);return Promise.all(r)}function l(){0==--c&&u("ready").then(function(){if(0===c){var e=a;a=[];for(var r=0;r<e.length;r++)e[r]()}})}function f(e){if("idle"!==i)throw Error("check() is only allowed in idle status");return u("check").then(p.hmrM).then(function(n){return n?u("prepare").then(function(){var o=[];return r=[],Promise.all(Object.keys(p.hmrC).reduce(function(e,t){return p.hmrC[t](n.c,n.r,n.m,e,r,o),e},[])).then(function(){var r;return r=function(){return e?h(e):u("ready").then(function(){return o})},0===c?r():new Promise(function(e){a.push(function(){e(r())})})})}):u(m()?"ready":"idle").then(function(){return null})})}function _(e){return"ready"!==i?Promise.resolve().then(function(){throw Error("apply() is only allowed in ready status (state: "+i+")")}):h(e)}function h(e){e=e||{},m();var o,t=r.map(function(r){return r(e)});r=void 0;var s=t.map(function(e){return e.error}).filter(Boolean);if(s.length>0)return u("abort").then(function(){throw s[0]});var d=u("dispose");t.forEach(function(e){e.dispose&&e.dispose()});var i=u("apply"),c=function(e){!o&&(o=e)},a=[];return t.forEach(function(e){if(e.apply){var r=e.apply(c);if(r)for(var n=0;n<r.length;n++)a.push(r[n])}}),Promise.all([d,i]).then(function(){return o?u("fail").then(function(){throw o}):n?h(e).then(function(e){return a.forEach(function(r){0>e.indexOf(r)&&e.push(r)}),e}):u("idle").then(function(){return a})})}function m(){if(n)return!r&&(r=[]),Object.keys(p.hmrI).forEach(function(e){n.forEach(function(n){p.hmrI[e](n,r)})}),n=void 0,!0}}(),!function(){function e(r){if("function"!=typeof WeakMap)return null;var n=new WeakMap,o=new WeakMap;return(e=function(e){return e?o:n})(r)}p.ir=function(r,n){if(!n&&r&&r.__esModule)return r;if(null===r||"object"!=typeof r&&"function"!=typeof r)return{default:r};var o=e(n);if(o&&o.has(r))return o.get(r);var t={},s=Object.defineProperty&&Object.getOwnPropertyDescriptor;for(var d in r)if("default"!==d&&Object.prototype.hasOwnProperty.call(r,d)){var i=s?Object.getOwnPropertyDescriptor(r,d):null;i&&(i.get||i.set)?Object.defineProperty(t,d,i):t[d]=r[d]}return t.default=r,o&&o.set(r,t),t}}(),p.es=function(e,r){return Object.keys(e).forEach(function(n){"default"!==n&&!Object.prototype.hasOwnProperty.call(r,n)&&Object.defineProperty(r,n,{enumerable:!0,get:function(){return e[n]}})}),e},e=[],p.O=function(r,n,o,t){if(n){t=t||0;for(var s=e.length;s>0&&e[s-1][2]>t;s--)e[s]=e[s-1];e[s]=[n,o,t];return}for(var d=1/0,s=0;s<e.length;s++){for(var n=e[s][0],o=e[s][1],t=e[s][2],i=!0,c=0;c<n.length;c++)d>=t&&Object.keys(p.O).every(function(e){return p.O[e](n[c])})?n.splice(c--,1):(i=!1,t<d&&(d=t));if(i){e.splice(s--,1);var a=o();void 0!==a&&(r=a)}}return r},r={"../interface/src/pages/app.stories.tsx":["interface_src_pages_app_stories_tsx","interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc","splitting~interface_src_pages_app_stories_tsx"],"../node_modules/@mdx-js/react/index.js":["1"],"../node_modules/@storybook/addon-docs/dist/DocsRenderer-EYKKDMVH.mjs":["interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc","node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs","splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs"],"../node_modules/@storybook/blocks/dist/Color-3YIJY6X7.mjs":["5"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/GlobalScrollAreaStyles-XIHNDKUY.mjs":["7"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/OverlayScrollbars-VAV6LJAB.mjs":["6"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/WithTooltip-3BDV6MYO.mjs":["0"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/formatter-UT3ZCDIS.mjs":["3"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/index.mjs":[],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/syntaxhighlighter-QTQ2UBB4.mjs":["4"]},p.el=function(e){var n=r[e];return void 0===n?Promise.resolve():n.length>1?Promise.all(n.map(p.e)):p.e(n[0])},p.g=function(){if("object"==typeof globalThis)return globalThis;try{return this||Function("return this")()}catch(e){if("object"==typeof window)return window}}(),p.h=function(){return"db303c57d99bdf4f"},p.hmrF=function(){return"runtime~main."+p.h()+".hot-update.json"},p.hu=function(e){return""+e+"."+p.h()+".hot-update.js"},p.k=function(e){return({0:"0.2d068005.iframe.bundle.css",1:"1.2d068005.iframe.bundle.css",2:"2.6be7d201.iframe.bundle.css",3:"3.2d068005.iframe.bundle.css",4:"4.2d068005.iframe.bundle.css",5:"5.2d068005.iframe.bundle.css",6:"6.2d068005.iframe.bundle.css",7:"7.2d068005.iframe.bundle.css",interface_src_pages_app_stories_tsx:"interface_src_pages_app_stories_tsx.5c0483b0.iframe.bundle.css","interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc":"interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc.2d068005.iframe.bundle.css",main:"main.75341c2b.iframe.bundle.css","node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.2d068005.iframe.bundle.css","runtime~main":"runtime~main.2d068005.iframe.bundle.css","splitting~interface_src_pages_app_stories_tsx":"splitting~interface_src_pages_app_stories_tsx.2d068005.iframe.bundle.css","splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.2d068005.iframe.bundle.css"})[e]},n={},p.l=function(e,r,o,t){if(void 0!==o){for(var s,d,i=document.getElementsByTagName("script"),c=0;c<i.length;c++){var a=i[c];if(a.getAttribute("src")==e){s=a;break}}}!s&&(d=!0,(s=document.createElement("script")).charset="utf-8",s.timeout=120,s.src=e),n[e]=[r];var u=function(r,o){s.onerror=s.onload=null,clearTimeout(l);var t=n[e];if(delete n[e],s.parentNode&&s.parentNode.removeChild(s),t&&t.forEach(function(e){return e(o)}),r)return r(o)},l=setTimeout(u.bind(null,void 0,{type:"timeout",target:s}),12e4);s.onerror=u.bind(null,s.onerror),s.onload=u.bind(null,s.onload),d&&document.head.appendChild(s)},p.o=function(e,r){return Object.prototype.hasOwnProperty.call(e,r)},p.p="",p.u=function(e){return({0:"0.cb55ad32.iframe.bundle.js",1:"1.c887fbe5.iframe.bundle.js",3:"3.e348b969.iframe.bundle.js",4:"4.6983a53a.iframe.bundle.js",5:"5.732ef84a.iframe.bundle.js",6:"6.fef3f55d.iframe.bundle.js",7:"7.cfb8087a.iframe.bundle.js",interface_src_pages_app_stories_tsx:"interface_src_pages_app_stories_tsx.1a1edac9.iframe.bundle.js","interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc":"interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc.70d00da8.iframe.bundle.js","node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.53c4e13c.iframe.bundle.js","splitting~interface_src_pages_app_stories_tsx":"splitting~interface_src_pages_app_stories_tsx.eaf5e17c.iframe.bundle.js","splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.7bf02a7b.iframe.bundle.js"})[e]},!function(){var e,r,n,o,t,s={"runtime~main":0};p.f.j=function(e,r){var n=p.o(s,e)?s[e]:void 0;if(0!==n){if(n)r.push(n[2]);else if(e){var o=new Promise(function(r,o){n=s[e]=[r,o]});r.push(n[2]=o);var t=p.p+p.u(e),d=Error();p.l(t,function(r){if(p.o(s,e)&&(0!==(n=s[e])&&(s[e]=void 0),n)){var o=r&&("load"===r.type?"missing":r.type),t=r&&r.target&&r.target.src;d.message="Loading chunk "+e+" failed.\n("+o+": "+t+")",d.name="ChunkLoadError",d.type=o,d.request=t,n[1](d)}},"chunk-"+e,e)}else s[e]=0}};var d={};function i(r,n){return e=n,new Promise(function(e,n){var o=p.p+p.hu(r);d[r]=e;var t=Error();p.l(o,function(e){if(d[r]){d[r]=void 0;var o=e&&("load"===e.type?"missing":e.type),s=e&&e.target&&e.target.src;t.message="Loading hot update chunk "+r+" failed.\n("+o+": "+s+")",t.name="ChunkLoadError",t.type=o,t.request=s,n(t)}})})}function c(e){p.f&&delete p.f.jsonpHmr,r=void 0;function d(e,r){for(var n=0;n<r.length;n++){var o=r[n];-1===e.indexOf(o)&&e.push(o)}}var i,c={},a=[],u={},l=function(e){console.warn("[HMR] unexpected require("+e.id+") to disposed module")};for(var f in n)if(p.o(n,f)){var _,h=n[f];_=h?function(e){for(var r=[e],n={},o=r.map(function(e){return{chain:[e],id:e}});o.length>0;){var t=o.pop(),s=t.id,i=t.chain,c=p.c[s];if(!!c&&(!c.hot._selfAccepted||!!c.hot._selfInvalidated)){if(c.hot._selfDeclined)return{type:"self-declined",chain:i,moduleId:s};if(c.hot._main)return{type:"unaccepted",chain:i,moduleId:s};for(var a=0;a<c.parents.length;a++){var u=c.parents[a],l=p.c[u];if(!l)continue;if(l.hot._declinedDependencies[s])return{type:"declined",chain:i.concat([u]),moduleId:s,parentId:u};if(-1===r.indexOf(u)){if(l.hot._acceptedDependencies[s]){!n[u]&&(n[u]=[]),d(n[u],[s]);continue}delete n[u],r.push(u),o.push({chain:i.concat([u]),id:u})}}}}return{type:"accepted",moduleId:e,outdatedModules:r,outdatedDependencies:n}}(f):{type:"disposed",moduleId:f};var m=!1,b=!1,v=!1,y="";switch(_.chain&&(y="\nUpdate propagation: "+_.chain.join(" -> ")),_.type){case"self-declined":e.onDeclined&&e.onDeclined(_),!e.ignoreDeclined&&(m=Error("Aborted because of self decline: "+_.moduleId+y));break;case"declined":e.onDeclined&&e.onDeclined(_),!e.ignoreDeclined&&(m=Error("Aborted because of declined dependency: "+_.moduleId+" in "+_.parentId+y));break;case"unaccepted":e.onUnaccepted&&e.onUnaccepted(_),!e.ignoreUnaccepted&&(m=Error("Aborted because "+f+" is not accepted"+y));break;case"accepted":e.onAccepted&&e.onAccepted(_),b=!0;break;case"disposed":e.onDisposed&&e.onDisposed(_),v=!0;break;default:throw Error("Unexception type "+_.type)}if(m)return{error:m};if(b)for(f in u[f]=h,d(a,_.outdatedModules),_.outdatedDependencies)p.o(_.outdatedDependencies,f)&&(!c[f]&&(c[f]=[]),d(c[f],_.outdatedDependencies[f]));v&&(d(a,[_.moduleId]),u[f]=l)}n=void 0;for(var g=[],k=0;k<a.length;k++){var E=a[k],D=p.c[E];D&&(D.hot._selfAccepted||D.hot._main)&&u[E]!==l&&!D.hot._selfInvalidated&&g.push({module:E,require:D.hot._requireSelf,errorHandler:D.hot._selfAccepted})}return{dispose:function(){o.forEach(function(e){delete s[e]}),o=void 0;for(var e,r,n=a.slice();n.length>0;){var t=n.pop(),d=p.c[t];if(!!d){var u={},l=d.hot._disposeHandlers;for(k=0;k<l.length;k++)l[k].call(null,u);for(p.hmrD[t]=u,d.hot.active=!1,delete p.c[t],delete c[t],k=0;k<d.children.length;k++){var f=p.c[d.children[k]];f&&(e=f.parents.indexOf(t))>=0&&f.parents.splice(e,1)}}}for(var _ in c)if(p.o(c,_)&&(d=p.c[_]))for(k=0,i=c[_];k<i.length;k++)r=i[k],(e=d.children.indexOf(r))>=0&&d.children.splice(e,1)},apply:function(r){for(var n in u)p.o(u,n)&&(p.m[n]=u[n]);for(var o=0;o<t.length;o++)t[o](p);for(var s in c)if(p.o(c,s)){var d=p.c[s];if(d){i=c[s];for(var l=[],f=[],_=[],h=0;h<i.length;h++){var m=i[h],b=d.hot._acceptedDependencies[m],v=d.hot._acceptedErrorHandlers[m];if(b){if(-1!==l.indexOf(b))continue;l.push(b),f.push(v),_.push(m)}}for(var y=0;y<l.length;y++)try{l[y].call(null,i)}catch(n){if("function"==typeof f[y])try{f[y](n,{moduleId:s,dependencyId:_[y]})}catch(o){e.onErrored&&e.onErrored({type:"accept-error-handler-errored",moduleId:s,dependencyId:_[y],error:o,originalError:n}),!e.ignoreErrored&&(r(o),r(n))}else e.onErrored&&e.onErrored({type:"accept-errored",moduleId:s,dependencyId:_[y],error:n}),!e.ignoreErrored&&r(n)}}}for(var k=0;k<g.length;k++){var E=g[k],D=E.module;try{E.require(D)}catch(n){if("function"==typeof E.errorHandler)try{E.errorHandler(n,{moduleId:D,module:p.c[D]})}catch(o){e.onErrored&&e.onErrored({type:"self-accept-error-handler-errored",moduleId:D,error:o,originalError:n}),!e.ignoreErrored&&(r(o),r(n))}else e.onErrored&&e.onErrored({type:"self-accept-errored",moduleId:D,error:n}),!e.ignoreErrored&&r(n)}}return a}}}self.hotUpdate=function(r,o,s){for(var i in o)p.o(o,i)&&(n[i]=o[i],e&&e.push(i));s&&t.push(s),d[r]&&(d[r](),d[r]=void 0)},p.hmrI.jsonp=function(e,r){!n&&(n={},t=[],o=[],r.push(c)),!p.o(n,e)&&(n[e]=p.m[e])},p.hmrC.jsonp=function(e,d,a,u,l,f){l.push(c),r={},o=d,n=a.reduce(function(e,r){return e[r]=!1,e},{}),t=[],e.forEach(function(e){p.o(s,e)&&void 0!==s[e]?(u.push(i(e,f)),r[e]=!0):r[e]=!1}),p.f&&(p.f.jsonpHmr=function(e,n){r&&p.o(r,e)&&!r[e]&&(n.push(i(e)),r[e]=!0)})},p.hmrM=function(){if("undefined"==typeof fetch)throw Error("No browser support: need fetch API");return fetch(p.p+p.hmrF()).then(function(e){if(404!==e.status){if(!e.ok)throw Error("Failed to fetch update manifest "+e.statusText);return e.json()}})},p.O.j=function(e){return 0===s[e]};var a=function(e,r){var n=r[0],o=r[1],t=r[2],d,i,c=0;if(n.some(function(e){return 0!==s[e]})){for(d in o)p.o(o,d)&&(p.m[d]=o[d]);if(t)var a=t(p)}for(e&&e(r);c<n.length;c++)i=n[c],p.o(s,i)&&s[i]&&s[i][0](),s[i]=0;return p.O(a)},u=self.webpackChunk_monetr_stories=self.webpackChunk_monetr_stories||[];u.forEach(a.bind(null,0)),u.push=a.bind(null,u.push.bind(u))}(),o={2:0,main:0},t="webpack",s="data-webpack-loading",d=function(e,r,n,o){var d,i,c="chunk-"+e;if(!o){for(var a=document.getElementsByTagName("link"),u=0;u<a.length;u++){var l=a[u],f=l.getAttribute("href")||l.href;if(f&&!f.startsWith(p.p)&&(f=p.p+(f.startsWith("/")?f.slice(1):f)),"stylesheet"==l.rel&&(f&&f.startsWith(r)||l.getAttribute("data-webpack")==t+":"+c)){d=l;break}}if(!n)return d}!d&&(i=!0,(d=document.createElement("link")).setAttribute("data-webpack",t+":"+c),d.setAttribute(s,1),d.rel="stylesheet",d.href=r);var _=function(e,r){if(d.onerror=d.onload=null,d.removeAttribute(s),clearTimeout(h),r&&"load"!=r.type&&d.parentNode.removeChild(d),n(r),e)return e(r)};if(d.getAttribute(s)){var h=setTimeout(_.bind(null,void 0,{type:"timeout",target:d}),12e4);d.onerror=_.bind(null,d.onerror),d.onload=_.bind(null,d.onload)}else _(void 0,{type:"load",target:d});return o?document.head.insertBefore(d,o):i&&document.head.appendChild(d),d},p.f.css=function(e,r){var n=p.o(o,e)?o[e]:void 0;if(0!==n){if(n)r.push(n[2]);else if(["interface_src_pages_app_stories_tsx"].indexOf(e)>-1){var t=new Promise(function(r,t){n=o[e]=[r,t]});r.push(n[2]=t);var s=p.p+p.k(e),i=Error();d(e,s,function(r){if(p.o(o,e)&&(0!==(n=o[e])&&(o[e]=void 0),n)){if("load"!==r.type){var t=r&&r.type,s=r&&r.target&&r.target.src;i.message="Loading css chunk "+e+" failed.\n("+t+": "+s+")",i.name="ChunkLoadError",i.type=t,i.request=s,n[1](i)}else n[0]()}})}else o[e]=0}},i=[],c=[],a=function(e){return{dispose:function(){},apply:function(){for(c.forEach(function(e){e[1].sheet.disabled=!1});i.length;){var e=i.pop();e.parentNode&&e.parentNode.removeChild(e)}for(;c.length;)c.pop();return[]}}},u=function(e){return Array.from(e.sheet.cssRules,function(e){return e.cssText}).join()},p.hmrC.css=function(e,r,n,o,t,s){t.push(a),e.forEach(function(e){var r=p.k(e),n=p.p+r,t=d(e,n);t&&o.push(new Promise(function(r,o){var a=d(e,n+(0>n.indexOf("?")?"?":"&")+"hmr="+Date.now(),function(d){if("load"!==d.type){var l=Error(),f=d&&d.type,p=d&&d.target&&d.target.src;l.message="Loading css hot update chunk "+e+" failed.\n("+f+": "+p+")",l.name="ChunkLoadError",l.type=f,l.request=p,o(l)}else{try{if(u(t)==u(a))return a.parentNode&&a.parentNode.removeChild(a),r()}catch(e){}s.push(n),a.sheet.disabled=!0,i.push(t),c.push([e,a]),r()}},t)}))})}}();
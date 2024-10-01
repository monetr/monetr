!function(){var e,r,n,o,t,s,i,d,a,c,u,l={},f={};function p(e){var r=f[e];if(void 0!==r){if(void 0!==r.error)throw r.error;return r.exports}var n=f[e]={id:e,exports:{}};try{var o={id:e,module:n,factory:l[e],require:p};p.i.forEach(function(e){e(o)}),n=o.module,!o.factory&&console.error("undefined factory",e),o.factory.call(n.exports,n,n.exports,o.require)}catch(e){throw n.error=e,e}return n.exports}p.m=l,p.c=f,p.i=[],p.f={},p.e=function(e){return Promise.all(Object.keys(p.f).reduce(function(r,n){return p.f[n](e,r),r},[]))},!function(){var e,r,n,o={},t=p.c,s=[],i=[],d="idle",a=0,c=[];p.hmrD=o,p.i.push(function(c){var h=c.module,m=function(r,n){var o=t[n];if(!o)return r;var i=function(i){if(o.hot.active){if(t[i]){var d=t[i].parents;-1===d.indexOf(n)&&d.push(n)}else s=[n],e=i;-1===o.children.indexOf(i)&&o.children.push(i)}else console.warn("[HMR] unexpected require("+i+") from disposed module "+n),s=[];return r(i)},c=function(e){return{configurable:!0,enumerable:!0,get:function(){return r[e]},set:function(n){r[e]=n}}};for(var f in r)Object.prototype.hasOwnProperty.call(r,f)&&"e"!==f&&Object.defineProperty(i,f,c(f));return i.e=function(e){return function(e){switch(d){case"ready":u("prepare");case"prepare":return a++,e.then(l,l),e;default:return e}}(r.e(e))},i}(c.require,c.id);h.hot=function(t,a){var c=e!==t,l={_acceptedDependencies:{},_acceptedErrorHandlers:{},_declinedDependencies:{},_selfAccepted:!1,_selfDeclined:!1,_selfInvalidated:!1,_disposeHandlers:[],_main:c,_requireSelf:function(){s=a.parents.slice(),e=c?void 0:t,p(t)},active:!0,accept:function(e,r,n){if(void 0===e)l._selfAccepted=!0;else if("function"==typeof e)l._selfAccepted=e;else if("object"==typeof e&&null!==e)for(var o=0;o<e.length;o++)l._acceptedDependencies[e[o]]=r||function(){},l._acceptedErrorHandlers[e[o]]=n;else l._acceptedDependencies[e]=r||function(){},l._acceptedErrorHandlers[e]=n},decline:function(e){if(void 0===e)l._selfDeclined=!0;else if("object"==typeof e&&null!==e)for(var r=0;r<e.length;r++)l._declinedDependencies[e[r]]=!0;else l._declinedDependencies[e]=!0},dispose:function(e){l._disposeHandlers.push(e)},addDisposeHandler:function(e){l._disposeHandlers.push(e)},removeDisposeHandler:function(e){var r=l._disposeHandlers.indexOf(e);r>0&&l._disposeHandlers.splice(r,1)},invalidate:function(){switch(this._selfInvalidated=!0,d){case"idle":r=[],Object.keys(p.hmrI).forEach(function(e){p.hmrI[e](t,r)}),u("ready");break;case"ready":Object.keys(p.hmrI).forEach(function(e){p.hmrI[e](t,r)});break;case"prepare":case"check":case"dispose":case"apply":(n=n||[]).push(t)}},check:f,apply:_,status:function(e){if(!e)return d;i.push(e)},addStatusHandler:function(e){i.push(e)},removeStatusHandler:function(e){var r=i.indexOf(e);r>=0&&i.splice(r,1)},data:o[t]};return e=void 0,l}(c.id,h),h.parents=s,h.children=[],s=[],c.require=m}),p.hmrC={},p.hmrI={};function u(e){d=e;for(var r=[],n=0;n<i.length;n++)r[n]=i[n].call(null,e);return Promise.all(r)}function l(){0==--a&&u("ready").then(function(){if(0===a){var e=c;c=[];for(var r=0;r<e.length;r++)e[r]()}})}function f(e){if("idle"!==d)throw Error("check() is only allowed in idle status");return u("check").then(p.hmrM).then(function(n){return n?u("prepare").then(function(){var o=[];return r=[],Promise.all(Object.keys(p.hmrC).reduce(function(e,t){return p.hmrC[t](n.c,n.r,n.m,e,r,o),e},[])).then(function(){var r;return r=function(){return e?h(e):u("ready").then(function(){return o})},0===a?r():new Promise(function(e){c.push(function(){e(r())})})})}):u(m()?"ready":"idle").then(function(){return null})})}function _(e){return"ready"!==d?Promise.resolve().then(function(){throw Error("apply() is only allowed in ready status (state: "+d+")")}):h(e)}function h(e){e=e||{},m();var o,t=r.map(function(r){return r(e)});r=void 0;var s=t.map(function(e){return e.error}).filter(Boolean);if(s.length>0)return u("abort").then(function(){throw s[0]});var i=u("dispose");t.forEach(function(e){e.dispose&&e.dispose()});var d=u("apply"),a=function(e){!o&&(o=e)},c=[];return t.forEach(function(e){if(e.apply){var r=e.apply(a);if(r)for(var n=0;n<r.length;n++)c.push(r[n])}}),Promise.all([i,d]).then(function(){return o?u("fail").then(function(){throw o}):n?h(e).then(function(e){return c.forEach(function(r){0>e.indexOf(r)&&e.push(r)}),e}):u("idle").then(function(){return c})})}function m(){if(n)return!r&&(r=[]),Object.keys(p.hmrI).forEach(function(e){n.forEach(function(n){p.hmrI[e](n,r)})}),n=void 0,!0}}(),!function(){function e(r){if("function"!=typeof WeakMap)return null;var n=new WeakMap,o=new WeakMap;return(e=function(e){return e?o:n})(r)}p.ir=function(r,n){if(!n&&r&&r.__esModule)return r;if(null===r||"object"!=typeof r&&"function"!=typeof r)return{default:r};var o=e(n);if(o&&o.has(r))return o.get(r);var t={},s=Object.defineProperty&&Object.getOwnPropertyDescriptor;for(var i in r)if("default"!==i&&Object.prototype.hasOwnProperty.call(r,i)){var d=s?Object.getOwnPropertyDescriptor(r,i):null;d&&(d.get||d.set)?Object.defineProperty(t,i,d):t[i]=r[i]}return t.default=r,o&&o.set(r,t),t}}(),p.es=function(e,r){return Object.keys(e).forEach(function(n){"default"!==n&&!Object.prototype.hasOwnProperty.call(r,n)&&Object.defineProperty(r,n,{enumerable:!0,get:function(){return e[n]}})}),e},e=[],p.O=function(r,n,o,t){if(n){t=t||0;for(var s=e.length;s>0&&e[s-1][2]>t;s--)e[s]=e[s-1];e[s]=[n,o,t];return}for(var i=1/0,s=0;s<e.length;s++){for(var n=e[s][0],o=e[s][1],t=e[s][2],d=!0,a=0;a<n.length;a++)i>=t&&Object.keys(p.O).every(function(e){return p.O[e](n[a])})?n.splice(a--,1):(d=!1,t<i&&(i=t));if(d){e.splice(s--,1);var c=o();void 0!==c&&(r=c)}}return r},r={"../interface/src/pages/app.stories.tsx":["interface_src_pages_app_stories_tsx","interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc","splitting~interface_src_pages_app_stories_tsx"],"../node_modules/@mdx-js/react/index.js":["1"],"../node_modules/@storybook/addon-docs/dist/DocsRenderer-EYKKDMVH.mjs":["interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc","node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs","splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs"],"../node_modules/@storybook/blocks/dist/Color-3YIJY6X7.mjs":["5"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/GlobalScrollAreaStyles-XIHNDKUY.mjs":["7"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/OverlayScrollbars-VAV6LJAB.mjs":["6"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/WithTooltip-3BDV6MYO.mjs":["0"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/formatter-UT3ZCDIS.mjs":["3"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/index.mjs":[],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/syntaxhighlighter-QTQ2UBB4.mjs":["4"]},p.el=function(e){var n=r[e];return void 0===n?Promise.resolve():n.length>1?Promise.all(n.map(p.e)):p.e(n[0])},p.g=function(){if("object"==typeof globalThis)return globalThis;try{return this||Function("return this")()}catch(e){if("object"==typeof window)return window}}(),p.h=function(){return"38b71b1227be3545"},p.hmrF=function(){return"runtime~main."+p.h()+".hot-update.json"},p.hu=function(e){return""+e+"."+p.h()+".hot-update.js"},p.k=function(e){return({0:"0.2d068005.iframe.bundle.css",1:"1.2d068005.iframe.bundle.css",2:"2.6be7d201.iframe.bundle.css",3:"3.2d068005.iframe.bundle.css",4:"4.2d068005.iframe.bundle.css",5:"5.2d068005.iframe.bundle.css",6:"6.2d068005.iframe.bundle.css",7:"7.2d068005.iframe.bundle.css",interface_src_pages_app_stories_tsx:"interface_src_pages_app_stories_tsx.b7a8ef91.iframe.bundle.css","interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc":"interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc.2d068005.iframe.bundle.css",main:"main.a41e7091.iframe.bundle.css","node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.2d068005.iframe.bundle.css","runtime~main":"runtime~main.2d068005.iframe.bundle.css","splitting~interface_src_pages_app_stories_tsx":"splitting~interface_src_pages_app_stories_tsx.2d068005.iframe.bundle.css","splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.2d068005.iframe.bundle.css"})[e]},n={},p.l=function(e,r,o,t){if(void 0!==o){for(var s,i,d=document.getElementsByTagName("script"),a=0;a<d.length;a++){var c=d[a];if(c.getAttribute("src")==e){s=c;break}}}!s&&(i=!0,(s=document.createElement("script")).charset="utf-8",s.timeout=120,s.src=e),n[e]=[r];var u=function(r,o){s.onerror=s.onload=null,clearTimeout(l);var t=n[e];if(delete n[e],s.parentNode&&s.parentNode.removeChild(s),t&&t.forEach(function(e){return e(o)}),r)return r(o)},l=setTimeout(u.bind(null,void 0,{type:"timeout",target:s}),12e4);s.onerror=u.bind(null,s.onerror),s.onload=u.bind(null,s.onload),i&&document.head.appendChild(s)},p.o=function(e,r){return Object.prototype.hasOwnProperty.call(e,r)},p.p="",p.u=function(e){return({0:"0.cb55ad32.iframe.bundle.js",1:"1.c887fbe5.iframe.bundle.js",3:"3.e348b969.iframe.bundle.js",4:"4.6983a53a.iframe.bundle.js",5:"5.732ef84a.iframe.bundle.js",6:"6.fef3f55d.iframe.bundle.js",7:"7.cfb8087a.iframe.bundle.js",interface_src_pages_app_stories_tsx:"interface_src_pages_app_stories_tsx.e1815198.iframe.bundle.js","interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc":"interface_src_pages_app_stories_tsx~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKD~d50ddc.70d00da8.iframe.bundle.js","node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.53c4e13c.iframe.bundle.js","splitting~interface_src_pages_app_stories_tsx":"splitting~interface_src_pages_app_stories_tsx.68ef4a05.iframe.bundle.js","splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.7bf02a7b.iframe.bundle.js"})[e]},!function(){var e,r,n,o,t,s={"runtime~main":0};p.f.j=function(e,r){var n=p.o(s,e)?s[e]:void 0;if(0!==n){if(n)r.push(n[2]);else if(e){var o=new Promise(function(r,o){n=s[e]=[r,o]});r.push(n[2]=o);var t=p.p+p.u(e),i=Error();p.l(t,function(r){if(p.o(s,e)&&(0!==(n=s[e])&&(s[e]=void 0),n)){var o=r&&("load"===r.type?"missing":r.type),t=r&&r.target&&r.target.src;i.message="Loading chunk "+e+" failed.\n("+o+": "+t+")",i.name="ChunkLoadError",i.type=o,i.request=t,n[1](i)}},"chunk-"+e,e)}else s[e]=0}};var i={};function d(r,n){return e=n,new Promise(function(e,n){var o=p.p+p.hu(r);i[r]=e;var t=Error();p.l(o,function(e){if(i[r]){i[r]=void 0;var o=e&&("load"===e.type?"missing":e.type),s=e&&e.target&&e.target.src;t.message="Loading hot update chunk "+r+" failed.\n("+o+": "+s+")",t.name="ChunkLoadError",t.type=o,t.request=s,n(t)}})})}function a(e){p.f&&delete p.f.jsonpHmr,r=void 0;function i(e,r){for(var n=0;n<r.length;n++){var o=r[n];-1===e.indexOf(o)&&e.push(o)}}var d,a={},c=[],u={},l=function(e){console.warn("[HMR] unexpected require("+e.id+") to disposed module")};for(var f in n)if(p.o(n,f)){var _,h=n[f];_=h?function(e){for(var r=[e],n={},o=r.map(function(e){return{chain:[e],id:e}});o.length>0;){var t=o.pop(),s=t.id,d=t.chain,a=p.c[s];if(!!a&&(!a.hot._selfAccepted||!!a.hot._selfInvalidated)){if(a.hot._selfDeclined)return{type:"self-declined",chain:d,moduleId:s};if(a.hot._main)return{type:"unaccepted",chain:d,moduleId:s};for(var c=0;c<a.parents.length;c++){var u=a.parents[c],l=p.c[u];if(!l)continue;if(l.hot._declinedDependencies[s])return{type:"declined",chain:d.concat([u]),moduleId:s,parentId:u};if(-1===r.indexOf(u)){if(l.hot._acceptedDependencies[s]){!n[u]&&(n[u]=[]),i(n[u],[s]);continue}delete n[u],r.push(u),o.push({chain:d.concat([u]),id:u})}}}}return{type:"accepted",moduleId:e,outdatedModules:r,outdatedDependencies:n}}(f):{type:"disposed",moduleId:f};var m=!1,b=!1,v=!1,y="";switch(_.chain&&(y="\nUpdate propagation: "+_.chain.join(" -> ")),_.type){case"self-declined":e.onDeclined&&e.onDeclined(_),!e.ignoreDeclined&&(m=Error("Aborted because of self decline: "+_.moduleId+y));break;case"declined":e.onDeclined&&e.onDeclined(_),!e.ignoreDeclined&&(m=Error("Aborted because of declined dependency: "+_.moduleId+" in "+_.parentId+y));break;case"unaccepted":e.onUnaccepted&&e.onUnaccepted(_),!e.ignoreUnaccepted&&(m=Error("Aborted because "+f+" is not accepted"+y));break;case"accepted":e.onAccepted&&e.onAccepted(_),b=!0;break;case"disposed":e.onDisposed&&e.onDisposed(_),v=!0;break;default:throw Error("Unexception type "+_.type)}if(m)return{error:m};if(b)for(f in u[f]=h,i(c,_.outdatedModules),_.outdatedDependencies)p.o(_.outdatedDependencies,f)&&(!a[f]&&(a[f]=[]),i(a[f],_.outdatedDependencies[f]));v&&(i(c,[_.moduleId]),u[f]=l)}n=void 0;for(var g=[],k=0;k<c.length;k++){var E=c[k],D=p.c[E];D&&(D.hot._selfAccepted||D.hot._main)&&u[E]!==l&&!D.hot._selfInvalidated&&g.push({module:E,require:D.hot._requireSelf,errorHandler:D.hot._selfAccepted})}return{dispose:function(){o.forEach(function(e){delete s[e]}),o=void 0;for(var e,r,n=c.slice();n.length>0;){var t=n.pop(),i=p.c[t];if(!!i){var u={},l=i.hot._disposeHandlers;for(k=0;k<l.length;k++)l[k].call(null,u);for(p.hmrD[t]=u,i.hot.active=!1,delete p.c[t],delete a[t],k=0;k<i.children.length;k++){var f=p.c[i.children[k]];f&&(e=f.parents.indexOf(t))>=0&&f.parents.splice(e,1)}}}for(var _ in a)if(p.o(a,_)&&(i=p.c[_]))for(k=0,d=a[_];k<d.length;k++)r=d[k],(e=i.children.indexOf(r))>=0&&i.children.splice(e,1)},apply:function(r){for(var n in u)p.o(u,n)&&(p.m[n]=u[n]);for(var o=0;o<t.length;o++)t[o](p);for(var s in a)if(p.o(a,s)){var i=p.c[s];if(i){d=a[s];for(var l=[],f=[],_=[],h=0;h<d.length;h++){var m=d[h],b=i.hot._acceptedDependencies[m],v=i.hot._acceptedErrorHandlers[m];if(b){if(-1!==l.indexOf(b))continue;l.push(b),f.push(v),_.push(m)}}for(var y=0;y<l.length;y++)try{l[y].call(null,d)}catch(n){if("function"==typeof f[y])try{f[y](n,{moduleId:s,dependencyId:_[y]})}catch(o){e.onErrored&&e.onErrored({type:"accept-error-handler-errored",moduleId:s,dependencyId:_[y],error:o,originalError:n}),!e.ignoreErrored&&(r(o),r(n))}else e.onErrored&&e.onErrored({type:"accept-errored",moduleId:s,dependencyId:_[y],error:n}),!e.ignoreErrored&&r(n)}}}for(var k=0;k<g.length;k++){var E=g[k],D=E.module;try{E.require(D)}catch(n){if("function"==typeof E.errorHandler)try{E.errorHandler(n,{moduleId:D,module:p.c[D]})}catch(o){e.onErrored&&e.onErrored({type:"self-accept-error-handler-errored",moduleId:D,error:o,originalError:n}),!e.ignoreErrored&&(r(o),r(n))}else e.onErrored&&e.onErrored({type:"self-accept-errored",moduleId:D,error:n}),!e.ignoreErrored&&r(n)}}return c}}}self.hotUpdate=function(r,o,s){for(var d in o)p.o(o,d)&&(n[d]=o[d],e&&e.push(d));s&&t.push(s),i[r]&&(i[r](),i[r]=void 0)},p.hmrI.jsonp=function(e,r){!n&&(n={},t=[],o=[],r.push(a)),!p.o(n,e)&&(n[e]=p.m[e])},p.hmrC.jsonp=function(e,i,c,u,l,f){l.push(a),r={},o=i,n=c.reduce(function(e,r){return e[r]=!1,e},{}),t=[],e.forEach(function(e){p.o(s,e)&&void 0!==s[e]?(u.push(d(e,f)),r[e]=!0):r[e]=!1}),p.f&&(p.f.jsonpHmr=function(e,n){r&&p.o(r,e)&&!r[e]&&(n.push(d(e)),r[e]=!0)})},p.hmrM=function(){if("undefined"==typeof fetch)throw Error("No browser support: need fetch API");return fetch(p.p+p.hmrF()).then(function(e){if(404!==e.status){if(!e.ok)throw Error("Failed to fetch update manifest "+e.statusText);return e.json()}})},p.O.j=function(e){return 0===s[e]};var c=function(e,r){var n=r[0],o=r[1],t=r[2],i,d,a=0;if(n.some(function(e){return 0!==s[e]})){for(i in o)p.o(o,i)&&(p.m[i]=o[i]);if(t)var c=t(p)}for(e&&e(r);a<n.length;a++)d=n[a],p.o(s,d)&&s[d]&&s[d][0](),s[d]=0;return p.O(c)},u=self.webpackChunk_monetr_stories=self.webpackChunk_monetr_stories||[];u.forEach(c.bind(null,0)),u.push=c.bind(null,u.push.bind(u))}(),o={2:0,main:0},t="webpack",s="data-webpack-loading",i=function(e,r,n,o){var i,d,a="chunk-"+e;if(!o){for(var c=document.getElementsByTagName("link"),u=0;u<c.length;u++){var l=c[u],f=l.getAttribute("href")||l.href;if(f&&!f.startsWith(p.p)&&(f=p.p+(f.startsWith("/")?f.slice(1):f)),"stylesheet"==l.rel&&(f&&f.startsWith(r)||l.getAttribute("data-webpack")==t+":"+a)){i=l;break}}if(!n)return i}!i&&(d=!0,(i=document.createElement("link")).setAttribute("data-webpack",t+":"+a),i.setAttribute(s,1),i.rel="stylesheet",i.href=r);var _=function(e,r){if(i.onerror=i.onload=null,i.removeAttribute(s),clearTimeout(h),r&&"load"!=r.type&&i.parentNode.removeChild(i),n(r),e)return e(r)};if(i.getAttribute(s)){var h=setTimeout(_.bind(null,void 0,{type:"timeout",target:i}),12e4);i.onerror=_.bind(null,i.onerror),i.onload=_.bind(null,i.onload)}else _(void 0,{type:"load",target:i});return o?document.head.insertBefore(i,o):d&&document.head.appendChild(i),i},p.f.css=function(e,r){var n=p.o(o,e)?o[e]:void 0;if(0!==n){if(n)r.push(n[2]);else if(["interface_src_pages_app_stories_tsx"].indexOf(e)>-1){var t=new Promise(function(r,t){n=o[e]=[r,t]});r.push(n[2]=t);var s=p.p+p.k(e),d=Error();i(e,s,function(r){if(p.o(o,e)&&(0!==(n=o[e])&&(o[e]=void 0),n)){if("load"!==r.type){var t=r&&r.type,s=r&&r.target&&r.target.src;d.message="Loading css chunk "+e+" failed.\n("+t+": "+s+")",d.name="ChunkLoadError",d.type=t,d.request=s,n[1](d)}else n[0]()}})}else o[e]=0}},d=[],a=[],c=function(e){return{dispose:function(){},apply:function(){for(a.forEach(function(e){e[1].sheet.disabled=!1});d.length;){var e=d.pop();e.parentNode&&e.parentNode.removeChild(e)}for(;a.length;)a.pop();return[]}}},u=function(e){return Array.from(e.sheet.cssRules,function(e){return e.cssText}).join()},p.hmrC.css=function(e,r,n,o,t,s){t.push(c),e.forEach(function(e){var r=p.k(e),n=p.p+r,t=i(e,n);t&&o.push(new Promise(function(r,o){var c=i(e,n+(0>n.indexOf("?")?"?":"&")+"hmr="+Date.now(),function(i){if("load"!==i.type){var l=Error(),f=i&&i.type,p=i&&i.target&&i.target.src;l.message="Loading css hot update chunk "+e+" failed.\n("+f+": "+p+")",l.name="ChunkLoadError",l.type=f,l.request=p,o(l)}else{try{if(u(t)==u(c))return c.parentNode&&c.parentNode.removeChild(c),r()}catch(e){}s.push(n),c.sheet.disabled=!0,d.push(t),a.push([e,c]),r()}},t)}))})}}();
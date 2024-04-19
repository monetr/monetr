!function(){var e,r,n,o,t,i,s,d,c,a,u,l={},f={};function p(e){var r=f[e];if(void 0!==r){if(void 0!==r.error)throw r.error;return r.exports}var n=f[e]={id:e,exports:{}};try{var o={id:e,module:n,factory:l[e],require:p};p.i.forEach(function(e){e(o)}),n=o.module,!o.factory&&console.error("undefined factory",e),o.factory.call(n.exports,n,n.exports,o.require)}catch(e){throw n.error=e,e}return n.exports}p.m=l,p.c=f,p.i=[],p.f={},p.e=function(e){return Promise.all(Object.keys(p.f).reduce(function(r,n){return p.f[n](e,r),r},[]))},!function(){var e,r,n,o={},t=p.c,i=[],s=[],d="idle",c=0,a=[];p.hmrD=o,p.i.push(function(a){var m=a.module,_=function(r,n){var o=t[n];if(!o)return r;var s=function(s){if(o.hot.active){if(t[s]){var d=t[s].parents;-1===d.indexOf(n)&&d.push(n)}else i=[n],e=s;-1===o.children.indexOf(s)&&o.children.push(s)}else console.warn("[HMR] unexpected require("+s+") from disposed module "+n),i=[];return r(s)},a=function(e){return{configurable:!0,enumerable:!0,get:function(){return r[e]},set:function(n){r[e]=n}}};for(var f in r)Object.prototype.hasOwnProperty.call(r,f)&&"e"!==f&&Object.defineProperty(s,f,a(f));return s.e=function(e){return function(e){switch(d){case"ready":u("prepare");case"prepare":return c++,e.then(l,l),e;default:return e}}(r.e(e))},s}(a.require,a.id);m.hot=function(t,c){var a=e!==t,l={_acceptedDependencies:{},_acceptedErrorHandlers:{},_declinedDependencies:{},_selfAccepted:!1,_selfDeclined:!1,_selfInvalidated:!1,_disposeHandlers:[],_main:a,_requireSelf:function(){i=c.parents.slice(),e=a?void 0:t,p(t)},active:!0,accept:function(e,r,n){if(void 0===e)l._selfAccepted=!0;else if("function"==typeof e)l._selfAccepted=e;else if("object"==typeof e&&null!==e)for(var o=0;o<e.length;o++)l._acceptedDependencies[e[o]]=r||function(){},l._acceptedErrorHandlers[e[o]]=n;else l._acceptedDependencies[e]=r||function(){},l._acceptedErrorHandlers[e]=n},decline:function(e){if(void 0===e)l._selfDeclined=!0;else if("object"==typeof e&&null!==e)for(var r=0;r<e.length;r++)l._declinedDependencies[e[r]]=!0;else l._declinedDependencies[e]=!0},dispose:function(e){l._disposeHandlers.push(e)},addDisposeHandler:function(e){l._disposeHandlers.push(e)},removeDisposeHandler:function(e){var r=l._disposeHandlers.indexOf(e);r>0&&l._disposeHandlers.splice(r,1)},invalidate:function(){switch(this._selfInvalidated=!0,d){case"idle":r=[],Object.keys(p.hmrI).forEach(function(e){p.hmrI[e](t,r)}),u("ready");break;case"ready":Object.keys(p.hmrI).forEach(function(e){p.hmrI[e](t,r)});break;case"prepare":case"check":case"dispose":case"apply":(n=n||[]).push(t)}},check:f,apply:h,status:function(e){if(!e)return d;s.push(e)},addStatusHandler:function(e){s.push(e)},removeStatusHandler:function(e){var r=s.indexOf(e);r>=0&&s.splice(r,1)},data:o[t]};return e=void 0,l}(a.id,m),m.parents=i,m.children=[],i=[],a.require=_}),p.hmrC={},p.hmrI={};function u(e){d=e;for(var r=[],n=0;n<s.length;n++)r[n]=s[n].call(null,e);return Promise.all(r)}function l(){0==--c&&u("ready").then(function(){if(0===c){var e=a;a=[];for(var r=0;r<e.length;r++)e[r]()}})}function f(e){if("idle"!==d)throw Error("check() is only allowed in idle status");return u("check").then(p.hmrM).then(function(n){return n?u("prepare").then(function(){var o=[];return r=[],Promise.all(Object.keys(p.hmrC).reduce(function(e,t){return p.hmrC[t](n.c,n.r,n.m,e,r,o),e},[])).then(function(){var r;return r=function(){return e?m(e):u("ready").then(function(){return o})},0===c?r():new Promise(function(e){a.push(function(){e(r())})})})}):u(_()?"ready":"idle").then(function(){return null})})}function h(e){return"ready"!==d?Promise.resolve().then(function(){throw Error("apply() is only allowed in ready status (state: "+d+")")}):m(e)}function m(e){e=e||{},_();var o,t=r.map(function(r){return r(e)});r=void 0;var i=t.map(function(e){return e.error}).filter(Boolean);if(i.length>0)return u("abort").then(function(){throw i[0]});var s=u("dispose");t.forEach(function(e){e.dispose&&e.dispose()});var d=u("apply"),c=function(e){!o&&(o=e)},a=[];return t.forEach(function(e){if(e.apply){var r=e.apply(c);if(r)for(var n=0;n<r.length;n++)a.push(r[n])}}),Promise.all([s,d]).then(function(){return o?u("fail").then(function(){throw o}):n?m(e).then(function(e){return a.forEach(function(r){0>e.indexOf(r)&&e.push(r)}),e}):u("idle").then(function(){return a})})}function _(){if(n)return!r&&(r=[]),Object.keys(p.hmrI).forEach(function(e){n.forEach(function(n){p.hmrI[e](n,r)})}),n=void 0,!0}}(),!function(){function e(r){if("function"!=typeof WeakMap)return null;var n=new WeakMap,o=new WeakMap;return(e=function(e){return e?o:n})(r)}p.ir=function(r,n){if(!n&&r&&r.__esModule)return r;if(null===r||"object"!=typeof r&&"function"!=typeof r)return{default:r};var o=e(n);if(o&&o.has(r))return o.get(r);var t={},i=Object.defineProperty&&Object.getOwnPropertyDescriptor;for(var s in r)if("default"!==s&&Object.prototype.hasOwnProperty.call(r,s)){var d=i?Object.getOwnPropertyDescriptor(r,s):null;d&&(d.get||d.set)?Object.defineProperty(t,s,d):t[s]=r[s]}return t.default=r,o&&o.set(r,t),t}}(),p.es=function(e,r){return Object.keys(e).forEach(function(n){"default"!==n&&!Object.prototype.hasOwnProperty.call(r,n)&&Object.defineProperty(r,n,{enumerable:!0,get:function(){return e[n]}})}),e},e=[],p.O=function(r,n,o,t){if(n){t=t||0;for(var i=e.length;i>0&&e[i-1][2]>t;i--)e[i]=e[i-1];e[i]=[n,o,t];return}for(var s=1/0,i=0;i<e.length;i++){for(var n=e[i][0],o=e[i][1],t=e[i][2],d=!0,c=0;c<n.length;c++)s>=t&&Object.keys(p.O).every(function(e){return p.O[e](n[c])})?n.splice(c--,1):(d=!1,t<s&&(s=t));if(d){e.splice(i--,1);var a=o();void 0!==a&&(r=a)}}return r},r={"../interface/src/pages/new.stories.tsx":["interface_src_pages_new_stories_tsx","splitting~interface_src_pages_new_stories_tsx"],"../node_modules/@mdx-js/react/index.js":["3"],"../node_modules/@storybook/addon-docs/dist/DocsRenderer-EYKKDMVH.mjs":["node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs","splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs"],"../node_modules/@storybook/blocks/dist/Color-3YIJY6X7.mjs":["7"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/GlobalScrollAreaStyles-XIHNDKUY.mjs":["1"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/OverlayScrollbars-VAV6LJAB.mjs":["2"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/WithTooltip-3BDV6MYO.mjs":["0"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/formatter-UT3ZCDIS.mjs":["5"],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/index.mjs":[],"../node_modules/storybook-builder-rspack/node_modules/@storybook/components/dist/syntaxhighlighter-QTQ2UBB4.mjs":["6"]},p.el=function(e){var n=r[e];return void 0===n?Promise.resolve():n.length>1?Promise.all(n.map(p.e)):p.e(n[0])},p.g=function(){if("object"==typeof globalThis)return globalThis;try{return this||Function("return this")()}catch(e){if("object"==typeof window)return window}}(),p.h=function(){return"b131b10744a9a5a8"},p.hmrF=function(){return"runtime~main."+p.h()+".hot-update.json"},p.hu=function(e){return""+e+"."+p.h()+".hot-update.js"},p.k=function(e){return({0:"0.2d068005.iframe.bundle.css",1:"1.2d068005.iframe.bundle.css",2:"2.2d068005.iframe.bundle.css",3:"3.2d068005.iframe.bundle.css",4:"4.6375a111.iframe.bundle.css",5:"5.2d068005.iframe.bundle.css",6:"6.2d068005.iframe.bundle.css",7:"7.2d068005.iframe.bundle.css",interface_src_pages_new_stories_tsx:"interface_src_pages_new_stories_tsx.0d777361.iframe.bundle.css",main:"main.159f6d89.iframe.bundle.css","node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.2d068005.iframe.bundle.css","runtime~main":"runtime~main.2d068005.iframe.bundle.css","splitting~interface_src_pages_new_stories_tsx":"splitting~interface_src_pages_new_stories_tsx.2d068005.iframe.bundle.css","splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.2d068005.iframe.bundle.css"})[e]},n={},p.l=function(e,r,o,t){if(void 0!==o){for(var i,s,d=document.getElementsByTagName("script"),c=0;c<d.length;c++){var a=d[c];if(a.getAttribute("src")==e){i=a;break}}}!i&&(s=!0,(i=document.createElement("script")).charset="utf-8",i.timeout=120,i.src=e),n[e]=[r];var u=function(r,o){i.onerror=i.onload=null,clearTimeout(l);var t=n[e];if(delete n[e],i.parentNode&&i.parentNode.removeChild(i),t&&t.forEach(function(e){return e(o)}),r)return r(o)},l=setTimeout(u.bind(null,void 0,{type:"timeout",target:i}),12e4);i.onerror=u.bind(null,i.onerror),i.onload=u.bind(null,i.onload),s&&document.head.appendChild(i)},p.o=function(e,r){return Object.prototype.hasOwnProperty.call(e,r)},p.p="",p.u=function(e){return({0:"0.cb55ad32.iframe.bundle.js",1:"1.eb8421a3.iframe.bundle.js",2:"2.f29653ec.iframe.bundle.js",3:"3.9006eb46.iframe.bundle.js",5:"5.0792f14b.iframe.bundle.js",6:"6.1f4a803e.iframe.bundle.js",7:"7.b0c19684.iframe.bundle.js",interface_src_pages_new_stories_tsx:"interface_src_pages_new_stories_tsx.2e405446.iframe.bundle.js","node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.53c4e13c.iframe.bundle.js","splitting~interface_src_pages_new_stories_tsx":"splitting~interface_src_pages_new_stories_tsx.050c950b.iframe.bundle.js","splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs":"splitting~node_modules_storybook_addon-docs_dist_DocsRenderer-EYKKDMVH_mjs.e3d61476.iframe.bundle.js"})[e]},!function(){var e,r,n,o,t,i={"runtime~main":0};p.f.j=function(e,r){var n=p.o(i,e)?i[e]:void 0;if(0!==n){if(n)r.push(n[2]);else if(e){var o=new Promise(function(r,o){n=i[e]=[r,o]});r.push(n[2]=o);var t=p.p+p.u(e),s=Error();p.l(t,function(r){if(p.o(i,e)&&(0!==(n=i[e])&&(i[e]=void 0),n)){var o=r&&("load"===r.type?"missing":r.type),t=r&&r.target&&r.target.src;s.message="Loading chunk "+e+" failed.\n("+o+": "+t+")",s.name="ChunkLoadError",s.type=o,s.request=t,n[1](s)}},"chunk-"+e,e)}else i[e]=0}};var s={};function d(r,n){return e=n,new Promise(function(e,n){var o=p.p+p.hu(r);s[r]=e;var t=Error();p.l(o,function(e){if(s[r]){s[r]=void 0;var o=e&&("load"===e.type?"missing":e.type),i=e&&e.target&&e.target.src;t.message="Loading hot update chunk "+r+" failed.\n("+o+": "+i+")",t.name="ChunkLoadError",t.type=o,t.request=i,n(t)}})})}function c(e){p.f&&delete p.f.jsonpHmr,r=void 0;function s(e,r){for(var n=0;n<r.length;n++){var o=r[n];-1===e.indexOf(o)&&e.push(o)}}var d,c={},a=[],u={},l=function(e){console.warn("[HMR] unexpected require("+e.id+") to disposed module")};for(var f in n)if(p.o(n,f)){var h,m=n[f];h=m?function(e){for(var r=[e],n={},o=r.map(function(e){return{chain:[e],id:e}});o.length>0;){var t=o.pop(),i=t.id,d=t.chain,c=p.c[i];if(!!c&&(!c.hot._selfAccepted||!!c.hot._selfInvalidated)){if(c.hot._selfDeclined)return{type:"self-declined",chain:d,moduleId:i};if(c.hot._main)return{type:"unaccepted",chain:d,moduleId:i};for(var a=0;a<c.parents.length;a++){var u=c.parents[a],l=p.c[u];if(!l)continue;if(l.hot._declinedDependencies[i])return{type:"declined",chain:d.concat([u]),moduleId:i,parentId:u};if(-1===r.indexOf(u)){if(l.hot._acceptedDependencies[i]){!n[u]&&(n[u]=[]),s(n[u],[i]);continue}delete n[u],r.push(u),o.push({chain:d.concat([u]),id:u})}}}}return{type:"accepted",moduleId:e,outdatedModules:r,outdatedDependencies:n}}(f):{type:"disposed",moduleId:f};var _=!1,v=!1,b=!1,y="";switch(h.chain&&(y="\nUpdate propagation: "+h.chain.join(" -> ")),h.type){case"self-declined":e.onDeclined&&e.onDeclined(h),!e.ignoreDeclined&&(_=Error("Aborted because of self decline: "+h.moduleId+y));break;case"declined":e.onDeclined&&e.onDeclined(h),!e.ignoreDeclined&&(_=Error("Aborted because of declined dependency: "+h.moduleId+" in "+h.parentId+y));break;case"unaccepted":e.onUnaccepted&&e.onUnaccepted(h),!e.ignoreUnaccepted&&(_=Error("Aborted because "+f+" is not accepted"+y));break;case"accepted":e.onAccepted&&e.onAccepted(h),v=!0;break;case"disposed":e.onDisposed&&e.onDisposed(h),b=!0;break;default:throw Error("Unexception type "+h.type)}if(_)return{error:_};if(v)for(f in u[f]=m,s(a,h.outdatedModules),h.outdatedDependencies)p.o(h.outdatedDependencies,f)&&(!c[f]&&(c[f]=[]),s(c[f],h.outdatedDependencies[f]));b&&(s(a,[h.moduleId]),u[f]=l)}n=void 0;for(var g=[],k=0;k<a.length;k++){var j=a[k],E=p.c[j];E&&(E.hot._selfAccepted||E.hot._main)&&u[j]!==l&&!E.hot._selfInvalidated&&g.push({module:j,require:E.hot._requireSelf,errorHandler:E.hot._selfAccepted})}return{dispose:function(){o.forEach(function(e){delete i[e]}),o=void 0;for(var e,r,n=a.slice();n.length>0;){var t=n.pop(),s=p.c[t];if(!!s){var u={},l=s.hot._disposeHandlers;for(k=0;k<l.length;k++)l[k].call(null,u);for(p.hmrD[t]=u,s.hot.active=!1,delete p.c[t],delete c[t],k=0;k<s.children.length;k++){var f=p.c[s.children[k]];f&&(e=f.parents.indexOf(t))>=0&&f.parents.splice(e,1)}}}for(var h in c)if(p.o(c,h)&&(s=p.c[h]))for(k=0,d=c[h];k<d.length;k++)r=d[k],(e=s.children.indexOf(r))>=0&&s.children.splice(e,1)},apply:function(r){for(var n in u)p.o(u,n)&&(p.m[n]=u[n]);for(var o=0;o<t.length;o++)t[o](p);for(var i in c)if(p.o(c,i)){var s=p.c[i];if(s){d=c[i];for(var l=[],f=[],h=[],m=0;m<d.length;m++){var _=d[m],v=s.hot._acceptedDependencies[_],b=s.hot._acceptedErrorHandlers[_];if(v){if(-1!==l.indexOf(v))continue;l.push(v),f.push(b),h.push(_)}}for(var y=0;y<l.length;y++)try{l[y].call(null,d)}catch(n){if("function"==typeof f[y])try{f[y](n,{moduleId:i,dependencyId:h[y]})}catch(o){e.onErrored&&e.onErrored({type:"accept-error-handler-errored",moduleId:i,dependencyId:h[y],error:o,originalError:n}),!e.ignoreErrored&&(r(o),r(n))}else e.onErrored&&e.onErrored({type:"accept-errored",moduleId:i,dependencyId:h[y],error:n}),!e.ignoreErrored&&r(n)}}}for(var k=0;k<g.length;k++){var j=g[k],E=j.module;try{j.require(E)}catch(n){if("function"==typeof j.errorHandler)try{j.errorHandler(n,{moduleId:E,module:p.c[E]})}catch(o){e.onErrored&&e.onErrored({type:"self-accept-error-handler-errored",moduleId:E,error:o,originalError:n}),!e.ignoreErrored&&(r(o),r(n))}else e.onErrored&&e.onErrored({type:"self-accept-errored",moduleId:E,error:n}),!e.ignoreErrored&&r(n)}}return a}}}self.hotUpdate=function(r,o,i){for(var d in o)p.o(o,d)&&(n[d]=o[d],e&&e.push(d));i&&t.push(i),s[r]&&(s[r](),s[r]=void 0)},p.hmrI.jsonp=function(e,r){!n&&(n={},t=[],o=[],r.push(c)),!p.o(n,e)&&(n[e]=p.m[e])},p.hmrC.jsonp=function(e,s,a,u,l,f){l.push(c),r={},o=s,n=a.reduce(function(e,r){return e[r]=!1,e},{}),t=[],e.forEach(function(e){p.o(i,e)&&void 0!==i[e]?(u.push(d(e,f)),r[e]=!0):r[e]=!1}),p.f&&(p.f.jsonpHmr=function(e,n){r&&p.o(r,e)&&!r[e]&&(n.push(d(e)),r[e]=!0)})},p.hmrM=function(){if("undefined"==typeof fetch)throw Error("No browser support: need fetch API");return fetch(p.p+p.hmrF()).then(function(e){if(404!==e.status){if(!e.ok)throw Error("Failed to fetch update manifest "+e.statusText);return e.json()}})},p.O.j=function(e){return 0===i[e]};var a=function(e,r){var n=r[0],o=r[1],t=r[2],s,d,c=0;if(n.some(function(e){return 0!==i[e]})){for(s in o)p.o(o,s)&&(p.m[s]=o[s]);if(t)var a=t(p)}for(e&&e(r);c<n.length;c++)d=n[c],p.o(i,d)&&i[d]&&i[d][0](),i[d]=0;return p.O(a)},u=self.webpackChunk_monetr_stories=self.webpackChunk_monetr_stories||[];u.forEach(a.bind(null,0)),u.push=a.bind(null,u.push.bind(u))}(),o={4:0,main:0},t="webpack",i="data-webpack-loading",s=function(e,r,n,o){var s,d,c="chunk-"+e;if(!o){for(var a=document.getElementsByTagName("link"),u=0;u<a.length;u++){var l=a[u],f=l.getAttribute("href")||l.href;if(f&&!f.startsWith(p.p)&&(f=p.p+(f.startsWith("/")?f.slice(1):f)),"stylesheet"==l.rel&&(f&&f.startsWith(r)||l.getAttribute("data-webpack")==t+":"+c)){s=l;break}}if(!n)return s}!s&&(d=!0,(s=document.createElement("link")).setAttribute("data-webpack",t+":"+c),s.setAttribute(i,1),s.rel="stylesheet",s.href=r);var h=function(e,r){if(s.onerror=s.onload=null,s.removeAttribute(i),clearTimeout(m),r&&"load"!=r.type&&s.parentNode.removeChild(s),n(r),e)return e(r)};if(s.getAttribute(i)){var m=setTimeout(h.bind(null,void 0,{type:"timeout",target:s}),12e4);s.onerror=h.bind(null,s.onerror),s.onload=h.bind(null,s.onload)}else h(void 0,{type:"load",target:s});return o?document.head.insertBefore(s,o):d&&document.head.appendChild(s),s},p.f.css=function(e,r){var n=p.o(o,e)?o[e]:void 0;if(0!==n){if(n)r.push(n[2]);else if(["interface_src_pages_new_stories_tsx"].indexOf(e)>-1){var t=new Promise(function(r,t){n=o[e]=[r,t]});r.push(n[2]=t);var i=p.p+p.k(e),d=Error();s(e,i,function(r){if(p.o(o,e)&&(0!==(n=o[e])&&(o[e]=void 0),n)){if("load"!==r.type){var t=r&&r.type,i=r&&r.target&&r.target.src;d.message="Loading css chunk "+e+" failed.\n("+t+": "+i+")",d.name="ChunkLoadError",d.type=t,d.request=i,n[1](d)}else n[0]()}})}else o[e]=0}},d=[],c=[],a=function(e){return{dispose:function(){},apply:function(){for(c.forEach(function(e){e[1].sheet.disabled=!1});d.length;){var e=d.pop();e.parentNode&&e.parentNode.removeChild(e)}for(;c.length;)c.pop();return[]}}},u=function(e){return Array.from(e.sheet.cssRules,function(e){return e.cssText}).join()},p.hmrC.css=function(e,r,n,o,t,i){t.push(a),e.forEach(function(e){var r=p.k(e),n=p.p+r,t=s(e,n);t&&o.push(new Promise(function(r,o){var a=s(e,n+(0>n.indexOf("?")?"?":"&")+"hmr="+Date.now(),function(s){if("load"!==s.type){var l=Error(),f=s&&s.type,p=s&&s.target&&s.target.src;l.message="Loading css hot update chunk "+e+" failed.\n("+f+": "+p+")",l.name="ChunkLoadError",l.type=f,l.request=p,o(l)}else{try{if(u(t)==u(a))return a.parentNode&&a.parentNode.removeChild(a),r()}catch(e){}i.push(n),a.sheet.disabled=!0,d.push(t),c.push([e,a]),r()}},t)}))})}}();
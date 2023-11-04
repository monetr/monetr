let INTEGRITY_CHECKSUM="3d6b9f06410d179a7f7404d4bf4c3c70",activeClientIds=new Set;async function handleRequest(e,t){let n=await resolveMainClient(e),a=await getResponse(e,n,t);return n&&activeClientIds.has(n.id)&&!async function(){let e=a.clone();sendToClient(n,{type:"RESPONSE",payload:{requestId:t,type:e.type,ok:e.ok,status:e.status,statusText:e.statusText,body:null===e.body?null:await e.text(),headers:Object.fromEntries(e.headers.entries()),redirected:e.redirected}})}(),a}async function resolveMainClient(e){let t=await self.clients.get(e.clientId);if(t?.frameType==="top-level")return t;let n=await self.clients.matchAll({type:"window"});return n.filter(e=>"visible"===e.visibilityState).find(e=>activeClientIds.has(e.id))}async function getResponse(e,t,n){let{request:a}=e,i=a.clone();function s(){let e=Object.fromEntries(i.headers.entries());return delete e["x-msw-bypass"],fetch(i,{headers:e})}if(!t||!activeClientIds.has(t.id)||"true"===a.headers.get("x-msw-bypass"))return s();let r=await sendToClient(t,{type:"REQUEST",payload:{id:n,url:a.url,method:a.method,headers:Object.fromEntries(a.headers.entries()),cache:a.cache,mode:a.mode,credentials:a.credentials,destination:a.destination,integrity:a.integrity,redirect:a.redirect,referrer:a.referrer,referrerPolicy:a.referrerPolicy,body:await a.text(),bodyUsed:a.bodyUsed,keepalive:a.keepalive}});switch(r.type){case"MOCK_RESPONSE":return respondWithMock(r.data);case"MOCK_NOT_FOUND":break;case"NETWORK_ERROR":{let{name:e,message:t}=r.data,n=Error(t);throw n.name=e,n}}return s()}function sendToClient(e,t){return new Promise((n,a)=>{let i=new MessageChannel;i.port1.onmessage=e=>{if(e.data&&e.data.error)return a(e.data.error);n(e.data)},e.postMessage(t,[i.port2])})}function sleep(e){return new Promise(t=>{setTimeout(t,e)})}async function respondWithMock(e){return await sleep(e.delay),new Response(e.body,e)}self.addEventListener("install",function(){self.skipWaiting()}),self.addEventListener("activate",function(e){e.waitUntil(self.clients.claim())}),self.addEventListener("message",async function(e){let t=e.source.id;if(!t||!self.clients)return;let n=await self.clients.get(t);if(!n)return;let a=await self.clients.matchAll({type:"window"});switch(e.data){case"KEEPALIVE_REQUEST":sendToClient(n,{type:"KEEPALIVE_RESPONSE"});break;case"INTEGRITY_CHECK_REQUEST":sendToClient(n,{type:"INTEGRITY_CHECK_RESPONSE",payload:"3d6b9f06410d179a7f7404d4bf4c3c70"});break;case"MOCK_ACTIVATE":activeClientIds.add(t),sendToClient(n,{type:"MOCKING_ENABLED",payload:!0});break;case"MOCK_DEACTIVATE":activeClientIds.delete(t);break;case"CLIENT_CLOSED":{activeClientIds.delete(t);let e=a.filter(e=>e.id!==t);0===e.length&&self.registration.unregister()}}}),self.addEventListener("fetch",function(e){let{request:t}=e,n=t.headers.get("accept")||"";if(n.includes("text/event-stream")||"navigate"===t.mode||"only-if-cached"===t.cache&&"same-origin"!==t.mode||0===activeClientIds.size)return;let a=Math.random().toString(16).slice(2);e.respondWith(handleRequest(e,a).catch(e=>{if("NetworkError"===e.name){console.warn('[MSW] Successfully emulated a network error for the "%s %s" request.',t.method,t.url);return}console.error(`\
[MSW] Caught an exception from the "%s %s" request (%s). This is probably not a problem with Mock Service Worker. There is likely an additional logging output above.`,t.method,t.url,`${e.name}: ${e.message}`)}))});
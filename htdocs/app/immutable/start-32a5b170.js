import{S as pt,i as mt,s as _t,a as gt,e as H,c as wt,b as Y,g as _e,t as K,d as ge,f as M,h as G,j as yt,o as je,k as bt,l as vt,m as Et,n as Re,p as B,q as kt,r as St,u as Rt,v as Le,w as Z,x as Q,y as De,z as x,A as ee,B as pe}from"./chunks/index-867754b5.js";import{S as lt,a as ct,I as $,g as Qe,f as xe,b as Ie,c as me,s as F,i as et,d as se,e as X,P as tt,h as Lt,j as It,k as At}from"./chunks/singletons-3caa26a7.js";function Pt(a,e){return a==="/"||e==="ignore"?a:e==="never"?a.endsWith("/")?a.slice(0,-1):a:e==="always"&&!a.endsWith("/")?a+"/":a}function Ot(a){return a.split("%25").map(decodeURI).join("%25")}function Ut(a){for(const e in a)a[e]=decodeURIComponent(a[e]);return a}const Nt=["href","pathname","search","searchParams","toString","toJSON"];function jt(a,e){const n=new URL(a);for(const i of Nt){let o=n[i];Object.defineProperty(n,i,{get(){return e(),o},enumerable:!0,configurable:!0})}return Tt(n),n}function Tt(a){Object.defineProperty(a,"hash",{get(){throw new Error("Cannot access event.url.hash. Consider using `$page.url.hash` inside a component instead")}})}const Dt="/__data.json";function Ct(a){return a.replace(/\/$/,"")+Dt}function ft(a){try{return JSON.parse(sessionStorage[a])}catch{}}function nt(a,e){const n=JSON.stringify(e);try{sessionStorage[a]=n}catch{}}function qt(...a){let e=5381;for(const n of a)if(typeof n=="string"){let i=n.length;for(;i;)e=e*33^n.charCodeAt(--i)}else if(ArrayBuffer.isView(n)){const i=new Uint8Array(n.buffer,n.byteOffset,n.byteLength);let o=i.length;for(;o;)e=e*33^i[--o]}else throw new TypeError("value must be a string or TypedArray");return(e>>>0).toString(36)}const we=window.fetch;window.fetch=(a,e)=>((a instanceof Request?a.method:(e==null?void 0:e.method)||"GET")!=="GET"&&oe.delete(Ce(a)),we(a,e));const oe=new Map;function Vt(a,e){const n=Ce(a,e),i=document.querySelector(n);if(i!=null&&i.textContent){const{body:o,...h}=JSON.parse(i.textContent),t=i.getAttribute("data-ttl");return t&&oe.set(n,{body:o,init:h,ttl:1e3*Number(t)}),Promise.resolve(new Response(o,h))}return we(a,e)}function $t(a,e,n){if(oe.size>0){const i=Ce(a,n),o=oe.get(i);if(o){if(performance.now()<o.ttl&&["default","force-cache","only-if-cached",void 0].includes(n==null?void 0:n.cache))return new Response(o.body,o.init);oe.delete(i)}}return we(e,n)}function Ce(a,e){let i=`script[data-sveltekit-fetched][data-url=${JSON.stringify(a instanceof Request?a.url:a)}]`;if(e!=null&&e.headers||e!=null&&e.body){const o=[];e.headers&&o.push([...new Headers(e.headers)].join(",")),e.body&&(typeof e.body=="string"||ArrayBuffer.isView(e.body))&&o.push(e.body),i+=`[data-hash="${qt(...o)}"]`}return i}const Bt=/^(\[)?(\.\.\.)?(\w+)(?:=(\w+))?(\])?$/;function Ft(a){const e=[];return{pattern:a==="/"?/^\/$/:new RegExp(`^${Kt(a).map(i=>{const o=/^\[\.\.\.(\w+)(?:=(\w+))?\]$/.exec(i);if(o)return e.push({name:o[1],matcher:o[2],optional:!1,rest:!0,chained:!0}),"(?:/(.*))?";const h=/^\[\[(\w+)(?:=(\w+))?\]\]$/.exec(i);if(h)return e.push({name:h[1],matcher:h[2],optional:!0,rest:!1,chained:!0}),"(?:/([^/]+))?";if(!i)return;const t=i.split(/\[(.+?)\](?!\])/);return"/"+t.map((m,p)=>{if(p%2){if(m.startsWith("x+"))return Ae(String.fromCharCode(parseInt(m.slice(2),16)));if(m.startsWith("u+"))return Ae(String.fromCharCode(...m.slice(2).split("-").map(L=>parseInt(L,16))));const g=Bt.exec(m);if(!g)throw new Error(`Invalid param: ${m}. Params and matcher names can only have underscores and alphanumeric characters.`);const[,y,N,D,C]=g;return e.push({name:D,matcher:C,optional:!!y,rest:!!N,chained:N?p===1&&t[0]==="":!1}),N?"(.*?)":y?"([^/]*)?":"([^/]+?)"}return Ae(m)}).join("")}).join("")}/?$`),params:e}}function Ht(a){return!/^\([^)]+\)$/.test(a)}function Kt(a){return a.slice(1).split("/").filter(Ht)}function Mt(a,e,n){const i={},o=a.slice(1);let h=0;for(let t=0;t<e.length;t+=1){const u=e[t],m=o[t-h];if(u.chained&&u.rest&&h){i[u.name]=o.slice(t-h,t+1).filter(p=>p).join("/"),h=0;continue}if(m===void 0){u.rest&&(i[u.name]="");continue}if(!u.matcher||n[u.matcher](m)){i[u.name]=m;continue}if(u.optional&&u.chained){h++;continue}return}if(!h)return i}function Ae(a){return a.normalize().replace(/[[\]]/g,"\\$&").replace(/%/g,"%25").replace(/\//g,"%2[Ff]").replace(/\?/g,"%3[Ff]").replace(/#/g,"%23").replace(/[.*+?^${}()|\\]/g,"\\$&")}function Gt(a,e,n,i){const o=new Set(e);return Object.entries(n).map(([u,[m,p,g]])=>{const{pattern:y,params:N}=Ft(u),D={id:u,exec:C=>{const L=y.exec(C);if(L)return Mt(L,N,i)},errors:[1,...g||[]].map(C=>a[C]),layouts:[0,...p||[]].map(t),leaf:h(m)};return D.errors.length=D.layouts.length=Math.max(D.errors.length,D.layouts.length),D});function h(u){const m=u<0;return m&&(u=~u),[m,a[u]]}function t(u){return u===void 0?u:[o.has(u),a[u]]}}function Jt(a){let e,n,i;var o=a[1][0];function h(t){return{props:{data:t[3],form:t[2]}}}return o&&(e=Z(o,h(a)),a[12](e)),{c(){e&&Q(e.$$.fragment),n=H()},l(t){e&&De(e.$$.fragment,t),n=H()},m(t,u){e&&x(e,t,u),Y(t,n,u),i=!0},p(t,u){const m={};if(u&8&&(m.data=t[3]),u&4&&(m.form=t[2]),o!==(o=t[1][0])){if(e){_e();const p=e;K(p.$$.fragment,1,0,()=>{ee(p,1)}),ge()}o?(e=Z(o,h(t)),t[12](e),Q(e.$$.fragment),M(e.$$.fragment,1),x(e,n.parentNode,n)):e=null}else o&&e.$set(m)},i(t){i||(e&&M(e.$$.fragment,t),i=!0)},o(t){e&&K(e.$$.fragment,t),i=!1},d(t){a[12](null),t&&G(n),e&&ee(e,t)}}}function zt(a){let e,n,i;var o=a[1][0];function h(t){return{props:{data:t[3],$$slots:{default:[Wt]},$$scope:{ctx:t}}}}return o&&(e=Z(o,h(a)),a[11](e)),{c(){e&&Q(e.$$.fragment),n=H()},l(t){e&&De(e.$$.fragment,t),n=H()},m(t,u){e&&x(e,t,u),Y(t,n,u),i=!0},p(t,u){const m={};if(u&8&&(m.data=t[3]),u&8215&&(m.$$scope={dirty:u,ctx:t}),o!==(o=t[1][0])){if(e){_e();const p=e;K(p.$$.fragment,1,0,()=>{ee(p,1)}),ge()}o?(e=Z(o,h(t)),t[11](e),Q(e.$$.fragment),M(e.$$.fragment,1),x(e,n.parentNode,n)):e=null}else o&&e.$set(m)},i(t){i||(e&&M(e.$$.fragment,t),i=!0)},o(t){e&&K(e.$$.fragment,t),i=!1},d(t){a[11](null),t&&G(n),e&&ee(e,t)}}}function Wt(a){let e,n,i;var o=a[1][1];function h(t){return{props:{data:t[4],form:t[2]}}}return o&&(e=Z(o,h(a)),a[10](e)),{c(){e&&Q(e.$$.fragment),n=H()},l(t){e&&De(e.$$.fragment,t),n=H()},m(t,u){e&&x(e,t,u),Y(t,n,u),i=!0},p(t,u){const m={};if(u&16&&(m.data=t[4]),u&4&&(m.form=t[2]),o!==(o=t[1][1])){if(e){_e();const p=e;K(p.$$.fragment,1,0,()=>{ee(p,1)}),ge()}o?(e=Z(o,h(t)),t[10](e),Q(e.$$.fragment),M(e.$$.fragment,1),x(e,n.parentNode,n)):e=null}else o&&e.$set(m)},i(t){i||(e&&M(e.$$.fragment,t),i=!0)},o(t){e&&K(e.$$.fragment,t),i=!1},d(t){a[10](null),t&&G(n),e&&ee(e,t)}}}function at(a){let e,n=a[6]&&rt(a);return{c(){e=bt("div"),n&&n.c(),this.h()},l(i){e=vt(i,"DIV",{id:!0,"aria-live":!0,"aria-atomic":!0,style:!0});var o=Et(e);n&&n.l(o),o.forEach(G),this.h()},h(){Re(e,"id","svelte-announcer"),Re(e,"aria-live","assertive"),Re(e,"aria-atomic","true"),B(e,"position","absolute"),B(e,"left","0"),B(e,"top","0"),B(e,"clip","rect(0 0 0 0)"),B(e,"clip-path","inset(50%)"),B(e,"overflow","hidden"),B(e,"white-space","nowrap"),B(e,"width","1px"),B(e,"height","1px")},m(i,o){Y(i,e,o),n&&n.m(e,null)},p(i,o){i[6]?n?n.p(i,o):(n=rt(i),n.c(),n.m(e,null)):n&&(n.d(1),n=null)},d(i){i&&G(e),n&&n.d()}}}function rt(a){let e;return{c(){e=kt(a[7])},l(n){e=St(n,a[7])},m(n,i){Y(n,e,i)},p(n,i){i&128&&Rt(e,n[7])},d(n){n&&G(e)}}}function Yt(a){let e,n,i,o,h;const t=[zt,Jt],u=[];function m(g,y){return g[1][1]?0:1}e=m(a),n=u[e]=t[e](a);let p=a[5]&&at(a);return{c(){n.c(),i=gt(),p&&p.c(),o=H()},l(g){n.l(g),i=wt(g),p&&p.l(g),o=H()},m(g,y){u[e].m(g,y),Y(g,i,y),p&&p.m(g,y),Y(g,o,y),h=!0},p(g,[y]){let N=e;e=m(g),e===N?u[e].p(g,y):(_e(),K(u[N],1,1,()=>{u[N]=null}),ge(),n=u[e],n?n.p(g,y):(n=u[e]=t[e](g),n.c()),M(n,1),n.m(i.parentNode,i)),g[5]?p?p.p(g,y):(p=at(g),p.c(),p.m(o.parentNode,o)):p&&(p.d(1),p=null)},i(g){h||(M(n),h=!0)},o(g){K(n),h=!1},d(g){u[e].d(g),g&&G(i),p&&p.d(g),g&&G(o)}}}function Xt(a,e,n){let{stores:i}=e,{page:o}=e,{constructors:h}=e,{components:t=[]}=e,{form:u}=e,{data_0:m=null}=e,{data_1:p=null}=e;yt(i.page.notify);let g=!1,y=!1,N=null;je(()=>{const S=i.page.subscribe(()=>{g&&(n(6,y=!0),n(7,N=document.title||"untitled page"))});return n(5,g=!0),S});function D(S){Le[S?"unshift":"push"](()=>{t[1]=S,n(0,t)})}function C(S){Le[S?"unshift":"push"](()=>{t[0]=S,n(0,t)})}function L(S){Le[S?"unshift":"push"](()=>{t[0]=S,n(0,t)})}return a.$$set=S=>{"stores"in S&&n(8,i=S.stores),"page"in S&&n(9,o=S.page),"constructors"in S&&n(1,h=S.constructors),"components"in S&&n(0,t=S.components),"form"in S&&n(2,u=S.form),"data_0"in S&&n(3,m=S.data_0),"data_1"in S&&n(4,p=S.data_1)},a.$$.update=()=>{a.$$.dirty&768&&i.page.set(o)},[t,h,u,m,p,g,y,N,i,o,D,C,L]}class Zt extends pt{constructor(e){super(),mt(this,e,Xt,Yt,_t,{stores:8,page:9,constructors:1,components:0,form:2,data_0:3,data_1:4})}}const Qt="modulepreload",xt=function(a,e){return new URL(a,e).href},st={},Pe=function(e,n,i){if(!n||n.length===0)return e();const o=document.getElementsByTagName("link");return Promise.all(n.map(h=>{if(h=xt(h,i),h in st)return;st[h]=!0;const t=h.endsWith(".css"),u=t?'[rel="stylesheet"]':"";if(!!i)for(let g=o.length-1;g>=0;g--){const y=o[g];if(y.href===h&&(!t||y.rel==="stylesheet"))return}else if(document.querySelector(`link[href="${h}"]${u}`))return;const p=document.createElement("link");if(p.rel=t?"stylesheet":Qt,t||(p.as="script",p.crossOrigin=""),p.href=h,document.head.appendChild(p),t)return new Promise((g,y)=>{p.addEventListener("load",g),p.addEventListener("error",()=>y(new Error(`Unable to preload CSS for ${h}`)))})})).then(()=>e())},en={},ye=[()=>Pe(()=>import("./chunks/0-9f97ed19.js"),["./chunks/0-9f97ed19.js","./components/pages/_layout.svelte-0f05afef.js","./chunks/index-867754b5.js","./chunks/SvelteToast.svelte_svelte_type_style_lang-d2ff673e.js","./assets/SvelteToast-8600cd0d.css","./assets/_layout-16d1691e.css"],import.meta.url),()=>Pe(()=>import("./chunks/1-b428053d.js"),["./chunks/1-b428053d.js","./components/error.svelte-be6e62e4.js","./chunks/index-867754b5.js","./chunks/singletons-3caa26a7.js"],import.meta.url),()=>Pe(()=>import("./chunks/2-4d14c2f6.js"),["./chunks/2-4d14c2f6.js","./components/pages/_page.svelte-e95ee34e.js","./chunks/index-867754b5.js","./chunks/SvelteToast.svelte_svelte_type_style_lang-d2ff673e.js","./assets/SvelteToast-8600cd0d.css","./assets/_page-b07ae230.css"],import.meta.url)],ut=[],tn={"/":[2]},nn={handleError:({error:a})=>{console.error(a)}};let ie=class{constructor(e,n){this.status=e,typeof n=="string"?this.body={message:n}:n?this.body=n:this.body={message:`Error: ${e}`}}toString(){return JSON.stringify(this.body)}},ot=class{constructor(e,n){this.status=e,this.location=n}};async function an(a){var e;for(const n in a)if(typeof((e=a[n])==null?void 0:e.then)=="function")return Object.fromEntries(await Promise.all(Object.entries(a).map(async([i,o])=>[i,await o])));return a}Object.getOwnPropertyNames(Object.prototype).sort().join("\0");const rn=-1,sn=-2,on=-3,ln=-4,cn=-5,fn=-6;function un(a){if(typeof a=="number")return i(a,!0);if(!Array.isArray(a)||a.length===0)throw new Error("Invalid input");const e=a,n=Array(e.length);function i(o,h=!1){if(o===rn)return;if(o===on)return NaN;if(o===ln)return 1/0;if(o===cn)return-1/0;if(o===fn)return-0;if(h)throw new Error("Invalid input");if(o in n)return n[o];const t=e[o];if(!t||typeof t!="object")n[o]=t;else if(Array.isArray(t))if(typeof t[0]=="string")switch(t[0]){case"Date":n[o]=new Date(t[1]);break;case"Set":const m=new Set;n[o]=m;for(let y=1;y<t.length;y+=1)m.add(i(t[y]));break;case"Map":const p=new Map;n[o]=p;for(let y=1;y<t.length;y+=2)p.set(i(t[y]),i(t[y+1]));break;case"RegExp":n[o]=new RegExp(t[1],t[2]);break;case"Object":n[o]=Object(t[1]);break;case"BigInt":n[o]=BigInt(t[1]);break;case"null":const g=Object.create(null);n[o]=g;for(let y=1;y<t.length;y+=2)g[t[y]]=i(t[y+1]);break}else{const u=new Array(t.length);n[o]=u;for(let m=0;m<t.length;m+=1){const p=t[m];p!==sn&&(u[m]=i(p))}}else{const u={};n[o]=u;for(const m in t){const p=t[m];u[m]=i(p)}}return n[o]}return i(0)}function dn(a){return a.filter(e=>e!=null)}const Oe=Gt(ye,ut,tn,en),dt=ye[0],Te=ye[1];dt();Te();const W=ft(lt)??{},ae=ft(ct)??{};function Ue(a){W[a]=se()}function hn({target:a}){var Ye;const e=document.documentElement,n=[],i=[];let o=null;const h={before_navigate:[],after_navigate:[]};let t={branch:[],error:null,url:null},u=!1,m=!1,p=!0,g=!1,y=!1,N=!1,D=!1,C,L=(Ye=history.state)==null?void 0:Ye[$];L||(L=Date.now(),history.replaceState({...history.state,[$]:L},"",location.href));const S=W[L];S&&(history.scrollRestoration="manual",scrollTo(S.x,S.y));let J,qe,le;async function Ve(){le=le||Promise.resolve(),await le,le=null;const r=new URL(location.href),s=ue(r,!0);o=null,await He(s,r,[])}function $e(r){i.some(s=>s==null?void 0:s.snapshot)&&(ae[r]=i.map(s=>{var c;return(c=s==null?void 0:s.snapshot)==null?void 0:c.capture()}))}function Be(r){var s;(s=ae[r])==null||s.forEach((c,l)=>{var d,f;(f=(d=i[l])==null?void 0:d.snapshot)==null||f.restore(c)})}async function be(r,{noScroll:s=!1,replaceState:c=!1,keepFocus:l=!1,state:d={},invalidateAll:f=!1},w,_){return typeof r=="string"&&(r=new URL(r,Qe(document))),he({url:r,scroll:s?se():null,keepfocus:l,redirect_chain:w,details:{state:d,replaceState:c},nav_token:_,accepted:()=>{f&&(D=!0)},blocked:()=>{},type:"goto"})}async function Fe(r){const s=ue(r,!1);if(!s)throw new Error(`Attempted to preload a URL that does not belong to this app: ${r}`);return o={id:s.id,promise:Ge(s).then(c=>(c.type==="loaded"&&c.state.error&&(o=null),c))},o.promise}async function ce(...r){const c=Oe.filter(l=>r.some(d=>l.exec(d))).map(l=>Promise.all([...l.layouts,l.leaf].map(d=>d==null?void 0:d[1]())));await Promise.all(c)}async function He(r,s,c,l,d,f={},w){var v,b;qe=f;let _=r&&await Ge(r);if(_||(_=await We(s,{id:null},await re(new Error(`Not found: ${s.pathname}`),{url:s,params:{},route:{id:null}}),404)),s=(r==null?void 0:r.url)||s,qe!==f)return!1;if(_.type==="redirect")if(c.length>10||c.includes(s.pathname))_=await fe({status:500,error:await re(new Error("Redirect loop"),{url:s,params:{},route:{id:null}}),url:s,route:{id:null}});else return be(new URL(_.location,s).href,{},[...c,s.pathname],f),!1;else((b=(v=_.props)==null?void 0:v.page)==null?void 0:b.status)>=400&&await F.updated.check()&&await ne(s);if(n.length=0,D=!1,g=!0,l&&(Ue(l),$e(l)),d&&d.details){const{details:k}=d,A=k.replaceState?0:1;if(k.state[$]=L+=A,history[k.replaceState?"replaceState":"pushState"](k.state,"",s),!k.replaceState){let O=L+1;for(;ae[O]||W[O];)delete ae[O],delete W[O],O+=1}}if(o=null,m?(t=_.state,_.props.page&&(_.props.page.url=s),C.$set(_.props)):Ke(_),d){const{scroll:k,keepfocus:A}=d,{activeElement:O}=document;await pe();const I=document.activeElement!==O&&document.activeElement!==document.body;if(!A&&!I&&await Ne(),p){const E=s.hash&&document.getElementById(decodeURIComponent(s.hash.slice(1)));k?scrollTo(k.x,k.y):E?E.scrollIntoView():scrollTo(0,0)}}else await pe();p=!0,_.props.page&&(J=_.props.page),w&&w(),g=!1}function Ke(r){var l;t=r.state;const s=document.querySelector("style[data-sveltekit]");s&&s.remove(),J=r.props.page,C=new Zt({target:a,props:{...r.props,stores:F,components:i},hydrate:!0}),Be(L);const c={from:null,to:{params:t.params,route:{id:((l=t.route)==null?void 0:l.id)??null},url:new URL(location.href)},willUnload:!1,type:"enter"};h.after_navigate.forEach(d=>d(c)),m=!0}async function te({url:r,params:s,branch:c,status:l,error:d,route:f,form:w}){let _="never";for(const I of c)(I==null?void 0:I.slash)!==void 0&&(_=I.slash);r.pathname=Pt(r.pathname,_),r.search=r.search;const v={type:"loaded",state:{url:r,params:s,branch:c,error:d,route:f},props:{constructors:dn(c).map(I=>I.node.component)}};w!==void 0&&(v.props.form=w);let b={},k=!J,A=0;for(let I=0;I<Math.max(c.length,t.branch.length);I+=1){const E=c[I],j=t.branch[I];(E==null?void 0:E.data)!==(j==null?void 0:j.data)&&(k=!0),E&&(b={...b,...E.data},k&&(v.props[`data_${A}`]=b),A+=1)}return(!t.url||r.href!==t.url.href||t.error!==d||w!==void 0&&w!==J.form||k)&&(v.props.page={error:d,params:s,route:{id:(f==null?void 0:f.id)??null},status:l,url:new URL(r),form:w??null,data:k?b:J.data}),v}async function ve({loader:r,parent:s,url:c,params:l,route:d,server_data_node:f}){var b,k,A;let w=null;const _={dependencies:new Set,params:new Set,parent:!1,route:!1,url:!1},v=await r();if((b=v.universal)!=null&&b.load){let O=function(...E){for(const j of E){const{href:q}=new URL(j,c);_.dependencies.add(q)}};const I={route:{get id(){return _.route=!0,d.id}},params:new Proxy(l,{get:(E,j)=>(_.params.add(j),E[j])}),data:(f==null?void 0:f.data)??null,url:jt(c,()=>{_.url=!0}),async fetch(E,j){let q;E instanceof Request?(q=E.url,j={body:E.method==="GET"||E.method==="HEAD"?void 0:await E.blob(),cache:E.cache,credentials:E.credentials,headers:E.headers,integrity:E.integrity,keepalive:E.keepalive,method:E.method,mode:E.mode,redirect:E.redirect,referrer:E.referrer,referrerPolicy:E.referrerPolicy,signal:E.signal,...j}):q=E;const V=new URL(q,c);return O(V.href),V.origin===c.origin&&(q=V.href.slice(c.origin.length)),m?$t(q,V.href,j):Vt(q,j)},setHeaders:()=>{},depends:O,parent(){return _.parent=!0,s()}};w=await v.universal.load.call(null,I)??null,w=w?await an(w):null}return{node:v,loader:r,server:f,universal:(k=v.universal)!=null&&k.load?{type:"data",data:w,uses:_}:null,data:w??(f==null?void 0:f.data)??null,slash:((A=v.universal)==null?void 0:A.trailingSlash)??(f==null?void 0:f.slash)}}function Me(r,s,c,l,d){if(D)return!0;if(!l)return!1;if(l.parent&&r||l.route&&s||l.url&&c)return!0;for(const f of l.params)if(d[f]!==t.params[f])return!0;for(const f of l.dependencies)if(n.some(w=>w(new URL(f))))return!0;return!1}function Ee(r,s){return(r==null?void 0:r.type)==="data"?{type:"data",data:r.data,uses:{dependencies:new Set(r.uses.dependencies??[]),params:new Set(r.uses.params??[]),parent:!!r.uses.parent,route:!!r.uses.route,url:!!r.uses.url},slash:r.slash}:(r==null?void 0:r.type)==="skip"?s??null:null}async function Ge({id:r,invalidating:s,url:c,params:l,route:d}){if((o==null?void 0:o.id)===r)return o.promise;const{errors:f,layouts:w,leaf:_}=d,v=[...w,_];f.forEach(R=>R==null?void 0:R().catch(()=>{})),v.forEach(R=>R==null?void 0:R[1]().catch(()=>{}));let b=null;const k=t.url?r!==t.url.pathname+t.url.search:!1,A=t.route?d.id!==t.route.id:!1;let O=!1;const I=v.map((R,T)=>{var z;const P=t.branch[T],U=!!(R!=null&&R[0])&&((P==null?void 0:P.loader)!==R[1]||Me(O,A,k,(z=P.server)==null?void 0:z.uses,l));return U&&(O=!0),U});if(I.some(Boolean)){try{b=await it(c,I)}catch(R){return fe({status:R instanceof ie?R.status:500,error:await re(R,{url:c,params:l,route:{id:d.id}}),url:c,route:d})}if(b.type==="redirect")return b}const E=b==null?void 0:b.nodes;let j=!1;const q=v.map(async(R,T)=>{var ke;if(!R)return;const P=t.branch[T],U=E==null?void 0:E[T];if((!U||U.type==="skip")&&R[1]===(P==null?void 0:P.loader)&&!Me(j,A,k,(ke=P.universal)==null?void 0:ke.uses,l))return P;if(j=!0,(U==null?void 0:U.type)==="error")throw U;return ve({loader:R[1],url:c,params:l,route:d,parent:async()=>{var Ze;const Xe={};for(let Se=0;Se<T;Se+=1)Object.assign(Xe,(Ze=await q[Se])==null?void 0:Ze.data);return Xe},server_data_node:Ee(U===void 0&&R[0]?{type:"skip"}:U??null,P==null?void 0:P.server)})});for(const R of q)R.catch(()=>{});const V=[];for(let R=0;R<v.length;R+=1)if(v[R])try{V.push(await q[R])}catch(T){if(T instanceof ot)return{type:"redirect",location:T.location};let P=500,U;if(E!=null&&E.includes(T))P=T.status??P,U=T.error;else if(T instanceof ie)P=T.status,U=T.body;else{if(await F.updated.check())return await ne(c);U=await re(T,{params:l,url:c,route:{id:d.id}})}const z=await Je(R,V,f);return z?await te({url:c,params:l,branch:V.slice(0,z.idx).concat(z.node),status:P,error:U,route:d}):await We(c,{id:d.id},U,P)}else V.push(void 0);return await te({url:c,params:l,branch:V,status:200,error:null,route:d,form:s?void 0:null})}async function Je(r,s,c){for(;r--;)if(c[r]){let l=r;for(;!s[l];)l-=1;try{return{idx:l+1,node:{node:await c[r](),loader:c[r],data:{},server:null,universal:null}}}catch{continue}}}async function fe({status:r,error:s,url:c,route:l}){const d={};let f=null;if(ut[0]===0)try{const b=await it(c,[!0]);if(b.type!=="data"||b.nodes[0]&&b.nodes[0].type!=="data")throw 0;f=b.nodes[0]??null}catch{(c.origin!==location.origin||c.pathname!==location.pathname||u)&&await ne(c)}const _=await ve({loader:dt,url:c,params:d,route:l,parent:()=>Promise.resolve({}),server_data_node:Ee(f)}),v={node:await Te(),loader:Te,universal:null,server:null,data:null};return await te({url:c,params:d,branch:[_,v],status:r,error:s,route:null})}function ue(r,s){if(et(r,X))return;const c=de(r);for(const l of Oe){const d=l.exec(c);if(d)return{id:r.pathname+r.search,invalidating:s,route:l,params:Ut(d),url:r}}}function de(r){return Ot(r.pathname.slice(X.length)||"/")}function ze({url:r,type:s,intent:c,delta:l}){var _,v;let d=!1;const f={from:{params:t.params,route:{id:((_=t.route)==null?void 0:_.id)??null},url:t.url},to:{params:(c==null?void 0:c.params)??null,route:{id:((v=c==null?void 0:c.route)==null?void 0:v.id)??null},url:r},willUnload:!c,type:s};l!==void 0&&(f.delta=l);const w={...f,cancel:()=>{d=!0}};return y||h.before_navigate.forEach(b=>b(w)),d?null:f}async function he({url:r,scroll:s,keepfocus:c,redirect_chain:l,details:d,type:f,delta:w,nav_token:_,accepted:v,blocked:b}){const k=ue(r,!1),A=ze({url:r,type:f,delta:w,intent:k});if(!A){b();return}const O=L;v(),y=!0,m&&F.navigating.set(A),await He(k,r,l,O,{scroll:s,keepfocus:c,details:d},_,()=>{y=!1,h.after_navigate.forEach(I=>I(A)),F.navigating.set(null)})}async function We(r,s,c,l){return r.origin===location.origin&&r.pathname===location.pathname&&!u?await fe({status:l,error:c,url:r,route:s}):await ne(r)}function ne(r){return location.href=r.href,new Promise(()=>{})}function ht(){let r;e.addEventListener("mousemove",f=>{const w=f.target;clearTimeout(r),r=setTimeout(()=>{l(w,2)},20)});function s(f){l(f.composedPath()[0],1)}e.addEventListener("mousedown",s),e.addEventListener("touchstart",s,{passive:!0});const c=new IntersectionObserver(f=>{for(const w of f)w.isIntersecting&&(ce(de(new URL(w.target.href))),c.unobserve(w.target))},{threshold:0});function l(f,w){const _=xe(f,e);if(!_)return;const{url:v,external:b}=Ie(_,X);if(b)return;const k=me(_);k.reload||(w<=k.preload_data?Fe(v):w<=k.preload_code&&ce(de(v)))}function d(){c.disconnect();for(const f of e.querySelectorAll("a")){const{url:w,external:_}=Ie(f,X);if(_)continue;const v=me(f);v.reload||(v.preload_code===tt.viewport&&c.observe(f),v.preload_code===tt.eager&&ce(de(w)))}}h.after_navigate.push(d),d()}return{after_navigate:r=>{je(()=>(h.after_navigate.push(r),()=>{const s=h.after_navigate.indexOf(r);h.after_navigate.splice(s,1)}))},before_navigate:r=>{je(()=>(h.before_navigate.push(r),()=>{const s=h.before_navigate.indexOf(r);h.before_navigate.splice(s,1)}))},disable_scroll_handling:()=>{(g||!m)&&(p=!1)},goto:(r,s={})=>be(r,s,[]),invalidate:r=>{if(typeof r=="function")n.push(r);else{const{href:s}=new URL(r,location.href);n.push(c=>c.href===s)}return Ve()},invalidateAll:()=>(D=!0,Ve()),preload_data:async r=>{const s=new URL(r,Qe(document));await Fe(s)},preload_code:ce,apply_action:async r=>{if(r.type==="error"){const s=new URL(location.href),{branch:c,route:l}=t;if(!l)return;const d=await Je(t.branch.length,c,l.errors);if(d){const f=await te({url:s,params:t.params,branch:c.slice(0,d.idx).concat(d.node),status:r.status??500,error:r.error,route:l});t=f.state,C.$set(f.props),pe().then(Ne)}}else if(r.type==="redirect")be(r.location,{invalidateAll:!0},[]);else{const s={form:r.data,page:{...J,form:r.data,status:r.status}};C.$set(s),r.type==="success"&&pe().then(Ne)}},_start_router:()=>{var r;history.scrollRestoration="manual",addEventListener("beforeunload",s=>{var l;let c=!1;if(!y){const d={from:{params:t.params,route:{id:((l=t.route)==null?void 0:l.id)??null},url:t.url},to:null,willUnload:!0,type:"leave",cancel:()=>c=!0};h.before_navigate.forEach(f=>f(d))}c?(s.preventDefault(),s.returnValue=""):history.scrollRestoration="auto"}),addEventListener("visibilitychange",()=>{document.visibilityState==="hidden"&&(Ue(L),nt(lt,W),$e(L),nt(ct,ae))}),(r=navigator.connection)!=null&&r.saveData||ht(),e.addEventListener("click",s=>{if(s.button||s.which!==1||s.metaKey||s.ctrlKey||s.shiftKey||s.altKey||s.defaultPrevented)return;const c=xe(s.composedPath()[0],e);if(!c)return;const{url:l,external:d,target:f}=Ie(c,X);if(!l)return;if(f==="_parent"||f==="_top"){if(window.parent!==window)return}else if(f&&f!=="_self")return;const w=me(c);if(!(c instanceof SVGAElement)&&l.protocol!==location.protocol&&!(l.protocol==="https:"||l.protocol==="http:"))return;if(d||w.reload){ze({url:l,type:"link"})||s.preventDefault(),y=!0;return}const[v,b]=l.href.split("#");if(b!==void 0&&v===location.href.split("#")[0]){N=!0,Ue(L),t.url=l,F.page.set({...J,url:l}),F.page.notify();return}he({url:l,scroll:w.noscroll?se():null,keepfocus:!1,redirect_chain:[],details:{state:{},replaceState:l.href===location.href},accepted:()=>s.preventDefault(),blocked:()=>s.preventDefault(),type:"link"})}),e.addEventListener("submit",s=>{if(s.defaultPrevented)return;const c=HTMLFormElement.prototype.cloneNode.call(s.target),l=s.submitter;if(((l==null?void 0:l.formMethod)||c.method)!=="get")return;const f=new URL((l==null?void 0:l.hasAttribute("formaction"))&&(l==null?void 0:l.formAction)||c.action);if(et(f,X))return;const w=s.target,{noscroll:_,reload:v}=me(w);if(v)return;s.preventDefault(),s.stopPropagation();const b=new FormData(w),k=l==null?void 0:l.getAttribute("name");k&&b.append(k,(l==null?void 0:l.getAttribute("value"))??""),f.search=new URLSearchParams(b).toString(),he({url:f,scroll:_?se():null,keepfocus:!1,redirect_chain:[],details:{state:{},replaceState:!1},nav_token:{},accepted:()=>{},blocked:()=>{},type:"form"})}),addEventListener("popstate",async s=>{var c;if((c=s.state)!=null&&c[$]){if(s.state[$]===L)return;const l=W[s.state[$]];if(t.url.href.split("#")[0]===location.href.split("#")[0]){W[L]=se(),L=s.state[$],scrollTo(l.x,l.y);return}const d=s.state[$]-L;let f=!1;await he({url:new URL(location.href),scroll:l,keepfocus:!1,redirect_chain:[],details:null,accepted:()=>{L=s.state[$]},blocked:()=>{history.go(-d),f=!0},type:"popstate",delta:d}),f||Be(L)}}),addEventListener("hashchange",()=>{N&&(N=!1,history.replaceState({...history.state,[$]:++L},"",location.href))});for(const s of document.querySelectorAll("link"))s.rel==="icon"&&(s.href=s.href);addEventListener("pageshow",s=>{s.persisted&&F.navigating.set(null)})},_hydrate:async({status:r=200,error:s,node_ids:c,params:l,route:d,data:f,form:w})=>{u=!0;const _=new URL(location.href);({params:l={},route:d={id:null}}=ue(_,!1)||{});let v;try{const b=c.map(async(k,A)=>{const O=f[A];return ve({loader:ye[k],url:_,params:l,route:d,parent:async()=>{const I={};for(let E=0;E<A;E+=1)Object.assign(I,(await b[E]).data);return I},server_data_node:Ee(O)})});v=await te({url:_,params:l,branch:await Promise.all(b),status:r,error:s,form:w,route:Oe.find(({id:k})=>k===d.id)??null})}catch(b){if(b instanceof ot){await ne(new URL(b.location,location.href));return}v=await fe({status:b instanceof ie?b.status:500,error:await re(b,{url:_,params:l,route:d}),url:_,route:d})}Ke(v)}}}async function it(a,e){var h;const n=new URL(a);n.pathname=Ct(a.pathname),n.searchParams.append("x-sveltekit-invalidated",e.map(t=>t?"1":"").join("_"));const i=await we(n.href),o=await i.json();if(!i.ok)throw new ie(i.status,o);return(h=o.nodes)==null||h.forEach(t=>{(t==null?void 0:t.type)==="data"&&(t.data=un(t.data),t.uses={dependencies:new Set(t.uses.dependencies??[]),params:new Set(t.uses.params??[]),parent:!!t.uses.parent,route:!!t.uses.route,url:!!t.uses.url})}),o}function re(a,e){return a instanceof ie?a.body:nn.handleError({error:a,event:e})??{message:e.route.id!=null?"Internal Error":"Not Found"}}function Ne(){const a=document.querySelector("[autofocus]");if(a)a.focus();else{const e=document.body,n=e.getAttribute("tabindex");return e.tabIndex=-1,e.focus({preventScroll:!0}),n!==null?e.setAttribute("tabindex",n):e.removeAttribute("tabindex"),new Promise(i=>{setTimeout(()=>{var o;i((o=getSelection())==null?void 0:o.removeAllRanges())})})}}async function wn({assets:a,env:e,hydrate:n,target:i,version:o}){It(a),At(o);const h=hn({target:i});Lt({client:h}),n?await h._hydrate(n):h.goto(location.href,{replaceState:!0}),h._start_router()}export{wn as start};

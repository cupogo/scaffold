import{aa as i,C as p}from"./index-867754b5.js";const _=p(!1),h=p(null),u=p([]),m=p(null);function g(r=1){var n,f;const a=i(u).length,s=i(u)[a-1];return i(h)||(n=s==null?void 0:s.callbacks)!=null&&n.onBeforeClose&&((f=s==null?void 0:s.callbacks)==null?void 0:f.onBeforeClose())===!1?!1:(i(_)&&a>0&&h.set(!0),_.set(!1),m.set("pop"),b(r),!0)}function k(){return g(1)}function I(r,a,s){i(h)||(m.set("push"),i(_)&&i(u).length&&h.set(!0),_.set(!1),s!=null&&s.replace?u.update(n=>[...n.slice(0,n.length-1),{component:r,props:a}]):u.update(n=>[...n,{component:r,props:a}]))}function b(r=1){u.update(a=>a.slice(0,Math.max(0,a.length-r)))}const x={duration:4e3,initial:1,next:0,pausable:!1,dismissable:!0,reversed:!1,intro:{x:256}},v=()=>{const{subscribe:r,update:a}=p([]);let s=0;const n={},f=e=>e instanceof Object;return{subscribe:r,push:(e,l={})=>{const o={target:"default",...f(e)?e:{...l,msg:e}},t=n[o.target]||{},c={...x,...t,...o,theme:{...t.theme,...o.theme},classes:[...t.classes||[],...o.classes||[]],id:++s};return a(d=>c.reversed?[...d,c]:[c,...d]),s},pop:e=>{a(l=>{if(!l.length||e===0)return[];if(f(e))return l.filter(t=>e(t));const o=e||Math.max(...l.map(t=>t.id));return l.filter(t=>t.id!==o)})},set:(e,l={})=>{const o=f(e)?{...e}:{...l,id:e};a(t=>{const c=t.findIndex(d=>d.id===o.id);return c>-1&&(t[c]={...t[c],...o}),t})},_init:(e="default",l={})=>(n[e]=l,n)}},M=v();export{M as a,k as c,_ as e,u as m,I as o,h as t};
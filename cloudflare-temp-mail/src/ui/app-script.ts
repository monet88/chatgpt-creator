export const uiScript = `
const S = {email:'',user:'',domain:'',domains:[],messages:[],refreshTimer:null,page:1,pageSize:20,selectedId:'',openRequest:0};
const $=id=>document.getElementById(id);
const LS_KEY='tempmail_email';

const toast=msg=>{const t=$('toast');t.textContent=msg;t.classList.add('show');setTimeout(()=>t.classList.remove('show'),2000)};
const showError=msg=>{const e=$('inline-error');e.textContent=msg;e.classList.add('show');setTimeout(()=>e.classList.remove('show'),3000)};
const copyText=(text,done='Đã sao chép!')=>{
  if(!text){toast('Không có gì để sao chép');return;}
  navigator.clipboard.writeText(text).then(()=>toast(done)).catch(()=>{
    const i=document.createElement('input');i.value=text;
    document.body.appendChild(i);i.select();document.execCommand('copy');
    document.body.removeChild(i);toast(done);
  });
};
const fmtDate=d=>new Date(d).toLocaleString('vi-VN',{day:'2-digit',month:'2-digit',year:'numeric',hour:'2-digit',minute:'2-digit'});

const api=async(method,path,body=null)=>{
  const opts={method,headers:{'content-type':'application/json'}};
  if(body)opts.body=JSON.stringify(body);
  const r=await fetch('/api/v1'+path,opts);
  const d=await r.json();
  if(!d.success)throw new Error(d.error?.message||'Lỗi không xác định');
  return d.data;
};

const setEmail=(email,user,domain)=>{
  S.email=email;S.user=user;S.domain=domain;
  $('current-email').textContent=email||'Chưa có địa chỉ';
  $('email-badge').textContent=email||'';
  $('username-input').value=user||'';
  const urlEl=$('url-email');
  if(email){
    const url=location.origin+'/#'+encodeURIComponent(email);
    urlEl.textContent=url;urlEl.href=url;urlEl.title='Bấm để sao chép link hộp thư';
    history.replaceState(null,'','/#'+encodeURIComponent(email));
    localStorage.setItem(LS_KEY,JSON.stringify({email,user,domain}));
  }else{
    urlEl.textContent='';urlEl.removeAttribute('href');history.replaceState(null,'','/');
    localStorage.removeItem(LS_KEY);
  }
};

const clearMessageDetail=()=>{
  S.selectedId='';
  $('message-detail').style.display='none';
  $('detail-body').replaceChildren();
};

const renderMessages=msgs=>{
  S.messages=msgs;
  $('inbox-count').textContent=msgs.length;
  $('unread-count').textContent=msgs.length;
  const empty=$('empty-state'),tableWrap=$('message-table-wrap'),pager=$('message-pager');
  if(!msgs.length){empty.style.display='flex';tableWrap.style.display='none';pager.style.display='none';clearMessageDetail();return;}
  if(S.selectedId&&!msgs.some(m=>m.id===S.selectedId))clearMessageDetail();
  empty.style.display='none';tableWrap.style.display='block';pager.style.display='flex';
  const totalPages=Math.max(1,Math.ceil(msgs.length/S.pageSize));
  if(S.page>totalPages)S.page=totalPages;
  const start=(S.page-1)*S.pageSize;
  const pageItems=msgs.slice(start,start+S.pageSize);
  $('pager-status').textContent='Hiển thị '+(start+1)+'-'+(start+pageItems.length)+' / '+msgs.length+' email';
  $('pager-page').textContent=S.page;
  $('prev-page').disabled=S.page<=1;
  $('next-page').disabled=S.page>=totalPages;
  $('inbox-body').replaceChildren(...pageItems.map(m=>{
    const tr=document.createElement('tr');
    if(m.id===S.selectedId)tr.className='selected';
    tr.innerHTML='<td class="from-cell"></td><td class="subject-cell"></td><td class="date-cell"></td><td class="actions-cell"></td>';
    tr.children[0].textContent=m.from;
    tr.children[1].textContent=m.subject||'(không có tiêu đề)';
    tr.children[2].textContent=fmtDate(m.receivedAt);
    const read=document.createElement('button');
    read.className='btn-read';read.textContent='View';
    read.onclick=()=>openMessage(m.id);
    const del=document.createElement('button');
    del.className='btn-del';del.textContent='Delete';
    del.onclick=()=>deleteOne(m.id);
    tr.children[3].append(read,del);
    return tr;
  }));
};

const loadOtp=async()=>{
  if(!S.user||!S.domain)return;
  try{
    const d=await api('GET','/email/'+S.domain+'/'+S.user+'/otp');
    const code=d?.otp||null;
    $('otp-section').style.display=code?'flex':'none';
    if(code)$('otp-code').textContent=code;
  }catch{}
};

const loadMessages=async()=>{
  if(!S.user||!S.domain)return;
  try{
    const d=await api('GET','/email/'+S.domain+'/'+S.user+'/messages');
    renderMessages(d.messages||[]);
    await loadOtp();
  }catch(e){showError(e.message)}
};

const openMessage=async id=>{
  const requestId=++S.openRequest;
  const mailboxKey=S.domain+'/'+S.user;
  try{
    const d=await api('GET','/email/'+S.domain+'/'+S.user+'/messages/'+id);
    if(requestId!==S.openRequest||mailboxKey!==S.domain+'/'+S.user)return;
    S.selectedId=id;
    $('detail-from').textContent=(d.from||'');
    $('detail-subject').textContent=d.subject||'(không có tiêu đề)';
    $('detail-date').textContent=fmtDate(d.receivedAt||Date.now());
    const bodyEl=$('detail-body');
    if(d.html){
      const iframe=document.createElement('iframe');
      iframe.className='detail-iframe';
      iframe.srcdoc=d.html;
      iframe.setAttribute('sandbox','');
      bodyEl.replaceChildren(iframe);
      bodyEl.dataset.type='html';
    }else{
      const pre=document.createElement('pre');
      pre.className='detail-text';
      pre.textContent=d.text||'(không có nội dung)';
      bodyEl.replaceChildren(pre);
      bodyEl.dataset.type='text';
    }
    $('message-detail').style.display='block';
    renderMessages(S.messages);
    $('message-detail').scrollIntoView({behavior:'smooth',block:'nearest'});
  }catch(e){showError(e.message)}
};

const deleteOne=async id=>{
  try{
    await api('DELETE','/email/'+S.domain+'/'+S.user+'/messages/'+id);
    if(S.selectedId===id)clearMessageDetail();
    await loadMessages();toast('Đã xóa');
  }catch(e){showError(e.message)}
};

const generate=async()=>{
  const user=$('username-input').value.trim();
  const domain=$('domain-select').value;
  if(!domain){showError('Chưa có domain');return;}
  try{
    const d=await api('POST','/email/generate',user?{domain,user}:{domain});
    setEmail(d.email,d.user,d.domain);
    S.messages=[];S.page=1;renderMessages([]);
    toast('Đã tạo: '+d.email);
  }catch(e){showError(e.message)}
};

const startRefresh=()=>{
  stopRefresh();
  if($('auto-refresh-toggle').checked)S.refreshTimer=setInterval(loadMessages,5000);
};
const stopRefresh=()=>{if(S.refreshTimer){clearInterval(S.refreshTimer);S.refreshTimer=null;}};

const init=async()=>{
  try{
    const d=await api('GET','/domains');
    const domains=d.domains||[];
    S.domains=domains;
    $('domain-count').textContent=domains.length;
    const sel=$('domain-select');
    sel.replaceChildren(...domains.map(dom=>{
      const o=document.createElement('option');
      o.value=dom;o.textContent=dom;return o;
    }));
    const hashEmail=location.hash?decodeURIComponent(location.hash.slice(1)):'';
    const saved=localStorage.getItem(LS_KEY);
    let restored=null;
    if(hashEmail&&hashEmail.includes('@')){
      const[user,domain]=hashEmail.split('@');
      if(domains.includes(domain))restored={email:hashEmail,user,domain};
    }
    if(!restored&&saved){
      try{const p=JSON.parse(saved);if(domains.includes(p.domain))restored=p;}catch{}
    }
    if(restored){
      setEmail(restored.email,restored.user,restored.domain);
      sel.value=restored.domain;
      await loadMessages();
    }
    startRefresh();
  }catch(e){showError('Không thể tải danh sách domain: '+e.message)}
};

$('generate-btn').onclick=generate;
$('copy-btn').onclick=()=>copyText(S.email,'Đã sao chép địa chỉ!');
$('url-email').onclick=e=>{e.preventDefault();copyText($('url-email').textContent,'Đã sao chép link hộp thư!')};
$('page-size').onchange=e=>{S.pageSize=Number(e.target.value);S.page=1;renderMessages(S.messages)};
$('prev-page').onclick=()=>{S.page=Math.max(1,S.page-1);renderMessages(S.messages)};
$('next-page').onclick=()=>{S.page+=1;renderMessages(S.messages)};
$('delete-all-btn').onclick=async()=>{
  if(!S.email||!confirm('Xóa tất cả email trong hộp thư này?'))return;
  try{
    await api('DELETE','/email/'+S.domain+'/'+S.user);
    renderMessages([]);setEmail('','','');
    localStorage.removeItem(LS_KEY);toast('Đã xóa hộp thư');
  }catch(e){showError(e.message)}
};
$('auto-refresh-toggle').onchange=()=>$('auto-refresh-toggle').checked?startRefresh():stopRefresh();
$('copy-otp-btn').onclick=()=>{
  const code=$('otp-code').textContent;
  if(!code||code==='—')return;
  copyText(code,'Đã sao chép mã OTP!');
};

init();
`;

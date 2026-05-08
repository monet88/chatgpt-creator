export const uiScript = `
const S = {email:'',user:'',domain:'',domains:[],messages:[],refreshTimer:null};
const $=id=>document.getElementById(id);
const LS_KEY='tempmail_email';

const toast=msg=>{const t=$('toast');t.textContent=msg;t.classList.add('show');setTimeout(()=>t.classList.remove('show'),2000)};
const showError=msg=>{const e=$('inline-error');e.textContent=msg;e.classList.add('show');setTimeout(()=>e.classList.remove('show'),3000)};

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
    urlEl.textContent=url;urlEl.href=url;
    history.replaceState(null,'','/#'+encodeURIComponent(email));
    localStorage.setItem(LS_KEY,JSON.stringify({email,user,domain}));
  }else{
    urlEl.textContent='';history.replaceState(null,'','/');
    localStorage.removeItem(LS_KEY);
  }
};

const renderMessages=msgs=>{
  S.messages=msgs;
  $('inbox-count').textContent=msgs.length;
  $('unread-count').textContent=msgs.length;
  const empty=$('empty-state'),table=$('message-table');
  if(!msgs.length){empty.style.display='flex';table.style.display='none';return;}
  empty.style.display='none';table.style.display='table';
  const fmt=d=>new Date(d).toLocaleString('vi-VN',{day:'2-digit',month:'2-digit',hour:'2-digit',minute:'2-digit'});
  $('inbox-body').replaceChildren(...msgs.map(m=>{
    const tr=document.createElement('tr');
    tr.innerHTML='<td class="from-cell"></td><td class="subject-cell"></td><td class="date-cell"></td><td></td>';
    tr.children[0].textContent=m.from;
    tr.children[1].textContent=m.subject||'(không có tiêu đề)';
    tr.children[2].textContent=fmt(m.receivedAt);
    const read=document.createElement('button');
    read.className='btn-read';read.textContent='Đọc';
    read.onclick=()=>openMessage(m.id);
    const del=document.createElement('button');
    del.className='btn-del';del.textContent='✕';
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
  try{
    const d=await api('GET','/email/'+S.domain+'/'+S.user+'/messages/'+id);
    $('modal-from').textContent='Từ: '+(d.from||'');
    $('modal-subject').textContent=d.subject||'(không có tiêu đề)';
    $('modal-body').textContent=d.text||d.html||'(không có nội dung)';
    $('msg-modal').showModal();
  }catch(e){showError(e.message)}
};

const deleteOne=async id=>{
  try{
    await api('DELETE','/email/'+S.domain+'/'+S.user+'/messages/'+id);
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
    S.messages=[];renderMessages([]);
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
    // Restore from URL hash first, then localStorage
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
$('copy-btn').onclick=()=>{
  if(!S.email){toast('Chưa có địa chỉ email');return;}
  navigator.clipboard.writeText(S.email).then(()=>toast('Đã sao chép!')).catch(()=>{
    const i=document.createElement('input');i.value=S.email;
    document.body.appendChild(i);i.select();document.execCommand('copy');
    document.body.removeChild(i);toast('Đã sao chép!');
  });
};
$('delete-all-btn').onclick=async()=>{
  if(!S.email||!confirm('Xóa tất cả email trong hộp thư này?'))return;
  try{
    await api('DELETE','/email/'+S.domain+'/'+S.user);
    renderMessages([]);setEmail('','','');
    localStorage.removeItem(LS_KEY);toast('Đã xóa hộp thư');
  }catch(e){showError(e.message)}
};
$('auto-refresh-toggle').onchange=()=>$('auto-refresh-toggle').checked?startRefresh():stopRefresh();
$('close-modal').onclick=()=>$('msg-modal').close();
$('msg-modal').onclick=e=>{if(e.target===$('msg-modal'))$('msg-modal').close()};
$('copy-otp-btn').onclick=()=>{
  const code=$('otp-code').textContent;
  if(!code||code==='—')return;
  navigator.clipboard.writeText(code).then(()=>toast('Đã sao chép mã OTP!')).catch(()=>toast(code));
};

init();
`;

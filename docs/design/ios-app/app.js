const icons = {
  leaf: '<path d="M20 4c-9-2-17 2-15 10 2 7 12 6 14-1 1-3 1-6 1-9Z"/><path d="M4 21c1-6 5-10 11-13"/>',
  home: '<path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1h-5v-7H9v7H4a1 1 0 0 1-1-1Z"/>',
  user: '<circle cx="12" cy="8" r="4"/><path d="M4 21v-2a8 8 0 0 1 16 0v2"/>',
  arrow: '<path d="m9 5 7 7-7 7"/>', back: '<path d="m15 5-7 7 7 7"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  file: '<path d="M14 3H5v18h14V8Z"/><path d="M14 3v5h5M8 12h8M8 16h5"/>',
  send: '<path d="m21 3-7 18-4-7-7-4Z"/><path d="m10 14 11-11"/>',
  check: '<path d="m5 12 4 4L19 6"/>',
  chat: '<path d="M21 11a9 9 0 0 1-9 9H4l-2 2 1-7a9 9 0 1 1 18-4Z"/><path d="M7 10h10M7 14h6"/>',
  lock: '<rect x="5" y="10" width="14" height="11" rx="3"/><path d="M8 10V7a4 4 0 0 1 8 0v3M12 14v3"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2M5 5l1 1M18 18l1 1M5 19l1-1M18 6l1-1"/>',
  out: '<path d="M10 4H4v16h6M10 12h11m-4-4 4 4-4 4"/>',
  eye: '<path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7S2 12 2 12Z"/><circle cx="12" cy="12" r="3"/>',
  shield: '<path d="m12 3 8 3v6c0 5-8 9-8 9s-8-4-8-9V6Z"/><path d="m8 12 3 3 5-6"/>',
  edit: '<path d="m15 4 5 5M4 20l5-1L21 7l-5-5L4 14Z"/>'
};
const icon = (name, cls = '') => `<svg class="icon ${cls}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${icons[name] || icons.file}</svg>`;
const pages = {login: '登录', register: '注册', home: '首页', profile: '我的', info: '个人信息'};
const params = new URLSearchParams(location.search);
let current = Object.hasOwn(pages, params.get('page')) ? params.get('page') : 'home';
let theme = params.get('theme') === 'native' ? 'native' : 'glass';
const screen = document.querySelector('#screen');
const device = document.querySelector('.device');
const overview = params.get('overview') === '1';
let toastTimer;

function avatar(large = false) {
  return `<div class="avatar ${large ? 'avatar-large' : ''}" aria-label="演示头像">${icon('user')}</div>`;
}
function header(title, back = false, trailing = '') {
  return `<header class="app-header">${back ? `<button class="back-button" data-page="profile" aria-label="返回我的">${icon('back')}<span>我的</span></button>` : '<span></span>'}<span class="nav-title">${title}</span>${trailing || '<span></span>'}</header>`;
}
function tabs(active) {
  return `<nav class="tab-bar" aria-label="App 主导航"><button data-page="home" ${active === 'home' ? 'aria-current="page"' : ''}>${icon('home')}<span>首页</span></button><button data-page="profile" ${active === 'profile' ? 'aria-current="page"' : ''}>${icon('user')}<span>我的</span></button></nav>`;
}
function auth(register) {
  return `<section class="auth-page">
    <div class="auth-top"><span class="mini-brand">leave / 校园请假</span><span class="auth-step">${register ? '创建账号' : '欢迎回来'}</span></div>
    <div class="auth-intro"><div class="brand-mark">${icon('leaf')}</div><p class="eyebrow">${register ? 'HELLO, NEW FRIEND' : 'A LITTLE ROOM FOR LIFE'}</p><h1>${register ? '从这里，开始。' : '给生活，<br>留一点空间。'}</h1><p>${register ? '创建你的账号，让校园请假更轻松。' : '登录校园账号，安心安排每一次离开。'}</p></div>
    <form id="auth-form" novalidate>
      <div class="form-group"><label for="student">学号</label><input id="student" name="student" inputmode="numeric" maxlength="12" autocomplete="off" placeholder="请输入 12 位学号" aria-describedby="form-error"></div>
      <div class="form-group"><label for="password">密码</label><div class="password-wrap"><input id="password" name="password" type="password" maxlength="30" autocomplete="off" placeholder="${register ? '设置 6–30 位密码' : '请输入密码'}" aria-describedby="form-error"><button type="button" class="eye-button" data-eye="password" aria-label="显示密码" aria-pressed="false">${icon('eye')}</button></div></div>
      ${register ? '<div class="form-group"><label for="confirm">确认密码</label><div class="password-wrap"><input id="confirm" name="confirm" type="password" maxlength="30" autocomplete="off" placeholder="请再次输入密码" aria-describedby="form-error"><button type="button" class="eye-button" data-eye="confirm" aria-label="显示确认密码" aria-pressed="false">' + icon('eye') + '</button></div></div>' : ''}
      <p id="form-error" class="form-error" role="alert"></p>
      <button class="primary-button" type="submit">${register ? '创建账号' : '登录'}${icon('arrow')}</button>
    </form>
    <p class="auth-switch">${register ? '已经有账号？' : '还没有账号？'}<button class="text-button" data-page="${register ? 'login' : 'register'}">${register ? '立即登录' : '创建账号'}</button></p>
    <div class="auth-footer">${icon('shield')}<span>校园身份，让每一次申请都有归属</span></div>
  </section>`;
}
function home() {
  return `<section class="tab-page home-page"><div class="page-content">
    <div class="home-top"><span class="mini-brand">leave<span class="brand-dot"></span></span><button data-page="profile" class="avatar-button" aria-label="打开我的">${avatar()}</button></div>
    <div class="greeting"><p class="eyebrow">MONDAY, OCTOBER 19</p><h1>早上好，林同学<span class="greeting-dot">。</span></h1><p>今天，也要好好照顾自己。</p></div>
    <div class="hero-card"><div class="hero-topline"><span class="capsule">校园生活 · 从容安排</span>${icon('leaf')}</div><div class="hero-art" aria-hidden="true"><div class="orbit orbit-one"></div><div class="orbit orbit-two"></div><div class="art-ticket">${icon('leaf')}<span>LEAVE PASS</span><i></i><b>给自己一点时间</b></div><span class="art-spark">✦</span></div><div class="hero-copy"><p class="eyebrow">TAKE YOUR TIME</p><h2>每次离开，<br>都安心一点。</h2><p>简单申请，安心出发。</p></div><button class="primary-button" data-demo="发起请假">${icon('plus')}发起请假${icon('arrow')}</button></div>
    <div class="section-heading"><h2>只需三步</h2><span>让流程更简单</span></div>
    <div class="flow-card">${[['file','填写申请'],['send','提交申请'],['check','审批通过']].map(([i,t],n) => `<div class="flow-step"><span class="flow-icon">${icon(i)}</span><span>${t}</span><small>0${n + 1}</small></div>${n < 2 ? '<span class="flow-line"></span>' : ''}`).join('')}</div>
    <p class="home-footnote">每一次出发，都期待你平安归来。</p>
    </div>${tabs('home')}</section>`;
}
const services = [['user','个人信息','查看与修改个人资料','info'],['file','请假记录','每一次申请，都有记录',''],['chat','用户反馈','让我们听到你的想法',''],['lock','修改密码','守护你的账号安全',''],['sun','显示与触感','选择更舒适的使用体验','']];
function profile() {
  return `<section class="tab-page profile-page"><div class="page-content"><div class="page-title"><p class="eyebrow">YOUR CAMPUS, YOUR SPACE</p><h1>我的</h1><p>关于你的一切，都在这里。</p></div>
    <button class="identity-card" data-page="info">${avatar()}<span class="identity-copy"><strong>林予安</strong><span>202401020318</span><small>软件工程 2403 班</small></span>${icon('arrow')}</button>
    <div class="section-heading"><h2>我的服务</h2><span>为日常，多一点便利</span></div><div class="service-list">${services.map(([i,t,d,p]) => `<button class="service-row" ${p ? `data-page="${p}"` : `data-demo="${t}"`}><span class="service-icon service-${i}">${icon(i)}</span><span class="service-copy"><strong>${t}</strong><small>${d}</small></span>${icon('arrow')}</button>`).join('')}</div>
    <button class="logout-button" data-page="login">${icon('out')}退出登录</button><p class="version-note">LEAVE · 让校园生活轻一点</p></div>${tabs('profile')}</section>`;
}
function info() {
  const groups = [['学生信息',[['学号','202401020318'],['班级','软件工程 2403 班'],['姓名','林予安'],['手机号','138 0000 2608'],['性别','女']]],['其他信息',[['家长姓名','林女士'],['家长电话','139 0000 2618'],['公寓名称','知行公寓'],['宿舍号','A 栋 506'],['导员姓名','陈老师']]]];
  return `<section class="info-page">${header('个人信息', true, '<button class="text-button" data-demo="修改个人信息">编辑</button>')}<div class="info-content"><div class="info-identity"><button class="avatar-button" data-demo="头像" aria-label="查看头像">${avatar(true)}</button><h1>林予安</h1><p>今天也很高兴见到你。</p></div>${groups.map(([t, rows]) => `<section class="info-section"><h2>${t}</h2><dl class="info-group">${rows.map(([k,v]) => `<div class="info-row"><dt>${k}</dt><dd>${v}</dd></div>`).join('')}</dl></section>`).join('')}<button class="secondary-button" data-demo="修改个人信息">${icon('edit')}修改个人信息</button><p class="privacy-note">${icon('shield')}个人信息仅用于校园请假服务</p></div></section>`;
}
function markup(page) {
  return page === 'login' || page === 'register' ? auth(page === 'register') : page === 'home' ? home() : page === 'profile' ? profile() : info();
}
function render() {
  screen.innerHTML = markup(current);
  screen.scrollTop = 0;
  document.querySelector('.page-picker').innerHTML = Object.entries(pages).map(([key,title]) => `<button data-page="${key}" ${key === current ? 'aria-current="page"' : ''}>${title}</button>`).join('');
  document.title = `${pages[current]} · Leave 设计预览`;
}
function applyTheme() {
  device.dataset.theme = theme;
  document.body.dataset.theme = theme;
  document.querySelectorAll('[data-theme].design-switch, .design-switch button').forEach(button => button.setAttribute('aria-pressed', String(button.dataset.theme === theme)));
  document.querySelector('#variant-label').textContent = theme === 'native' ? 'A / 原生简洁' : 'B / 轻盈玻璃 · 精修版';
}
function toast(message) {
  const el = document.querySelector('#toast');
  el.textContent = message;
  el.classList.add('visible');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => el.classList.remove('visible'), 2600);
}
document.addEventListener('click', event => {
  const button = event.target.closest('button');
  if (!button) return;
  if (button.dataset.page) { current = button.dataset.page; render(); }
  if (button.dataset.theme) { theme = button.dataset.theme; applyTheme(); }
  if (button.dataset.demo) toast(`${button.dataset.demo} · 功能将在后续页面中展示`);
  if (button.dataset.eye) {
    const input = document.getElementById(button.dataset.eye);
    const visible = input.type === 'password';
    input.type = visible ? 'text' : 'password';
    button.setAttribute('aria-pressed', String(visible));
    button.setAttribute('aria-label', `${visible ? '隐藏' : '显示'}${button.dataset.eye === 'confirm' ? '确认' : ''}密码`);
  }
});
document.addEventListener('submit', event => {
  if (event.target.id !== 'auth-form') return;
  event.preventDefault();
  const form = event.target;
  const student = form.elements.student;
  const password = form.elements.password;
  const confirm = form.elements.confirm;
  const error = document.querySelector('#form-error');
  form.querySelectorAll('input').forEach(input => input.removeAttribute('aria-invalid'));
  let invalid = null;
  let message = '';
  if (!/^\d{12}$/.test(student.value.trim())) { invalid = student; message = '请输入 12 位数字学号'; }
  else if (!password.value.trim()) { invalid = password; message = '请输入密码'; }
  else if (current === 'register' && (password.value.length < 6 || password.value.length > 30)) { invalid = password; message = '密码长度应为 6 至 30 位'; }
  else if (current === 'register' && password.value !== confirm.value) { invalid = confirm; message = '两次输入的密码不一致'; }
  if (invalid) { error.textContent = message; invalid.setAttribute('aria-invalid', 'true'); invalid.focus(); return; }
  form.reset();
  current = 'home';
  render();
  toast('已进入演示首页，欢迎体验');
});
if (overview) {
  document.body.classList.add('overview');
  document.body.innerHTML = `<header class="overview-heading"><span>leave / DESIGN STUDY</span><h1>${theme === 'native' ? 'A · 原生简洁' : 'B · 轻盈玻璃'}</h1><p>${theme === 'native' ? '用熟悉的秩序，让日常更从容。' : '让光与层次，为日常留一点呼吸。'}</p><small>五屏视觉总览 / 相同内容 · 不同表达 / 演示数据</small></header><div class="overview-grid">${Object.keys(pages).map(p => `<section class="overview-item"><p>${pages[p]}</p><div class="device" data-theme="${theme}"><div class="status-bar" aria-hidden="true"><span>9:41</span><span>••• ▰</span></div><div class="screen">${markup(p)}</div><div class="home-indicator"></div></div></section>`).join('')}</div>`;
  document.body.dataset.theme = theme;
} else { render(); applyTheme(); }

const paths = {
  home: '<path d="m3 10 9-7 9 7v10H15v-7H9v7H3Z"/>',
  records:
    '<rect x="5" y="3" width="14" height="18" rx="3"/><path d="M9 8h6M9 12h6M9 16h3"/>',
  user: '<circle cx="12" cy="8" r="4"/><path d="M4 21v-2a8 8 0 0 1 16 0v2"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  arrow: '<path d="m9 5 7 7-7 7"/>',
  back: '<path d="m15 5-7 7 7 7"/>',
  close: '<path d="m6 6 12 12M6 18 18 6"/>',
  check: '<path d="m5 12 4 4L19 6"/>',
  clock: '<circle cx="12" cy="12" r="9"/><path d="M12 6v6l4 2"/>',
  calendar:
    '<rect x="3" y="5" width="18" height="16" rx="3"/><path d="M7 3v4M17 3v4M3 11h18M8 15h2M14 15h2"/>',
  location:
    '<path d="M19 10c0 6-7 11-7 11S5 16 5 10a7 7 0 1 1 14 0Z"/><circle cx="12" cy="10" r="2"/>',
  lock: '<rect x="5" y="10" width="14" height="11" rx="3"/><path d="M8 10V7a4 4 0 0 1 8 0v3M12 14v3"/>',
  shield:
    '<path d="m12 3 8 3v6c0 5-8 9-8 9s-8-4-8-9V6Z"/><path d="m8 12 3 3 5-6"/>',
  eye: '<path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7S2 12 2 12Z"/><circle cx="12" cy="12" r="3"/>',
  bubble:
    '<path d="M21 11a9 9 0 0 1-9 9H4l-2 2 1-7a9 9 0 1 1 18-4Z"/><path d="M7 10h10M7 14h6"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2M5 5l1 1M18 18l1 1M5 19l1-1M18 6l1-1"/>',
  door: '<path d="M4 21V3h12v18M2 21h20M8 12h1M13 12h9m-3-3 3 3-3 3"/>',
  edit: '<path d="m15 4 5 5M4 20l5-1L21 7l-5-5L4 14Z"/>',
  school: '<path d="m2 9 10-5 10 5-10 5ZM6 11v6q6 5 12 0v-6M22 9v7"/>',
};
const icon = (name) =>
  `<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${paths[name] || paths.records}</svg>`;
const pages = {
  home: "总览",
  login: "登录",
  register: "注册",
  records: "记录",
  profile: "我的",
  info: "个人信息",
  apply: "申请请假",
  edit: "编辑资料",
  complete: "完善资料",
  password: "修改密码",
  feedback: "用户反馈",
  appearance: "显示与触感",
  about: "关于应用",
  launch: "启动检查",
  update: "版本更新",
};
const params = new URLSearchParams(location.search);
let current = Object.hasOwn(pages, params.get("page"))
  ? params.get("page")
  : "home";
let person = {
  name: "林予安",
  student: "202401020318",
  class: "软件工程 2403 班",
  phone: "138 0000 2608",
  gender: "女",
  parent: "林女士",
  parentPhone: "139 0000 2618",
  apartment: "知行公寓",
  room: "A 栋 506",
  teacher: "陈老师",
};
let requests = [
  {
    id: 1,
    type: "事假",
    status: "已同意",
    date: "10月20日",
    start: "2026-10-20T09:00",
    end: "2026-10-20T18:00",
    reason: "处理个人事务",
    outside: "离校",
    duration: "9 小时",
    course: "大学英语",
    destination: "北京市,北京市,东城区",
    travel: "公交",
  },
  {
    id: 2,
    type: "病假",
    status: "已同意",
    date: "10月12日",
    start: "2026-10-12T08:00",
    end: "2026-10-12T18:00",
    reason: "身体不适，前往医院就诊",
    outside: "离校",
    duration: "10 小时",
    course: "高等数学",
    destination: "北京市,北京市,海淀区",
    travel: "地铁",
  },
  {
    id: 3,
    type: "事假",
    status: "已同意",
    date: "9月28日",
    start: "2026-09-28T09:00",
    end: "2026-09-28T17:00",
    reason: "家庭事务",
    outside: "不离校",
    duration: "8 小时",
  },
];
const app = document.querySelector("#app");
const sheet = document.querySelector("#sheet");
let timer;
const escapeHTML = (s) =>
  String(s).replace(
    /[&<>"']/g,
    (c) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        c
      ],
  );
const avatar = (large = false, source = person.avatarURL) =>
  `<span class="avatar ${large ? "large" : ""}" aria-label="${escapeHTML(person.name || "同学")}的头像">${source && !person.avatarFailed ? `<img src="${escapeHTML(source)}" alt="" data-avatar>` : escapeHTML((person.name || "同").slice(0, 1))}</span>`;
const mark = () => `<div class="app-mark">${icon("door")}</div>`;
function navHeader(title, action = "", back = "") {
  return `<header class="navigation">${back ? `<button class="round-control" data-go="${back}" aria-label="返回">${icon("back")}</button>` : "<span></span>"}<strong>${title}</strong>${action || "<span></span>"}</header>`;
}
function tabs(active) {
  return `<footer class="bottom-controls"><nav class="tab-bar" aria-label="主导航">${[
    ["home", "总览"],
    ["records", "记录"],
    ["profile", "我的"],
  ]
    .map(
      ([key, label]) =>
        `<button data-go="${key}" ${active === key ? 'aria-current="page"' : ""}>${icon(key === "profile" ? "user" : key)}<span>${label}</span></button>`,
    )
    .join(
      "",
    )}</nav><button class="compose" data-action="apply" aria-label="申请请假">${icon("plus")}</button></footer>`;
}
function section(title, action = "") {
  return `<div class="section-heading"><h2>${title}</h2>${action}</div>`;
}
function badge() {
  return `<span class="badge approved">${icon("check")}已同意</span>`;
}
function requestRow(r) {
  return `<button class="record-row" data-request="${r.id}"><span class="record-icon ${r.type.startsWith("病假") ? "orange" : ""}">${icon(r.type.startsWith("病假") ? "plus" : "calendar")}</span><span class="record-text"><strong>${escapeHTML(r.type)} <small>${escapeHTML(r.duration)}</small></strong><span>${escapeHTML(r.date)} · ${escapeHTML(r.outside)}</span></span><span class="record-status">${badge()}${icon("arrow")}</span></button>`;
}
function home() {
  const latest = requests[0];
  return `<section class="tab-page home-page"><div class="scroll-content"><header class="home-header"><div class="home-brand"><span class="home-symbol">${icon("door")}</span><h1>请假</h1></div><button class="account-control" data-go="profile" aria-label="个人中心">${icon("user")}</button></header>
    <button class="new-request" data-action="apply"><span class="new-request-copy"><strong>申请请假</strong><small>填写信息，生成假条</small></span><span class="application-visual" aria-hidden="true"><span class="visual-back"></span><span class="visual-page">${icon("calendar")}<span class="visual-rule"></span><span class="visual-rule short"></span></span></span></button>
    ${section("最新假条", `<button class="text-button" data-go="records">查看记录${icon("arrow")}</button>`)}
    ${latest ? `<button class="active-request" data-request="${latest.id}"><div class="request-top"><span class="type-label">${icon("calendar")}${escapeHTML(latest.type)}</span>${badge()}</div><div class="request-date"><span>${escapeHTML(latest.date)}</span><strong>${escapeHTML(latest.start.slice(11, 16))}<i>—</i>${escapeHTML(latest.end.slice(11, 16))}</strong></div><div class="request-meta"><span>${icon("clock")}${escapeHTML(latest.duration)}</span><span>${icon("location")}${escapeHTML(latest.outside)}</span></div><div class="request-link"><span>查看假条</span>${icon("arrow")}</div></button>` : `<div class="empty-card">${icon("calendar")}<h3>还没有假条</h3><p>需要请假时，轻点下方的加号。</p></div>`}
    <div class="home-shortcuts"><button data-go="info">${icon("user")}<span>个人资料<small>核对身份与联系信息</small></span>${icon("arrow")}</button><button data-action="application-help">${icon("records")}<span>填写提示<small>时间、课程与去向</small></span>${icon("arrow")}</button></div>
    </div>${tabs("home")}</section>`;
}
function auth(register) {
  return `<section class="auth-page"><div class="auth-scroll"><div class="auth-heading">${mark()}<h1>校园请假</h1></div>
    <div class="auth-modes" role="group" aria-label="登录或注册">
      <button type="button" data-auth-mode="login" aria-pressed="${!register}">登录</button>
      <button type="button" data-auth-mode="register" aria-pressed="${register}">注册</button>
    </div>
    <form id="auth-form" novalidate>
      <div class="group auth-fields">
        <label class="inline-field"><span>学号</span><input name="student" id="student" value="${escapeHTML(person.student)}" inputmode="numeric" maxlength="12" autocomplete="username" placeholder="12 位学号" aria-describedby="form-error"></label>
        <label class="inline-field"><span>密码</span><input name="password" id="password" type="password" maxlength="30" autocomplete="${register ? "new-password" : "current-password"}" placeholder="${register ? "6–30 位密码" : "输入密码"}" aria-describedby="form-error"><button class="eye-button" type="button" data-eye="password" aria-pressed="false" aria-label="显示密码">${icon("eye")}</button></label>
        <label class="inline-field auth-confirm" ${register ? "" : "hidden"}><span>确认</span><input id="confirm" name="confirm" type="password" maxlength="30" autocomplete="new-password" placeholder="再次输入密码" aria-label="确认密码" aria-describedby="form-error" ${register ? "" : "disabled"}><button class="eye-button" type="button" data-eye="confirm" aria-pressed="false" aria-label="显示确认密码">${icon("eye")}</button></label>
      </div>
      <p class="form-error" id="form-error" role="alert"></p>
      <button class="primary" type="submit">${register ? "注册" : "登录"}</button>
    </form></div></section>`;
}
function switchAuthMode(mode) {
  const form = app.querySelector("#auth-form");
  if (!form || !["login", "register"].includes(mode)) return;
  const register = mode === "register";
  current = mode;
  app.querySelectorAll("[data-auth-mode]").forEach((button) => {
    button.setAttribute(
      "aria-pressed",
      String(button.dataset.authMode === mode),
    );
  });
  form.querySelector(".auth-confirm").hidden = !register;
  form.elements.confirm.disabled = !register;
  if (!register) form.elements.confirm.value = "";
  form.elements.password.placeholder = register ? "6–30 位密码" : "输入密码";
  form.elements.password.autocomplete = register
    ? "new-password"
    : "current-password";
  form.querySelector('[type="submit"]').textContent = register
    ? "注册"
    : "登录";
  form.querySelector("#form-error").textContent = "";
  form
    .querySelectorAll("input")
    .forEach((input) => input.removeAttribute("aria-invalid"));
  document.querySelectorAll(".screen-picker [data-go]").forEach((button) => {
    if (button.dataset.go === "login")
      button.setAttribute("aria-current", "page");
    else button.removeAttribute("aria-current");
    button.querySelector("small").textContent =
      button.dataset.go === "login" ? "正在预览" : "";
  });
  document.title = `${pages[mode]} · 校园请假`;
}
function settingsRow(symbol, label, action, colour = "blue", meta = "") {
  return `<button class="settings-row" ${action.startsWith("go:") ? `data-go="${action.slice(3)}"` : `data-action="${action}"`}><span class="settings-icon ${colour}">${icon(symbol)}</span><span>${label}</span>${meta ? `<small>${meta}</small>` : ""}${icon("arrow")}</button>`;
}
function profile() {
  return `<section class="tab-page"><div class="scroll-content"><div class="title-row"><h1>我的</h1><button class="round-control" data-action="about" aria-label="关于应用">${icon("sun")}</button></div><button class="student-card" data-go="info"><div class="student-card-top"><span>${icon("school")}校园身份</span>${icon("arrow")}</div><div class="student-card-main">${avatar()}<div><h2>${escapeHTML(person.name)}</h2><p>${escapeHTML(person.class)}</p></div></div><div class="student-card-foot"><span>学号</span><strong>${escapeHTML(person.student)}</strong></div><div class="card-orb" aria-hidden="true"></div></button>
    ${section("资料与账号")}<div class="group">${settingsRow("user", "个人信息", "go:info")}${settingsRow("edit", "编辑资料", "edit", "grey")}${settingsRow("lock", "修改密码", "password", "grey")}</div>
    ${section("偏好与支持")}<div class="group">${settingsRow("sun", "显示与触感", "appearance", "purple")}${settingsRow("bubble", "反馈与建议", "feedback", "orange")}${settingsRow("school", "关于应用", "about", "grey")}</div>
    <div class="group logout-group"><button class="logout" data-action="logout">退出登录</button></div><p class="content-footer">校园请假 · 设计预览</p></div>${tabs("profile")}</section>`;
}
function info() {
  const groups = [
    [
      "学生资料",
      [
        ["学号", "student"],
        ["班级", "class"],
        ["性别", "gender"],
        ["导员", "teacher"],
      ],
    ],
    [
      "联系方式",
      [
        ["手机号", "phone"],
        ["家长姓名", "parent"],
        ["家长电话", "parentPhone"],
      ],
    ],
    [
      "住宿信息",
      [
        ["公寓", "apartment"],
        ["宿舍号", "room"],
      ],
    ],
  ];
  return `<section class="info-page">${navHeader("个人信息", '<button class="text-button" data-action="edit">编辑</button>', "profile")}<div class="info-scroll"><div class="contact-header"><button class="avatar-button" data-action="avatar" aria-label="查看头像">${avatar(true)}</button><h1>${escapeHTML(person.name)}</h1><p>${escapeHTML(person.class)}</p></div><div class="contact-actions"><button data-action="edit">${icon("edit")}<span>编辑资料</span></button><button data-action="password">${icon("lock")}<span>账号安全</span></button></div>${groups.map(([title, fields]) => `${section(title)}<dl class="group information">${fields.map(([label, key]) => `<div><dt>${label}</dt><dd>${escapeHTML(person[key])}</dd></div>`).join("")}</dl>`).join("")}<p class="content-footer">${icon("shield")}资料仅用于请假申请与身份核验</p></div></section>`;
}
function records() {
  return `<section class="tab-page"><div class="scroll-content"><div class="title-row"><h1>记录</h1><button class="round-control" data-action="apply" aria-label="新建申请">${icon("plus")}</button></div><p class="list-caption">共 ${requests.length} 张假条</p><div class="group">${requests.map(requestRow).join("") || '<p class="empty-text">暂无假条</p>'}</div></div>${tabs("records")}</section>`;
}
const renderPage = (key) =>
  !["home", "login", "register", "profile", "info", "records"].includes(key)
    ? additionalPage(key)
    : key === "home"
      ? home()
      : key === "login" || key === "register"
        ? auth(key === "register")
        : key === "profile"
          ? profile()
          : key === "info"
            ? info()
            : records();
function render(focus = false) {
  app.innerHTML = renderPage(current);
  document.querySelector(".screen-picker").innerHTML = Object.entries(pages)
    .filter(([key]) => key !== "register")
    .map(
      ([key, label], i) =>
        `<button data-go="${key}" ${key === current || (key === "login" && current === "register") ? 'aria-current="page"' : ""}><span>0${i + 1}</span>${key === "login" ? "登录 / 注册" : label}<small>${key === current || (key === "login" && current === "register") ? "正在预览" : ""}</small></button>`,
    )
    .join("");
  document.title = `${pages[current]} · 校园请假`;
  if (focus) {
    const heading = app.querySelector("h1");
    if (heading) {
      heading.tabIndex = -1;
      heading.focus({ preventScroll: true });
    }
  }
  onPageRendered();
}
function toast(message) {
  const t = document.querySelector("#toast");
  t.textContent = message;
  t.classList.add("visible");
  clearTimeout(timer);
  timer = setTimeout(() => t.classList.remove("visible"), 2500);
}
function closeSheet() {
  sheet.close();
  sheet.innerHTML = "";
}
function positionSheet() {
  const phone = document.querySelector(".phone");
  const rect = phone.getBoundingClientRect();
  const border = parseFloat(getComputedStyle(phone).borderLeftWidth);
  sheet.style.setProperty("--sheet-left", `${rect.left + border}px`);
  sheet.style.setProperty(
    "--sheet-bottom",
    `${Math.max(0, innerHeight - rect.bottom + border)}px`,
  );
  sheet.style.setProperty("--sheet-width", `${rect.width - border * 2}px`);
  sheet.style.setProperty(
    "--sheet-height",
    `${Math.min(rect.height - 65, innerHeight - 40)}px`,
  );
}
function openSheet(title, content, wide = false) {
  sheet.classList.toggle("full-sheet", wide);
  sheet.innerHTML = `<div class="sheet-handle" aria-hidden="true"></div><header class="sheet-nav"><h2 id="sheet-title">${title}</h2><button class="round-control" data-action="close" aria-label="关闭">${icon("close")}</button></header><div class="sheet-content">${content}</div>`;
  positionSheet();
  sheet.showModal();
}
window.addEventListener("resize", () => {
  if (sheet.open) positionSheet();
});
function applicationForm() {
  return `<form id="apply-form" novalidate>
      <div class="group editor-fields">
        <label><span>请假类型 <i class="required">*</i></span><select name="type" required><option value="">请选择</option><option>事假-本科生</option><option>病假-本科生</option></select></label>
      </div>
      <div class="group editor-fields application-group">
        <label><span>开始时间 <i class="required">*</i></span><input name="start" type="datetime-local" step="1" required></label>
        <label><span>结束时间 <i class="required">*</i></span><input name="end" type="datetime-local" step="1" required></label>
        <div class="computed-field"><span>请假时长</span><output name="duration" aria-live="polite">—</output></div>
      </div>
      <div class="group editor-fields application-group"><label><span>影响课程</span><input name="course" maxlength="255" placeholder="请输入影响课程"></label></div>
      <div class="group editor-fields application-group"><label><span>是否离校 <i class="required">*</i></span><select name="outside"><option value="是">是</option><option value="否">否</option></select></label></div>
      <fieldset class="group destination-fields application-group" id="destination-fields"><legend>请假去向 <i class="required">*</i></legend>
        <label><span>省份</span><select name="province" required><option value="">请选择省份</option>${destinationRegions.map((r, i) => `<option value="${i}">${escapeHTML(r.title)}</option>`).join("")}</select></label>
        <label><span>城市</span><select name="city" required disabled><option value="">请选择城市</option></select></label>
        <label><span>区县</span><select name="district" required disabled><option value="">请选择区县</option></select></label>
      </fieldset>
      <label class="reason-label" for="reason">请假理由 <i class="required">*</i></label><div class="reason-editor"><textarea id="reason" name="reason" placeholder="请详细说明请假原因" maxlength="200" required></textarea><span class="reason-count" id="reason-count">0/200</span></div>
      <div class="group editor-fields application-group"><label><span>出行方式</span><input name="travel" maxlength="255" placeholder="请输入出行方式"></label></div>
      <p class="form-error" id="apply-error" role="alert"></p><button class="primary" type="submit">提交申请</button><p class="sheet-footnote">仅演示流程，不会提交到学校。</p>
    </form>`;
}
function openApplication() {
  requestNavigate("apply");
}

function durationText(startValue, endValue) {
  const difference =
    new Date(endValue).getTime() - new Date(startValue).getTime();
  if (!Number.isFinite(difference) || difference <= 0) return "—";
  const hours = Math.ceil(difference / 3600000);
  return `${Math.floor(hours / 24)}天${hours % 24}小时`;
}

function updateApplication(event) {
  const form = event.target.closest("#apply-form");
  if (!form) return;
  const fields = form.elements;
  if (["start", "end"].includes(event.target.name)) {
    fields.duration.value = durationText(fields.start.value, fields.end.value);
  }
  if (event.target.name === "reason") {
    document.querySelector("#reason-count").textContent =
      `${fields.reason.value.length}/200`;
  }
  if (event.target.name === "province") {
    const province = destinationRegions[Number(fields.province.value)];
    fields.city.innerHTML =
      '<option value="">请选择城市</option>' +
      (fields.province.value === ""
        ? ""
        : province.children
            .map(
              (r, i) => `<option value="${i}">${escapeHTML(r.title)}</option>`,
            )
            .join(""));
    fields.city.disabled = fields.province.value === "";
    fields.district.innerHTML = '<option value="">请选择区县</option>';
    fields.district.disabled = true;
  }
  if (event.target.name === "city") {
    const city =
      destinationRegions[Number(fields.province.value)]?.children[
        Number(fields.city.value)
      ];
    fields.district.innerHTML =
      '<option value="">请选择区县</option>' +
      (fields.city.value === ""
        ? ""
        : (city?.children || [])
            .map(
              (r, i) => `<option value="${i}">${escapeHTML(r.title)}</option>`,
            )
            .join(""));
    fields.district.disabled = fields.city.value === "";
  }
  if (event.target.name === "outside") {
    const outside = fields.outside.value === "是";
    document.querySelector("#destination-fields").hidden = !outside;
    fields.province.disabled = !outside;
    fields.city.disabled = !outside || fields.province.value === "";
    fields.district.disabled = !outside || fields.city.value === "";
  }
}
document.addEventListener("input", updateApplication);
document.addEventListener("change", updateApplication);
function editSheet() {
  openEditor(false);
}
document.addEventListener("click", (e) => {
  const b = e.target.closest("button");
  if (!b) return;
  if (b.dataset.authMode) switchAuthMode(b.dataset.authMode);
  if (b.dataset.go) {
    requestNavigate(b.dataset.go);
  }
  if (b.dataset.eye) {
    const input = document.getElementById(b.dataset.eye);
    const show = input.type === "password";
    input.type = show ? "text" : "password";
    b.setAttribute("aria-pressed", String(show));
    b.setAttribute(
      "aria-label",
      `${show ? "隐藏" : "显示"}${b.dataset.eye === "confirm" ? "确认" : ""}密码`,
    );
  }
  if (b.dataset.request) {
    toast("假条详情后续按当前项目的 1:1 样式复刻");
  }
  const action = b.dataset.action;
  if (action === "close") requestDismissSheet();
  if (action === "apply") openApplication();
  if (action === "application-help") {
    openSheet(
      "填写提示",
      `<ul class="guidance-list">
      <li>${icon("clock")}<div><strong>确认请假时间</strong><p>准确选择开始和结束时间，请假时长会自动计算。</p></div></li>
      <li>${icon("school")}<div><strong>核对影响课程</strong><p>如有受影响的课程，可在申请时一并填写。</p></div></li>
      <li>${icon("location")}<div><strong>填写离校去向</strong><p>选择离校时，需填写省、市和区县；不离校则无需填写。</p></div></li>
    </ul><button class="primary" data-action="start-from-help">开始申请</button>`,
    );
  }
  if (action === "start-from-help") {
    openApplication();
  }
  if (action === "edit") editSheet();
  if (action === "logout")
    openSheet(
      "退出登录",
      `<p class="sheet-description">要退出当前演示账号吗？</p><button class="primary destructive" data-action="confirm-logout">退出登录</button><button class="secondary" data-action="close">取消</button>`,
    );
  if (action === "confirm-logout") {
    closeSheet();
    requestNavigate("login");
  }
  if (action === "avatar") showAvatar();
  if (["password", "appearance", "feedback", "about"].includes(action))
    requestNavigate(action);
});
sheet.addEventListener("click", (e) => {
  if (e.target === sheet) {
    const r = sheet.getBoundingClientRect();
    if (
      e.clientX < r.left ||
      e.clientX > r.right ||
      e.clientY < r.top ||
      e.clientY > r.bottom
    )
      requestDismissSheet();
  }
});
sheet.addEventListener("close", () => {
  if (!sheet.open) sheet.innerHTML = "";
});
document.addEventListener("submit", (e) => {
  const form = e.target;
  if (!["auth-form", "apply-form", "edit-form"].includes(form.id)) return;
  e.preventDefault();
  const values = Object.fromEntries(new FormData(form));
  if (form.id === "auth-form") {
    let field = "",
      message = "";
    if (!/^\d{12}$/.test(values.student.trim())) {
      field = "student";
      message = "请输入 12 位数字学号";
    } else if (!values.password.trim()) {
      field = "password";
      message = "请输入密码";
    } else if (
      current === "register" &&
      (values.password.length < 6 || values.password.length > 30)
    ) {
      field = "password";
      message = "密码长度应为 6 至 30 位";
    } else if (current === "register" && values.password !== values.confirm) {
      field = "confirm";
      message = "两次输入的密码不一致";
    }
    form
      .querySelectorAll("input")
      .forEach((x) => x.removeAttribute("aria-invalid"));
    if (field) {
      document.querySelector("#form-error").textContent = message;
      const input = form.elements[field];
      input.setAttribute("aria-invalid", "true");
      input.focus();
      return;
    }
    if (consumeSubmissionFailure()) {
      document.querySelector("#form-error").textContent =
        "登录或注册失败，请重试";
      if (current === "login") form.elements.password.value = "";
      return;
    }
    const registering = current === "register";
    person.student = values.student.trim();
    form.reset();
    if (registering) {
      person = {
        ...person,
        name: "",
        class: "",
        phone: "",
        gender: "",
        parent: "",
        parentPhone: "",
        apartment: "",
        room: "",
        teacher: "",
        avatarURL: "",
      };
      profileCompleted = false;
    }
    current = profileCompleted ? "home" : "complete";
    render(true);
    toast(registering ? "账号已创建，请完善资料" : "已进入演示账号");
  }
  if (form.id === "apply-form") {
    const applicationError = document.querySelector("#apply-error");
    if (!values.type) {
      applicationError.textContent = "请选择请假类型";
      form.elements.type.focus();
      return;
    }
    if (!values.start || !values.end) {
      applicationError.textContent = "请选择请假时间";
      return;
    }
    const start = new Date(values.start),
      end = new Date(values.end);
    if (!Number.isFinite(+start) || !Number.isFinite(+end) || end <= start) {
      document.querySelector("#apply-error").textContent =
        "结束时间应晚于开始时间";
      return;
    }
    if (!values.reason.trim()) {
      applicationError.textContent = "请填写请假理由";
      return;
    }
    let destination = "";
    if (values.outside === "是") {
      if (
        [values.province, values.city, values.district].some(
          (v) => v === undefined || v === "",
        )
      ) {
        applicationError.textContent = "请选择请假去向";
        return;
      }
      const province = destinationRegions[Number(values.province)];
      const city = province.children[Number(values.city)];
      const district = city.children[Number(values.district)];
      destination = `${province.title},${city.title},${district.title}`;
    }
    const startDay = `${start.getMonth() + 1}月${start.getDate()}日`;
    const endDay = `${end.getMonth() + 1}月${end.getDate()}日`;
    if (consumeSubmissionFailure()) {
      applicationError.textContent = "提交失败，已保留填写内容，请重试";
      return;
    }
    requests.unshift({
      id: Date.now(),
      type: values.type,
      status: "已同意",
      date: startDay === endDay ? startDay : `${startDay} — ${endDay}`,
      start: values.start,
      end: values.end,
      reason: values.reason.trim(),
      outside: values.outside === "是" ? "离校" : "不离校",
      duration: durationText(values.start, values.end),
      course: values.course,
      travel: values.travel,
      destination,
    });
    backPages = [];
    current = "records";
    render(true);
    toast("假条已创建，申请已同意");
  }
});

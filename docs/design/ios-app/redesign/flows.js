// All preview state is in memory. No account requests or persistent storage.
let profileCompleted = true;
let backPages = [];
let pageGeneration = 0;
let renderedPage = "";
let sheetSnapshot = "";
let stagedAvatar = "";
let editorFirst = false;
let editorSource = "profile";
let editorSnapshot = "";
let applicationSource = "home";
let applicationSnapshot = "";
let editorAvatarStart = "";
let imageURLs = new Set();
let pendingConfirmation = null;
let feedbackImages = [];
let feedbackDraft = { content: "", anonymous: false };
let feedbackResult = false;
const preferences = {
  motion: false,
  transparency: false,
  contrast: false,
  haptics: false,
};
const scenarios = {
  startup: "home",
  submit: "success",
  version: "success",
  download: "success",
  install: "success",
  records: "sample",
};
let launchState = "checking";
let versionState = "loading";
let downloadProgress = 0;
const classOptions = [
  "软件工程 2403 班",
  "计算机科学 2401 班",
  "数字媒体 2402 班",
];
const apartmentOptions = {
  女: ["知行公寓", "明德公寓"],
  男: ["致远公寓", "博学公寓"],
};
const confirmation = document.querySelector("#confirmation");

function utility(title, content, back = "profile") {
  return `<section class="utility-page">${navHeader(title, "", back)}<div class="utility-scroll">${content}</div></section>`;
}
function additionalPage(key) {
  if (key === "apply") return applicationPage();
  if (key === "edit" || key === "complete")
    return profileEditorPage(key === "complete");
  if (key === "password") return passwordPage();
  if (key === "feedback") return feedbackPage();
  if (key === "appearance") return appearancePage();
  if (key === "about") return aboutPage();
  if (key === "launch") return launchPage();
  if (key === "update") return updatePage();
  return home();
}
function snapshot(form) {
  return form
    ? JSON.stringify(
        [...new FormData(form)].map(([k, v]) => [
          k,
          typeof v === "string" ? v : v.name,
        ]),
      )
    : "";
}
function trackSheetForm() {
  sheetSnapshot = snapshot(sheet.querySelector("form"));
}
function sheetDirty() {
  const form = sheet.querySelector("form");
  return !!form && snapshot(form) !== sheetSnapshot;
}
function confirmAction(title, text, action, button = "确定") {
  pendingConfirmation = action;
  positionSheet();
  confirmation.style.cssText = sheet.style.cssText;
  confirmation.innerHTML = `<div class="sheet-content confirm-content"><h2 id="confirmation-title">${title}</h2><p>${text}</p><button class="primary destructive" data-flow="confirm">${button}</button><button class="secondary" data-flow="cancel-confirm">继续编辑</button></div>`;
  confirmation.showModal();
}
function requestDismissSheet(after) {
  if (sheet.querySelector("form[data-busy]")) {
    toast("正在保存，请稍候");
    return;
  }
  const finish = () => {
    closeSheet();
    if (after) after();
  };
  if (sheetDirty())
    confirmAction(
      "放弃未保存的内容？",
      "离开后，本次尚未保存的填写内容将被清空。",
      finish,
      "放弃修改",
    );
  else finish();
}
function requestNavigate(target) {
  if (!Object.hasOwn(pages, target)) return;
  if (app.querySelector("#apply-form")) {
    if (target !== current) leaveApplication(() => navigateTo(target));
    return;
  }
  if (target === current && app.querySelector("#profile-form")) return;
  if (app.querySelector("#profile-form")) {
    leaveEditor(() => navigateTo(target));
    return;
  }
  if (target === "edit" || target === "complete") {
    openEditor(target === "complete");
    return;
  }
  navigateTo(target);
}
function navigateTo(target) {
  if (
    ["login", "register"].includes(current) &&
    ["login", "register"].includes(target) &&
    app.querySelector("#auth-form")
  ) {
    switchAuthMode(target);
    return;
  }
  const go = () => {
    if (renderedPage === "feedback" && current !== "feedback")
      feedbackDraft = { content: "", anonymous: false };
    if (["home", "records", "profile"].includes(target)) backPages = [];
    else if (target !== current) backPages.push(current);
    if (target === "apply") applicationSource = current;
    current = target;
    pageGeneration++;
    render(true);
    if (
      preferences.haptics &&
      ["home", "records", "profile"].includes(target) &&
      navigator.vibrate
    ) {
      try {
        navigator.vibrate(8);
      } catch {}
    }
  };
  if (sheet.open) requestDismissSheet(go);
  else go();
}
function onPageRendered() {
  document.querySelectorAll(".navigation [data-go]").forEach((b) => {
    b.removeAttribute("data-go");
    b.dataset.flow = "back";
  });
  if (current !== renderedPage) {
    pageGeneration++;
    if (current === "launch") launchState = "checking";
    if (current === "update") {
      versionState = "loading";
      downloadProgress = 0;
    }
    renderedPage = current;
  }
  const generation = pageGeneration;
  if (current === "apply")
    applicationSnapshot = snapshot(app.querySelector("#apply-form"));
  const editor = app.querySelector("#profile-form");
  if (editor) {
    editorSnapshot = snapshot(editor);
    editorAvatarStart = stagedAvatar;
  }
  if (current === "launch" && launchState === "checking") {
    setTimeout(() => {
      if (current !== "launch" || generation !== pageGeneration) return;
      if (scenarios.startup === "failure") {
        launchState = "failed";
        render();
        return;
      }
      const destination = scenarios.startup;
      if (destination === "complete") profileCompleted = false;
      current = destination;
      render(true);
    }, 650);
  }
  if (current === "update" && versionState === "loading") {
    setTimeout(() => {
      if (current !== "update" || generation !== pageGeneration) return;
      versionState = scenarios.version === "failure" ? "load-error" : "ready";
      render();
    }, 500);
  }
  syncPreferences();
}
function goBack() {
  if (app.querySelector("#apply-form")) {
    leaveApplication();
    return;
  }
  if (app.querySelector("#profile-form")) {
    leaveEditor();
    return;
  }
  const fallback = ["login", "register", "launch", "update"].includes(current)
    ? "login"
    : "profile";
  const target = backPages.pop() || fallback;
  current = target;
  pageGeneration++;
  render(true);
}
sheet.addEventListener("cancel", (e) => {
  e.preventDefault();
  requestDismissSheet();
});
confirmation.addEventListener("close", () => {
  confirmation.innerHTML = "";
  pendingConfirmation = null;
});

function avatarPickerMarkup() {
  return `<div class="editor-avatar"><span id="staged-avatar">${avatar(true, stagedAvatar)}</span><button class="text-button" type="button" data-flow="choose-avatar">选择头像</button><input class="file-input" id="avatar-file" type="file" accept="image/*" aria-label="选择头像图片"></div>`;
}
function editorForm(first = false) {
  const value = (key) => escapeHTML(person[key] || "");
  const options = (list, selected) =>
    '<option value="">请选择</option>' +
    list
      .map(
        (x) =>
          `<option ${x === selected ? "selected" : ""}>${escapeHTML(x)}</option>`,
      )
      .join("");
  const textField = (label, key, max, required = true, phone = false) =>
    `<label><span>${label}${required ? ' <i class="required">*</i>' : ""}</span><input name="${key}" value="${phone ? value(key).replace(/\s/g, "") : value(key)}" maxlength="${max}" ${phone ? 'inputmode="numeric"' : ""} ${required ? "required" : ""}></label>`;
  const gender = person.gender;
  return `<form id="profile-form" novalidate>${avatarPickerMarkup()}
    <p class="field-help">学号 ${escapeHTML(person.student)} 不可修改。</p>
    <div class="group editor-fields application-group">
      <label><span>班级 <i class="required">*</i></span><select name="class" required>${options(classOptions, person.class)}</select></label>
      ${textField("姓名", "name", 20)}${textField("手机号", "phone", 11, true, true)}
      <label><span>性别 <i class="required">*</i></span><select name="gender" required>${options(["男", "女"], gender)}</select></label>
      ${textField("家长姓名", "parent", 20)}${textField("家长电话", "parentPhone", 11, true, true)}
      <label><span>公寓名称</span><select name="apartment" ${gender ? "" : "disabled"}>${options(apartmentOptions[gender] || [], person.apartment)}</select></label>
      ${textField("宿舍号", "room", 20, false)}${textField("导员姓名", "teacher", 20)}
    </div><p class="form-error" id="profile-error" role="alert"></p><button class="primary" type="submit">${first ? "完成并进入首页" : "保存修改"}</button></form>`;
}
function applicationPage() {
  return `<section class="utility-page application-page">${navHeader("申请请假", `<button class="round-control" data-flow="close-application" aria-label="关闭请假申请">${icon("close")}</button>`, "home")}<div class="utility-scroll">${applicationForm()}</div></section>`;
}
function leaveApplication(after) {
  const form = app.querySelector("#apply-form");
  const finish = () => {
    if (after) after();
    else {
      current = applicationSource;
      if (backPages.at(-1) === applicationSource) backPages.pop();
      render(true);
    }
  };
  if (form && snapshot(form) !== applicationSnapshot) {
    confirmAction(
      "放弃未提交的申请？",
      "本次填写的内容将不会保存。",
      finish,
      "放弃申请",
    );
  } else finish();
}
function profileEditorPage(first = false) {
  editorFirst = first;
  stagedAvatar = person.avatarURL || "";
  return `<section class="utility-page profile-editor-page">${navHeader(first ? "完善资料" : "编辑资料", `<button class="round-control" data-flow="close-editor" aria-label="关闭资料编辑">${icon("close")}</button>`, "profile")}<div class="utility-scroll"><p class="profile-editor-description">${first ? "完善必填信息，即可开始请假。" : "更新个人信息，保存后返回我的。"}</p>${editorForm(first)}</div></section>`;
}
function openEditor(first = false) {
  editorSource = first
    ? "register"
    : current === "edit" || current === "complete"
      ? "profile"
      : current;
  current = first ? "complete" : "edit";
  render(true);
}
function leaveEditor(after) {
  const form = app.querySelector("#profile-form");
  if (form?.dataset.busy) {
    toast("正在保存，请稍候");
    return;
  }
  const finish = () => {
    if (stagedAvatar !== person.avatarURL) releaseImage(stagedAvatar);
    if (after) after();
    else {
      current = current === "complete" ? "register" : editorSource;
      render(true);
    }
  };
  if (
    form &&
    (snapshot(form) !== editorSnapshot || stagedAvatar !== editorAvatarStart)
  ) {
    confirmAction(
      "放弃未保存的内容？",
      "本次修改将不会保存。",
      finish,
      "放弃修改",
    );
  } else finish();
}
document.addEventListener("keydown", (e) => {
  if (
    e.key === "Escape" &&
    !sheet.open &&
    !confirmation.open &&
    app.querySelector("#profile-form, #apply-form")
  ) {
    e.preventDefault();
    if (current === "apply") leaveApplication();
    else leaveEditor();
  }
});
document.addEventListener("focusin", (e) => {
  if (
    e.target.matches(
      "#profile-form input:not([type=file]), #profile-form select, #apply-form input, #apply-form select, #apply-form textarea",
    )
  ) {
    requestAnimationFrame(() => e.target.scrollIntoView({ block: "nearest" }));
  }
});
function scrollEditorFocusIntoView() {
  const focused = document.activeElement;
  if (focused?.closest("#profile-form, #apply-form"))
    focused.scrollIntoView({ block: "nearest" });
}
window.visualViewport?.addEventListener("resize", scrollEditorFocusIntoView);
function makeImageURL(file) {
  const url = URL.createObjectURL(file);
  imageURLs.add(url);
  return url;
}
function releaseImage(url) {
  if (url && imageURLs.has(url)) {
    URL.revokeObjectURL(url);
    imageURLs.delete(url);
  }
}
function showAvatar() {
  openSheet(
    "头像",
    `<div class="avatar-preview">${avatar(true)}<h3>${escapeHTML(person.name)}</h3><button class="text-button" data-flow="choose-avatar">更换头像</button><input class="file-input" id="avatar-file" type="file" accept="image/*" aria-label="选择头像图片"></div><button class="secondary" data-action="close">完成</button>`,
  );
}
document.addEventListener(
  "error",
  (e) => {
    if (e.target.matches?.("img[data-avatar]")) {
      person.avatarFailed = true;
      e.target.parentElement.innerHTML = escapeHTML(
        (person.name || "同").slice(0, 1),
      );
    }
  },
  true,
);
window.addEventListener("beforeunload", () =>
  imageURLs.forEach((url) => URL.revokeObjectURL(url)),
);

function passwordPage() {
  const fields = [
    ["旧密码", "old"],
    ["新密码", "new"],
    ["确认新密码", "confirm-new"],
  ];
  return utility(
    "修改密码",
    `<div class="utility-intro"><span class="utility-mark">${icon("lock")}</span><h1>为账号设置新密码</h1><p>修改后需要重新登录。</p></div><form id="password-form" novalidate><div class="group password-fields">${fields.map(([label, key]) => `<label class="password-row"><span>${label}</span><div><input name="${key}" id="${key}" type="password" maxlength="30" placeholder="${label === "旧密码" ? "输入当前密码" : label === "新密码" ? "6–30 位新密码" : "再次输入新密码"}" autocomplete="off"><button class="eye-button" type="button" data-eye="${key}" aria-label="显示${label}" aria-pressed="false">${icon("eye")}</button></div></label>`).join("")}</div><p class="field-help">新密码不少于 6 位，最多 30 位。</p><p class="form-error" id="password-error" role="alert"></p><button class="primary" type="submit">确认修改</button></form>`,
  );
}
function feedbackPage() {
  if (feedbackResult)
    return utility(
      "用户反馈",
      `<div class="result-state"><span class="result-symbol">${icon("check")}</span><h1>感谢你的反馈</h1><p>你的想法，让校园请假更好用。</p><button class="primary" data-flow="feedback-again">再写一条</button><button class="secondary" data-go="profile">返回我的</button></div>`,
    );
  return utility(
    "用户反馈",
    `<div class="utility-intro"><h1>我们想听听你的想法</h1><p>描述遇到的问题，或分享改进建议。</p></div><form id="feedback-form" novalidate><label class="form-label" for="feedback-content">反馈内容 <i class="required">*</i></label><div class="reason-editor"><textarea id="feedback-content" name="content" maxlength="500" placeholder="请描述问题或建议">${escapeHTML(feedbackDraft.content)}</textarea><span class="reason-count" id="feedback-count">${feedbackDraft.content.length}/500</span></div><div class="section-heading"><h2>附件图片</h2><span class="section-note">选填 · 最多 3 张</span></div><div id="attachment-list" class="attachment-list">${attachmentMarkup()}</div><input id="feedback-files" type="file" accept="image/*" multiple class="file-input" aria-label="添加反馈图片"><div class="group application-group"><label class="switch-row"><span><strong>匿名提交</strong><small>不展示个人身份信息</small></span><input name="anonymous" type="checkbox" role="switch" ${feedbackDraft.anonymous ? "checked" : ""}></label></div><p class="form-error" id="feedback-error" role="alert"></p><button class="primary" type="submit">提交反馈</button></form>`,
  );
}
function attachmentMarkup() {
  return (
    feedbackImages
      .map(
        (item, i) =>
          `<div class="attachment"><button type="button" data-preview-image="${i}" aria-label="预览第 ${i + 1} 张图片"><img src="${item.url}" alt="反馈附件 ${i + 1}"></button><button type="button" class="remove-image" data-remove-image="${i}" aria-label="删除第 ${i + 1} 张图片">${icon("close")}</button></div>`,
      )
      .join("") +
    (feedbackImages.length < 3
      ? `<button type="button" class="attachment-add" data-flow="choose-feedback">${icon("plus")}添加图片</button>`
      : "")
  );
}
function appearancePage() {
  const toggles = [
    ["motion", "减少动态效果", "关闭位移与弹性过渡"],
    ["transparency", "减少透明度", "使用实色导航与面板"],
    ["contrast", "增强对比度", "加深文字和边界"],
    ["haptics", "切换轻触反馈", "切换标签页时尝试轻微振动"],
  ];
  return utility(
    "显示与触感",
    `<div class="utility-intro"><h1>选择舒服的体验</h1><p>让界面适应你的使用习惯。</p></div><div class="group">${toggles.map(([key, title, desc]) => `<label class="switch-row"><span><strong>${title}</strong><small>${desc}</small></span><input type="checkbox" role="switch" data-preference="${key}" aria-label="${title}" ${preferences[key] ? "checked" : ""}></label>`).join("")}</div><p class="field-help">系统辅助设置同样生效。触觉反馈取决于浏览器和设备支持；本次预览中的选择会保留到刷新前。</p>`,
  );
}
function syncPreferences() {
  const phone = document.querySelector(".phone");
  if (!phone) return;
  phone.classList.toggle("reduce-motion", preferences.motion);
  phone.classList.toggle(
    "reduce-transparency",
    preferences.transparency || preferences.contrast,
  );
  phone.classList.toggle("high-contrast", preferences.contrast);
}
function aboutPage() {
  return utility(
    "关于应用",
    `<div class="about-hero">${mark()}<h1>校园请假</h1><p>让每次请假，清楚而简单。</p><span>版本 1.0.0</span></div><div class="group">${settingsRow("sun", "版本更新", "go:update", "blue", "查看")}${settingsRow("bubble", "反馈与建议", "go:feedback", "orange")}</div><div class="about-note">此处是独立的交互设计预览。账号、版本和申请均为虚构示例。</div>`,
  );
}
function launchPage() {
  return `<section class="utility-page"><div class="utility-scroll"><div class="launch-state">${mark()}<h1>校园请假</h1>${launchState === "checking" ? '<span class="spinner" role="status" aria-label="正在检查版本"></span><h2>正在检查版本</h2><p>即将进入你的校园空间。</p>' : '<span class="error-symbol">!</span><h2>暂时无法检查版本</h2><p>请检查连接后重试。</p><button class="primary" data-flow="retry-launch">重新检查</button>'}</div></div></section>`;
}
function updatePage() {
  let content = "";
  if (versionState === "loading")
    content =
      '<div class="loading-state"><span class="spinner" role="status" aria-label="获取版本信息"></span><p>正在获取版本信息</p></div>';
  else if (versionState === "load-error")
    content =
      '<div class="state-panel"><h2>无法获取版本信息</h2><p>请检查连接后重试。</p><button class="primary" data-flow="retry-version">重试</button></div>';
  else {
    content = `<div class="group information"><div><dt>当前版本</dt><dd>1.0.0</dd></div><div><dt>最新版本</dt><dd>1.1.0</dd></div></div><div class="release-notes"><h2>本次更新</h2><ul><li>统一页面布局与导航体验</li><li>改善资料填写和请假操作</li><li>优化辅助显示设置</li></ul></div>`;
    if (versionState === "downloading")
      content += `<div class="download-state"><div><span>正在下载</span><strong>${downloadProgress}%</strong></div><progress max="100" value="${downloadProgress}" aria-label="下载进度"></progress><button class="text-button" data-flow="cancel-download">取消下载</button></div>`;
    else if (versionState === "installing")
      content +=
        '<div class="loading-state"><span class="spinner"></span><p>正在交接安装</p></div>';
    else if (versionState === "done")
      content +=
        '<div class="state-panel"><h2>安装交接演示完成</h2><p>没有下载文件或启动真实安装程序。</p><button class="primary" data-go="login">返回登录</button></div>';
    else
      content += `${versionState === "download-error" || versionState === "install-error" ? `<p class="form-error" role="alert">${versionState === "download-error" ? "下载失败，请重试" : "未能调起安装，请重试"}</p>` : ""}<button class="primary" data-flow="${versionState === "install-error" ? "retry-install" : "download"}">${versionState === "install-error" ? "重试安装" : versionState === "download-error" ? "重新下载" : "下载并安装"}</button>`;
  }
  return utility(
    "版本更新",
    `<div class="utility-intro update-intro"><span class="utility-mark">${icon("sun")}</span><h1>发现新版本</h1><p>更新后继续使用校园请假。</p></div>${content}`,
    "about",
  );
}
function startDownload() {
  versionState = "downloading";
  downloadProgress = 0;
  render();
  const generation = pageGeneration;
  const tick = () => {
    if (
      current !== "update" ||
      generation !== pageGeneration ||
      versionState !== "downloading"
    )
      return;
    downloadProgress += 20;
    if (downloadProgress === 60 && scenarios.download === "failure") {
      versionState = "download-error";
      render();
      return;
    }
    if (downloadProgress >= 100) {
      startInstall();
      return;
    }
    render();
    setTimeout(tick, 200);
  };
  setTimeout(tick, 200);
}
function startInstall() {
  versionState = "installing";
  render();
  const generation = pageGeneration;
  setTimeout(() => {
    if (
      current !== "update" ||
      generation !== pageGeneration ||
      versionState !== "installing"
    )
      return;
    versionState = scenarios.install === "failure" ? "install-error" : "done";
    render();
  }, 600);
}

function setFormError(id, message) {
  document.getElementById(id).textContent = message;
}
function consumeSubmissionFailure() {
  if (scenarios.submit !== "failure") return false;
  scenarios.submit = "success";
  const control = document.querySelector("[data-scenario=submit]");
  if (control) control.value = "success";
  return true;
}
function simulateForm(form, callback, errorId) {
  const button = form.querySelector("[type=submit]");
  const original = button.textContent;
  const controls = [...form.elements];
  const disabled = controls.map((control) => control.disabled);
  controls.forEach((control) => (control.disabled = true));
  form.dataset.busy = "true";
  button.textContent = "正在提交…";
  setTimeout(() => {
    if (!form.isConnected) return;
    controls.forEach((control, i) => (control.disabled = disabled[i]));
    delete form.dataset.busy;
    button.textContent = original;
    if (consumeSubmissionFailure()) {
      setFormError(errorId, "提交失败，已保留内容，请重试");
      return;
    }
    callback();
  }, 350);
}
document.addEventListener("submit", (e) => {
  const form = e.target;
  if (!["profile-form", "password-form", "feedback-form"].includes(form.id))
    return;
  e.preventDefault();
  const values = Object.fromEntries(new FormData(form));
  if (form.id === "profile-form") {
    const required = [
      "class",
      "name",
      "phone",
      "gender",
      "parent",
      "parentPhone",
      "teacher",
    ];
    form
      .querySelectorAll("[aria-invalid]")
      .forEach((input) => input.removeAttribute("aria-invalid"));
    const missing = required.find((key) => !values[key]?.trim());
    if (missing) {
      setFormError("profile-error", "请完整填写必填信息");
      form.elements[missing].setAttribute("aria-invalid", "true");
      form.elements[missing].focus();
      return;
    }
    if (
      !/^1[3-9]\d{9}$/.test(values.phone) ||
      !/^1[3-9]\d{9}$/.test(values.parentPhone)
    ) {
      setFormError("profile-error", "请输入正确的 11 位手机号");
      const invalid = !/^1[3-9]\d{9}$/.test(values.phone)
        ? "phone"
        : "parentPhone";
      form.elements[invalid].setAttribute("aria-invalid", "true");
      form.elements[invalid].focus();
      return;
    }
    simulateForm(
      form,
      () => {
        const old = person.avatarURL;
        person = {
          ...person,
          ...values,
          avatarURL: stagedAvatar,
          avatarFailed: false,
        };
        if (old !== stagedAvatar) releaseImage(old);
        profileCompleted = true;
        backPages = [];
        current = editorFirst ? "home" : "profile";
        render(true);
        toast("个人信息已保存");
      },
      "profile-error",
    );
  }
  if (form.id === "password-form") {
    if (
      !values.old.trim() ||
      !values.new.trim() ||
      !values["confirm-new"].trim()
    ) {
      setFormError("password-error", "请完整填写密码信息");
      return;
    }
    if (values.new.length < 6) {
      setFormError("password-error", "新密码不能少于 6 位");
      return;
    }
    if (values.new !== values["confirm-new"]) {
      setFormError("password-error", "两次输入的新密码不一致");
      return;
    }
    simulateForm(
      form,
      () => {
        form.reset();
        backPages = [];
        current = "login";
        render(true);
        toast("密码修改成功，请重新登录");
      },
      "password-error",
    );
  }
  if (form.id === "feedback-form") {
    if (!values.content.trim()) {
      setFormError("feedback-error", "请填写反馈内容");
      return;
    }
    simulateForm(
      form,
      () => {
        feedbackImages.forEach((x) => releaseImage(x.url));
        feedbackImages = [];
        feedbackDraft = { content: "", anonymous: false };
        feedbackResult = true;
        render(true);
      },
      "feedback-error",
    );
  }
});
document.addEventListener("input", (e) => {
  if (e.target.id === "feedback-content") {
    feedbackDraft.content = e.target.value;
    document.querySelector("#feedback-count").textContent =
      `${e.target.value.length}/500`;
  }
});
document.addEventListener("change", (e) => {
  if (e.target.dataset.preference) {
    preferences[e.target.dataset.preference] = e.target.checked;
    syncPreferences();
  }
  if (e.target.name === "anonymous") feedbackDraft.anonymous = e.target.checked;
  if (e.target.name === "gender" && e.target.closest("#profile-form")) {
    const apartment = e.target
      .closest("form")
      .querySelector("[name=apartment]");
    apartment.innerHTML =
      '<option value="">请选择</option>' +
      (apartmentOptions[e.target.value] || [])
        .map((x) => `<option>${x}</option>`)
        .join("");
    apartment.disabled = !e.target.value;
  }
  if (e.target.id === "avatar-file") {
    const file = e.target.files[0];
    if (!file) return;
    if (!file.type.startsWith("image/")) {
      toast("请选择图片文件");
      return;
    }
    const url = makeImageURL(file);
    if (e.target.closest("#profile-form")) {
      if (stagedAvatar !== person.avatarURL) releaseImage(stagedAvatar);
      stagedAvatar = url;
      person.avatarFailed = false;
      document.querySelector("#staged-avatar").innerHTML = avatar(
        true,
        stagedAvatar,
      );
    } else {
      releaseImage(person.avatarURL);
      person.avatarURL = url;
      person.avatarFailed = false;
      closeSheet();
      render();
      showAvatar();
    }
  }
  if (e.target.id === "feedback-files") {
    for (const file of e.target.files) {
      if (feedbackImages.length >= 3) {
        toast("最多添加三张图片");
        break;
      }
      if (!file.type.startsWith("image/")) continue;
      const key = `${file.name}:${file.size}:${file.lastModified}`;
      if (!feedbackImages.some((x) => x.key === key))
        feedbackImages.push({ key, url: makeImageURL(file) });
    }
    e.target.value = "";
    document.querySelector("#attachment-list").innerHTML = attachmentMarkup();
  }
});
document.addEventListener("click", (e) => {
  const b = e.target.closest("button");
  if (!b) return;
  const action = b.dataset.flow;
  if (action === "back") goBack();
  if (action === "close-editor") leaveEditor();
  if (action === "close-application") leaveApplication();
  if (action === "choose-avatar")
    document.querySelector("#avatar-file").click();
  if (action === "choose-feedback")
    document.querySelector("#feedback-files").click();
  if (action === "confirm") {
    const callback = pendingConfirmation;
    confirmation.close();
    if (callback) callback();
  }
  if (action === "cancel-confirm") confirmation.close();
  if (action === "feedback-again") {
    feedbackResult = false;
    render(true);
  }
  if (b.dataset.removeImage !== undefined) {
    releaseImage(feedbackImages[Number(b.dataset.removeImage)].url);
    feedbackImages.splice(Number(b.dataset.removeImage), 1);
    document.querySelector("#attachment-list").innerHTML = attachmentMarkup();
  }
  if (b.dataset.previewImage !== undefined)
    openSheet(
      "附件预览",
      `<img class="full-image" src="${feedbackImages[Number(b.dataset.previewImage)].url}" alt="反馈图片"><button class="secondary" data-action="close">完成</button>`,
    );
  if (action === "retry-launch") {
    launchState = "checking";
    render();
  }
  if (action === "retry-version") {
    versionState = "loading";
    render();
  }
  if (action === "download") startDownload();
  if (action === "cancel-download") {
    versionState = "ready";
    render();
  }
  if (action === "retry-install") startInstall();
});

function setupScenarios() {
  const definitions = [
    [
      "startup",
      "启动检查",
      [
        ["home", "进入首页"],
        ["login", "进入登录"],
        ["complete", "进入完善资料"],
        ["update", "发现更新"],
        ["failure", "检查失败"],
      ],
    ],
    [
      "submit",
      "表单提交",
      [
        ["success", "成功"],
        ["failure", "下一次失败"],
      ],
    ],
    [
      "version",
      "版本信息",
      [
        ["success", "加载成功"],
        ["failure", "加载失败"],
      ],
    ],
    [
      "download",
      "下载过程",
      [
        ["success", "下载成功"],
        ["failure", "下载失败"],
      ],
    ],
    [
      "install",
      "安装交接",
      [
        ["success", "交接完成"],
        ["failure", "交接失败"],
      ],
    ],
    [
      "records",
      "假条数据",
      [
        ["sample", "示例记录"],
        ["empty", "空列表"],
      ],
    ],
  ];
  document.querySelector("#scenario-options").innerHTML =
    definitions
      .map(
        ([key, title, options]) =>
          `<label>${title}<select data-scenario="${key}">${options.map(([value, label]) => `<option value="${value}">${label}</option>`).join("")}</select></label>`,
      )
      .join("") +
    '<button class="text-button" data-flow="avatar-failure">演示头像加载失败</button>';
  const sampleRequests = structuredClone(requests);
  document.addEventListener("change", (e) => {
    const key = e.target.dataset.scenario;
    if (!key) return;
    scenarios[key] = e.target.value;
    if (key === "records") {
      requests =
        e.target.value === "empty" ? [] : structuredClone(sampleRequests);
      if (["home", "records"].includes(current)) render();
    }
  });
  document.addEventListener("click", (e) => {
    if (e.target.closest("[data-flow=avatar-failure]")) {
      person.avatarFailed = true;
      render();
      toast("已使用姓名头像作为回退");
    }
  });
}
function renderOverview(group) {
  const groups = {
    main: ["home", "apply", "records", "profile", "info"],
    account: ["login", "register", "complete", "edit", "password"],
    support: ["feedback", "appearance", "about", "launch", "update"],
  };
  const order = groups[group] || groups.main;
  const previewMarkup = (key) => {
    if (key === "apply") return applicationPage();
    if (key === "edit" || key === "complete") {
      return profileEditorPage(key === "complete");
    }
    return renderPage(key);
  };
  versionState = "ready";
  document.body.classList.add("overview");
  document.body.dataset.overviewGroup = group;
  document.body.innerHTML = `<header class="overview-heading"><span>LEAVE / COMPLETE INTERFACE STUDY</span><h1>${{ main: "请假与个人中心", account: "账号与资料", support: "偏好、反馈与更新" }[group] || "请假与个人中心"}</h1><p>统一 iOS 风格 · 完整交互预览</p><small>假条详情后续 1:1 复刻 / 虚构数据</small></header><div class="overview-grid">${order.map((key) => `<section class="overview-item"><p>${pages[key]}</p><div class="phone"><div class="status"><span>9:41</span><span>••• ▰</span></div><div class="app">${previewMarkup(key)}</div><div class="home-indicator"></div></div></section>`).join("")}</div>`;
}
if (params.has("overview")) renderOverview(params.get("overview"));
else {
  setupScenarios();
  render();
}
